package model

import "testing"

func TestTriggerWarningsScan_InvalidJSONReturnsError(t *testing.T) {
	var tw TriggerWarnings
	err := tw.Scan([]byte(`{"invalid_json":`))
	if err == nil {
		t.Fatal("expected scan to return error for invalid JSON bytes")
	}
}
