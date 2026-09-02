package gameerrors

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type Reporter interface {
	Capture(context.Context, string, Capture) error
}

type Handler struct{ reporter Reporter }

func NewHandler(reporter Reporter) *Handler { return &Handler{reporter: reporter} }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method must be POST"})
		return
	}

	var failure Failure
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&failure); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid failure payload"})
		return
	}
	capture, decision, err := Classify(failure)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := h.reporter.Capture(r.Context(), failure.OccurrenceID, capture); err != nil {
		status := http.StatusBadGateway
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Status >= 400 && apiErr.Status < 500 {
			status = apiErr.Status
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, decision)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
