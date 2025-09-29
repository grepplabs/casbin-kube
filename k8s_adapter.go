package casbinkube

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/grepplabs/casbin-kube/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type k8sAdapter struct {
	k8sClient *k8sClient[*v1alpha1.Rule, *v1alpha1.RuleList]
}

func newK8sAdapter(config *AdapterConfig) (*k8sAdapter, error) {
	kubeConfig := config.KubeConfig
	c, err := newClient(kubeConfig)
	if err != nil {
		return nil, err
	}
	namespace := kubeConfig.Namespace
	if namespace == "" {
		namespace = "default"
	}
	kc := &k8sClient[*v1alpha1.Rule, *v1alpha1.RuleList]{
		New: func() *v1alpha1.Rule {
			return &v1alpha1.Rule{}
		},
		NewList: func() *v1alpha1.RuleList {
			return &v1alpha1.RuleList{}
		},
		Client:    c,
		Namespace: namespace,
	}
	return &k8sAdapter{
		k8sClient: kc,
	}, nil
}

func keyFor(r CasbinRule) string {
	const delimiter = "\x1f" // unlikely delimiter \u001F
	parts := []string{r.PType, r.V0, r.V1, r.V2, r.V3, r.V4, r.V5}
	base := strings.Join(parts, delimiter)
	sum := sha256.Sum256([]byte(base))
	return "rule-" + hex.EncodeToString(sum[:])
}

func (s *k8sAdapter) GetAllPolicies(ctx context.Context) ([]CasbinRule, error) {
	l, err := s.k8sClient.List(ctx)
	if err != nil {
		return nil, err
	}
	type record struct {
		line            CasbinRule
		created         metav1.Time
		resourceVersion string
	}
	recs := make([]record, 0, len(l.Items))
	for _, rule := range l.Items {
		if checkResultRuleValidState(&rule) {
			recs = append(recs, record{
				line:            fromRule(&rule),
				created:         rule.CreationTimestamp,
				resourceVersion: rule.ResourceVersion,
			})
		}
	}
	sort.Slice(recs, func(i, j int) bool {
		if recs[i].created.Equal(&recs[j].created) {
			return recs[i].resourceVersion < recs[j].resourceVersion
		}
		return recs[i].created.Before(&recs[j].created)
	})
	lines := make([]CasbinRule, 0, len(recs))
	for _, r := range recs {
		lines = append(lines, r.line)
	}
	return lines, nil
}

func (s *k8sAdapter) CreatePolicy(ctx context.Context, r CasbinRule) error {
	rule := toRule(s.k8sClient.Namespace, r)
	err := s.k8sClient.Create(ctx, &rule)
	if err != nil {
		return client.IgnoreAlreadyExists(err)
	}
	return nil
}

func (s *k8sAdapter) DeletePolicy(ctx context.Context, r CasbinRule) error {
	rule, err := s.k8sClient.Get(ctx, keyFor(r))
	if err != nil {
		return client.IgnoreNotFound(err)
	}
	err = s.k8sClient.Delete(ctx, rule)
	if err != nil {
		return client.IgnoreNotFound(err)
	}
	return nil
}

func (s *k8sAdapter) DeleteAllPolicies(ctx context.Context) error {
	err := s.k8sClient.DeleteAllOf(ctx, &v1alpha1.Rule{})
	if err != nil {
		return err
	}
	return nil
}

func (s *k8sAdapter) DeleteFilteredPolicies(ctx context.Context, pattern CasbinRule) error {
	fields := map[string]string{}
	if pattern.PType != "" {
		fields["spec.ptype"] = pattern.PType
	}
	if pattern.V0 != "" {
		fields["spec.v0"] = pattern.V0
	}
	if pattern.V1 != "" {
		fields["spec.v1"] = pattern.V1
	}
	if pattern.V2 != "" {
		fields["spec.v2"] = pattern.V2
	}
	if pattern.V3 != "" {
		fields["spec.v3"] = pattern.V3
	}
	if pattern.V4 != "" {
		fields["spec.v4"] = pattern.V4
	}
	if pattern.V5 != "" {
		fields["spec.v5"] = pattern.V5
	}
	var opts []client.DeleteAllOfOption
	if len(fields) > 0 {
		opts = append(opts, client.MatchingFields(fields))
	}
	err := s.k8sClient.DeleteAllOf(ctx, &v1alpha1.Rule{}, opts...)
	if err != nil {
		return err
	}
	return nil
}

func checkResultRuleValidState(rule *v1alpha1.Rule) bool {
	return rule.GetDeletionTimestamp().IsZero()
}

func fromRule(rule *v1alpha1.Rule) CasbinRule {
	return CasbinRule{
		PType: rule.Spec.PType,
		V0:    rule.Spec.V0,
		V1:    rule.Spec.V1,
		V2:    rule.Spec.V2,
		V3:    rule.Spec.V3,
		V4:    rule.Spec.V4,
		V5:    rule.Spec.V5,
	}
}

func toRule(namespace string, cr CasbinRule) v1alpha1.Rule {
	return v1alpha1.Rule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      keyFor(cr),
			Namespace: namespace,
		},
		Spec: v1alpha1.RuleSpec{
			PType: cr.PType,
			V0:    cr.V0,
			V1:    cr.V1,
			V2:    cr.V2,
			V3:    cr.V3,
			V4:    cr.V4,
			V5:    cr.V5,
		},
	}
}
