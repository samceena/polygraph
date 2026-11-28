package analyzer

import (
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

type Mode int

const (
	// ModeCluster means the input was read from a live cluster.
	ModeCluster Mode = iota
	// ModeFiles means the input was read from YAML manifest file, which may not include every object in the cluster.
	ModeFiles
)

// SourceLocation points to the YAML document an object came from (file mode only).
type SourceLocation struct {
	File string
	Line int
}

type Workload struct {
	Kind      string
	Name      string
	Namespace string
	PodLabels map[string]string
	Source    SourceLocation
}

type Input struct {
	Namespaces       []corev1.Namespace
	NetworkPolicies  []networkingv1.NetworkPolicy
	Workloads        []Workload
	Mode             Mode
	PolicySources    map[string]SourceLocation
	NamespaceSources map[string]SourceLocation
}

type Options struct {
	ExcludeNamespaces []string
	DisabledChecks    []string
}

var DefaultExcludedNamespaces = []string{
	"kube-system", "kube-public", "kube-node-lease", "local-path-storage",
}

func DefaultOptions() Options {
	return Options{
		ExcludeNamespaces: append([]string(nil), DefaultExcludedNamespaces...),
	}
}
