package gameerrors

import (
	"fmt"
	"strings"
)

type Workload string

const (
	PlayerAsset     Workload = "player_asset"
	LiveEvent       Workload = "live_event"
	ModerationQueue Workload = "moderation_queue"
)

type Failure struct {
	Workload     Workload `json:"workload"`
	Operation    string   `json:"operation"`
	OccurrenceID string   `json:"occurrence_id"`
	Attempt      int      `json:"attempt"`
	Exception    string   `json:"exception"`
}

type Capture struct {
	Title       string            `json:"title"`
	Message     string            `json:"message"`
	Level       string            `json:"level"`
	Fingerprint []string          `json:"fingerprint"`
	Exception   string            `json:"exception"`
	Context     map[string]string `json:"context"`
}

type Decision struct {
	Level       string   `json:"level"`
	Fingerprint []string `json:"fingerprint"`
	State       string   `json:"state"`
}

func Classify(f Failure) (Capture, Decision, error) {
	if !validWorkload(f.Workload) {
		return Capture{}, Decision{}, fmt.Errorf("unknown workload %q", f.Workload)
	}
	if strings.TrimSpace(f.Operation) == "" || strings.TrimSpace(f.OccurrenceID) == "" || strings.TrimSpace(f.Exception) == "" {
		return Capture{}, Decision{}, fmt.Errorf("operation, occurrence_id, and exception are required")
	}
	if f.Attempt < 1 {
		return Capture{}, Decision{}, fmt.Errorf("attempt must be at least 1")
	}

	level := "warning"
	if f.Workload == LiveEvent || f.Attempt >= 3 {
		level = "error"
	}
	fingerprint := []string{"game-backend", string(f.Workload), f.Operation}
	capture := Capture{
		Title:       fmt.Sprintf("%s operation failed", f.Workload),
		Message:     fmt.Sprintf("%s failed on attempt %d", f.Operation, f.Attempt),
		Level:       level,
		Fingerprint: fingerprint,
		Exception:   f.Exception,
		Context: map[string]string{
			"workload":      string(f.Workload),
			"operation":     f.Operation,
			"occurrence_id": f.OccurrenceID,
			"attempt":       fmt.Sprint(f.Attempt),
		},
	}
	return capture, Decision{Level: level, Fingerprint: fingerprint, State: "captured"}, nil
}

func validWorkload(workload Workload) bool {
	return workload == PlayerAsset || workload == LiveEvent || workload == ModerationQueue
}
