package analyzer

import (
	"encoding/json"
	"testing"
)

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		in      string
		want    Severity
		wantErr bool
	}{
		{in: "info", want: SeverityInfo},
		{in: "HIGH", want: SeverityHigh},
		{in: " Medium ", want: SeverityMedium},
		{in: "critical", want: SeverityCritical},
		{in: "none", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseSeverity(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseSeverity(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseSeverity(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestSeverityJSONRoundTrip(t *testing.T) {
	for s := SeverityInfo; s <= SeverityCritical; s++ {
		b, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("marshal %v: %v", s, err)
		}
		var back Severity
		if err := json.Unmarshal(b, &back); err != nil {
			t.Fatalf("unmarshal %s: %v", b, err)
		}
		if back != s {
			t.Errorf("round trip: got %v, want %v", back, s)
		}
	}
}

func TestSeverityOrdering(t *testing.T) {
	if !(SeverityCritical > SeverityHigh && SeverityHigh >
		SeverityMedium &&
		SeverityMedium > SeverityLow && SeverityLow > SeverityInfo) {
		t.Fatal("severities must be ordered info < low < medium < high < critical")
	}
}
