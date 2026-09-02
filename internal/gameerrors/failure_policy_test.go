package gameerrors

import (
	"context"
	"reflect"
	"testing"
)

type recordingReporter struct {
	key     string
	capture Capture
}

func (r *recordingReporter) Capture(_ context.Context, key string, capture Capture) error {
	r.key, r.capture = key, capture
	return nil
}

func TestClassifyGameFailures(t *testing.T) {
	tests := []struct {
		name      string
		failure   Failure
		wantLevel string
	}{
		{"live event escalates immediately", Failure{LiveEvent, "close_tournament", "run-17", 1, "round close transaction failed"}, "error"},
		{"asset retry begins as warning", Failure{PlayerAsset, "publish_map", "asset-42-run-1", 1, "bundle validation failed"}, "warning"},
		{"moderation backlog escalates", Failure{ModerationQueue, "scan_upload", "queue-9-item-7", 3, "classifier job failed"}, "error"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			capture, decision, err := Classify(test.failure)
			if err != nil {
				t.Fatal(err)
			}
			wantFingerprint := []string{"game-backend", string(test.failure.Workload), test.failure.Operation}
			if decision.Level != test.wantLevel || capture.Level != test.wantLevel {
				t.Fatalf("level = %q, want %q", decision.Level, test.wantLevel)
			}
			if !reflect.DeepEqual(capture.Fingerprint, wantFingerprint) {
				t.Fatalf("fingerprint = %v, want %v", capture.Fingerprint, wantFingerprint)
			}
			if capture.Context["occurrence_id"] != test.failure.OccurrenceID {
				t.Fatal("occurrence context was not retained")
			}
		})
	}
}
