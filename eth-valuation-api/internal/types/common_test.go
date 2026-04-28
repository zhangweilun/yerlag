package types

import (
	"encoding/json"
	"testing"
)

func TestTimeSeriesPointJSON(t *testing.T) {
	point := TimeSeriesPoint{
		Timestamp: 1700000000,
		Value:     3456.78,
	}

	data, err := json.Marshal(point)
	if err != nil {
		t.Fatalf("failed to marshal TimeSeriesPoint: %v", err)
	}

	var decoded TimeSeriesPoint
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal TimeSeriesPoint: %v", err)
	}

	if decoded.Timestamp != point.Timestamp {
		t.Errorf("Timestamp mismatch: got %d, want %d", decoded.Timestamp, point.Timestamp)
	}
	if decoded.Value != point.Value {
		t.Errorf("Value mismatch: got %f, want %f", decoded.Value, point.Value)
	}
}

func TestAPIResponseJSON(t *testing.T) {
	resp := SuccessResponse("hello", Meta{
		LastUpdated: 1700000000,
		Source:      "live",
		NextRefresh: 1700000010,
	})

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal APIResponse: %v", err)
	}

	var decoded APIResponse[string]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal APIResponse: %v", err)
	}

	if decoded.Code != 0 {
		t.Errorf("Code mismatch: got %d, want 0", decoded.Code)
	}
	if decoded.Message != "ok" {
		t.Errorf("Message mismatch: got %q, want %q", decoded.Message, "ok")
	}
	if decoded.Data != "hello" {
		t.Errorf("Data mismatch: got %q, want %q", decoded.Data, "hello")
	}
	if decoded.Meta.Source != "live" {
		t.Errorf("Meta.Source mismatch: got %q, want %q", decoded.Meta.Source, "live")
	}
}

func TestErrorResponse(t *testing.T) {
	resp := ErrorResponse(500, "internal error")

	if resp.Code != 500 {
		t.Errorf("Code mismatch: got %d, want 500", resp.Code)
	}
	if resp.Message != "internal error" {
		t.Errorf("Message mismatch: got %q, want %q", resp.Message, "internal error")
	}
	if resp.Data != nil {
		t.Errorf("Data should be nil, got %v", resp.Data)
	}
}

func TestMetaJSON(t *testing.T) {
	meta := Meta{
		LastUpdated: 1700000000,
		Source:      "cache",
		NextRefresh: 1700000300,
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("failed to marshal Meta: %v", err)
	}

	var decoded Meta
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Meta: %v", err)
	}

	if decoded.LastUpdated != meta.LastUpdated {
		t.Errorf("LastUpdated mismatch: got %d, want %d", decoded.LastUpdated, meta.LastUpdated)
	}
	if decoded.Source != meta.Source {
		t.Errorf("Source mismatch: got %q, want %q", decoded.Source, meta.Source)
	}
	if decoded.NextRefresh != meta.NextRefresh {
		t.Errorf("NextRefresh mismatch: got %d, want %d", decoded.NextRefresh, meta.NextRefresh)
	}
}
