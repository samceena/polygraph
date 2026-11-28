package analyzer

import (
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func isEmptySelector(sel *metav1.LabelSelector) bool {
	return sel != nil && len(sel.MatchLabels) == 0 && len(sel.MatchExpressions) == 0
}

func selectorMatches(sel *metav1.LabelSelector, set map[string]string) (bool, error) {
	s, err := metav1.LabelSelectorAsSelector(sel)
	if err != nil {
		return false, fmt.Errorf("invalid label selector: %w", err)
	}
	return s.Matches(labels.Set(set)), nil
}

// policyTypes reports which directions (ingress, egress) a policy isolates.
// If spec.policyTypes is empty, Ingress is always implied, and Egress is
// implied only when the policy has at least one egress rule.
func policyTypes(p *networkingv1.NetworkPolicy) (ingress, egress bool) {
	if len(p.Spec.PolicyTypes) == 0 {
		return true, len(p.Spec.Egress) > 0
	}
	for _, t := range p.Spec.PolicyTypes {
		switch t {
		case networkingv1.PolicyTypeIngress:
			ingress = true
		case networkingv1.PolicyTypeEgress:
			egress = true
		}
	}
	return ingress, egress
}

func policySelects(p *networkingv1.NetworkPolicy, w Workload) (bool, error) {
	if p.Namespace != w.Namespace {
		return false, nil
	}
	return selectorMatches(&p.Spec.PodSelector, w.PodLabels)
}
