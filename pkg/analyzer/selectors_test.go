package analyzer

import (
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPolicyTypes(t *testing.T) {
	egressRule := []networkingv1.NetworkPolicyEgressRule{{}}
	ingressOnly := []networkingv1.PolicyType{networkingv1.PolicyTypeIngress}
	egressOnly := []networkingv1.PolicyType{networkingv1.PolicyTypeEgress}
	both := []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress}

	tests := []struct {
		name        string
		spec        networkingv1.NetworkPolicySpec
		wantIngress bool
		wantEgress  bool
	}{
		{"omitted, no egress rules", networkingv1.NetworkPolicySpec{}, true, false},
		{"omitted, has egress rules", networkingv1.NetworkPolicySpec{Egress: egressRule}, true, true},
		{"omitted, empty egress list", networkingv1.NetworkPolicySpec{Egress: []networkingv1.NetworkPolicyEgressRule{}}, true, false},
		{"explicit ingress only", networkingv1.NetworkPolicySpec{PolicyTypes: ingressOnly}, true, false},
		{"explicit egress only", networkingv1.NetworkPolicySpec{PolicyTypes: egressOnly}, false, true},
		{"explicit both", networkingv1.NetworkPolicySpec{PolicyTypes: both}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &networkingv1.NetworkPolicy{Spec: tt.spec}
			gotIngress, gotEgress := policyTypes(p)
			if gotIngress != tt.wantIngress || gotEgress != tt.wantEgress {
				t.Errorf("policyTypes() = (%v, %v), want (%v, %v)",
					gotIngress, gotEgress, tt.wantIngress, tt.wantEgress)
			}
		})
	}
}

func TestIsEmptySelector(t *testing.T) {
	tests := []struct {
		name string
		sel  *metav1.LabelSelector
		want bool
	}{
		{"nil", nil, false},
		{"empty", &metav1.LabelSelector{}, true},
		{"only matchLabels", &metav1.LabelSelector{
			MatchLabels: map[string]string{"app": "web"},
		}, false},
		{"only matchExpressions", &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "app", Operator: metav1.LabelSelectorOpExists},
			},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEmptySelector(tt.sel); got != tt.want {
				t.Errorf("isEmptySelector() = %v, want %v",
					got, tt.want)
			}
		})
	}
}

func TestSelectorMatches(t *testing.T) {
	expr := func(key string, op metav1.LabelSelectorOperator, values ...string) *metav1.LabelSelector {
		return &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: key, Operator: op, Values: values},
			},
		}
	}
	web := map[string]string{"app": "web"}

	tests := []struct {
		name    string
		sel     *metav1.LabelSelector
		labels  map[string]string
		want    bool
		wantErr bool
	}{
		{name: "empty matches anything", sel: &metav1.LabelSelector{}, labels: web, want: true},
		{name: "empty matches no labels", sel: &metav1.LabelSelector{}, labels: nil, want: true},
		{name: "nil matches nothing", sel: nil, labels: web, want: false},
		{name: "matchLabels hit",
			sel:    &metav1.LabelSelector{MatchLabels: web},
			labels: map[string]string{"app": "web", "tier": "fe"}, want: true},
		{name: "matchLabels miss",
			sel:    &metav1.LabelSelector{MatchLabels: web},
			labels: map[string]string{"app": "api"}, want: false},
		{name: "In", sel: expr("app", metav1.LabelSelectorOpIn,
			"web", "api"),
			labels: map[string]string{"app": "api"}, want: true},
		{name: "NotIn", sel: expr("app",
			metav1.LabelSelectorOpNotIn, "web"),
			labels: map[string]string{"app": "api"}, want: true},
		{name: "Exists", sel: expr("team",
			metav1.LabelSelectorOpExists),
			labels: web, want: false},
		{name: "DoesNotExist", sel: expr("team",
			metav1.LabelSelectorOpDoesNotExist),
			labels: web, want: true},
		{name: "invalid operator", sel: expr("app", "Bogus", "web"),
			labels: web, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectorMatches(tt.sel, tt.labels)
			if (err != nil) != tt.wantErr {
				t.Fatalf("selectorMatches() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("selectorMatches() = %v, want %v",
					got, tt.want)
			}
		})
	}
}

func TestPolicySelects(t *testing.T) {
	policy := func(ns string, podLabels map[string]string) *networkingv1.NetworkPolicy {
		return &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: ns},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{MatchLabels: podLabels},
			},
		}
	}
	workload := func(ns string, labels map[string]string) Workload {
		return Workload{Kind: "Deployment", Name: "w", Namespace: ns, PodLabels: labels}
	}
	web := map[string]string{"app": "web"}

	tests := []struct {
		name string
		p    *networkingv1.NetworkPolicy
		w    Workload
		want bool
	}{
		{"same namespace, labels match", policy("shop", web),
			workload("shop", web), true},
		{"different namespace, labels match", policy("shop", web),
			workload("other", web), false},
		{"empty podSelector selects all", policy("shop", nil),
			workload("shop", map[string]string{"app": "anything"}), true},
		{"same namespace, labels differ", policy("shop", web),
			workload("shop", map[string]string{"app": "db"}),
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := policySelects(tt.p, tt.w)
			if err != nil {
				t.Fatalf("policySelects() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("policySelects() = %v, want %v",
					got, tt.want)
			}
		})
	}
}
