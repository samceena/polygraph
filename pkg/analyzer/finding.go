package analyzer

import (
	"fmt"
	"strings"
)

// Severity is used to tell how serious a finding is. Higher values mean its more sever
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityLow
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

var severityNames = [...]string{"info", "low", "medium", "high", "critical"}

func (s Severity) String() string {
	if s < SeverityInfo || s > SeverityCritical {
		return fmt.Sprintf("Severity(%d)", int(s))
	}
	return severityNames[s]
}

func ParseSeverity(v string) (Severity, error) {
	for i, name := range severityNames {
		if strings.EqualFold(strings.TrimSpace(v), name) {
			return Severity(i), nil
		}
	}
	return 0, fmt.Errorf("Unknown severity %q (want one of %s)", v, strings.Join(severityNames[:], ","))
}

func (s Severity) MarshalText() ([]byte, error) {
	if s < SeverityInfo || s > SeverityCritical {
		return nil, fmt.Errorf("invalid severity %d", int(s))
	}
	return []byte(s.String()), nil
}

func (s *Severity) UnmarshalText(b []byte) error {
	parsed, err := ParseSeverity(string(b))
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}

type Finding struct {
	CheckID        string   `json:"checkId"`
	Title          string   `json:"title"`
	Severity       Severity `json:"severity"`
	Namespace      string   `json:"namespace"`
	ResourceKind   string   `json:"resourceKind"`
	ResourceName   string   `json:"resourceName"`
	Message        string   `json:"message"`
	Recommendation string   `json:"recommendation"`
	Remediation    string   `json:"remediation,omitempty"`
	SourceFile     string   `json:"sourceFile,omitempty"`
	SourceLine     int      `json:"sourceLine,omitempty"`
}
