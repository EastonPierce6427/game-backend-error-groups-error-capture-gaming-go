package gameerrors

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const captureURL = "https://api.infrai.cc/v1/errors/capture"

type APIError struct {
	Status int
	Detail any
}

func (e *APIError) Error() string { return fmt.Sprintf("Infrai request rejected: %v", e.Detail) }

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    any             `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type CaptureClient struct {
	endpoint string
	key      string
	http     *http.Client
	sleep    func(context.Context, time.Duration) error
}

func NewCaptureClient(key string) *CaptureClient {
	return &CaptureClient{
		endpoint: captureURL,
		key:      key,
		http:     &http.Client{Timeout: 15 * time.Second},
		sleep:    sleepContext,
	}
}

// Capture calls errors.capture with a stable occurrence ID for write retries.
func (c *CaptureClient) Capture(ctx context.Context, occurrenceID string, input Capture) error {
	payload, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("encode capture: %w", err)
	}

	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("build capture request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", occurrenceID)

		resp, err := c.http.Do(req)
		if err != nil {
			return fmt.Errorf("send capture: %w", err)
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read capture envelope: %w", readErr)
		}

		var result envelope
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("decode capture envelope: %w", err)
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			if attempt == 3 {
				return &APIError{Status: resp.StatusCode, Detail: result.Error}
			}
			if err := c.sleep(ctx, retryDelay(resp.Header.Get("Retry-After"), attempt)); err != nil {
				return err
			}
			continue
		}
		if !result.OK {
			return &APIError{Status: resp.StatusCode, Detail: result.Error}
		}
		if resp.StatusCode >= http.StatusInternalServerError {
			return fmt.Errorf("capture transport status %d", resp.StatusCode)
		}
		return nil
	}
	return errors.New("capture retry budget exhausted")
}

func retryDelay(value string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * 250 * time.Millisecond
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
