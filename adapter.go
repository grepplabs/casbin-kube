package casbinkube

import (
	"context"
	"errors"
	"time"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/grepplabs/loggo/zlog"
)

// CasbinRule is used to determine which policy line to load.
type CasbinRule struct {
	PType string `json:"ptype"`
	V0    string `json:"v0,omitempty"`
	V1    string `json:"v1,omitempty"`
	V2    string `json:"v2,omitempty"`
	V3    string `json:"v3,omitempty"`
	V4    string `json:"v4,omitempty"`
	V5    string `json:"v5,omitempty"`
}

type AdapterConfig struct {
	// Kubernetes client configuration
	KubeConfig KubeConfig
}

type Adapter struct {
	store *k8sAdapter
}

var _ persist.BatchAdapter = &Adapter{}

func NewAdapter(config *AdapterConfig) (*Adapter, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	s, err := newK8sAdapter(config)
	if err != nil {
		return nil, err
	}
	a := &Adapter{
		store: s,
	}
	return a, nil
}

func loadPolicyLine(line CasbinRule, model model.Model) error {
	var p = []string{line.PType, line.V0, line.V1, line.V2, line.V3, line.V4, line.V5}
	index := len(p) - 1
	for index >= 0 && p[index] == "" {
		index--
	}
	if index < 0 {
		p = []string{}
	} else {
		p = p[:index+1]
	}
	return persist.LoadPolicyArray(p, model)
}

func (a *Adapter) savePolicyLine(ptype string, rule []string) CasbinRule {
	line := CasbinRule{}
	line.PType = ptype
	if len(rule) > 0 {
		line.V0 = rule[0]
	}
	if len(rule) > 1 {
		line.V1 = rule[1]
	}
	if len(rule) > 2 {
		line.V2 = rule[2]
	}
	if len(rule) > 3 {
		line.V3 = rule[3]
	}
	if len(rule) > 4 {
		line.V4 = rule[4]
	}
	if len(rule) > 5 {
		line.V5 = rule[5]
	}
	return line
}

// LoadPolicy loads all policy rules from the storage.
func (a *Adapter) LoadPolicy(model model.Model) error {
	return a.LoadPolicyCtx(context.Background(), model)
}

func (a *Adapter) LoadPolicyCtx(ctx context.Context, model model.Model) error {
	defer logDuration("loading policies", time.Now())
	zlog.Debugw("loading policies")
	lines, err := a.store.GetAllPolicies(ctx)
	if err != nil {
		return err
	}
	zlog.Infow("loading policies count", "count", len(lines))
	for _, line := range lines {
		err := loadPolicyLine(line, model)
		if err != nil {
			return err
		}
	}
	return nil
}

// SavePolicy saves all policy rules to the storage.
func (a *Adapter) SavePolicy(model model.Model) error {
	return a.SavePolicyCtx(context.Background(), model)
}

// SavePolicyCtx saves policy to the storage.
func (a *Adapter) SavePolicyCtx(ctx context.Context, model model.Model) error {
	defer logDuration("saving policies", time.Now())
	zlog.Debugw("saving policies")

	err := a.store.DeleteAllPolicies(ctx)
	if err != nil {
		return err
	}

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			line := a.savePolicyLine(ptype, rule)
			err = a.store.CreatePolicy(ctx, line)
			if err != nil {
				return err
			}
		}
	}
	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			line := a.savePolicyLine(ptype, rule)
			err = a.store.CreatePolicy(ctx, line)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AddPolicy adds a policy rule to the storage.
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	return a.AddPolicyCtx(context.Background(), sec, ptype, rule)
}

// AddPolicies adds policy rules to the storage.
func (a *Adapter) AddPolicies(sec string, ptype string, rules [][]string) error {
	ctx := context.Background()
	for _, rule := range rules {
		err := a.AddPolicyCtx(ctx, sec, ptype, rule)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddPolicyCtx adds a policy rule to the storage.
func (a *Adapter) AddPolicyCtx(ctx context.Context, sec string, ptype string, rule []string) error {
	line := a.savePolicyLine(ptype, rule)
	return a.store.CreatePolicy(ctx, line)
}

// RemovePolicy removes a policy rule from the storage.
func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return a.RemovePolicyCtx(context.Background(), sec, ptype, rule)
}

// RemovePolicies removes policy rules from the storage.
func (a *Adapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
	ctx := context.Background()
	for _, rule := range rules {
		err := a.RemovePolicyCtx(ctx, sec, ptype, rule)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemovePolicyCtx removes a policy rule from the storage.
func (a *Adapter) RemovePolicyCtx(ctx context.Context, sec string, ptype string, rule []string) error {
	line := a.savePolicyLine(ptype, rule)
	return a.store.DeletePolicy(ctx, line)
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return a.RemoveFilteredPolicyCtx(context.Background(), sec, ptype, fieldIndex, fieldValues...)
}

// RemoveFilteredPolicyCtx removes policy rules that match the filter from the storage.
func (a *Adapter) RemoveFilteredPolicyCtx(ctx context.Context, sec string, ptype string, fieldIndex int, fieldValues ...string) error { //nolint:cyclop
	line := CasbinRule{}
	line.PType = ptype
	if fieldIndex == -1 {
		return a.store.DeleteFilteredPolicies(ctx, line)
	}
	err := a.checkQueryField(fieldValues)
	if err != nil {
		return err
	}
	if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) {
		line.V0 = fieldValues[0-fieldIndex]
	}
	if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) {
		line.V1 = fieldValues[1-fieldIndex]
	}
	if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) {
		line.V2 = fieldValues[2-fieldIndex]
	}
	if fieldIndex <= 3 && 3 < fieldIndex+len(fieldValues) {
		line.V3 = fieldValues[3-fieldIndex]
	}
	if fieldIndex <= 4 && 4 < fieldIndex+len(fieldValues) {
		line.V4 = fieldValues[4-fieldIndex]
	}
	if fieldIndex <= 5 && 5 < fieldIndex+len(fieldValues) {
		line.V5 = fieldValues[5-fieldIndex]
	}
	return a.store.DeleteFilteredPolicies(ctx, line)
}

func (a *Adapter) checkQueryField(fieldValues []string) error {
	for _, fieldValue := range fieldValues {
		if fieldValue != "" {
			return nil
		}
	}
	return errors.New("the query field cannot all be empty strings")
}

func logDuration(name string, start time.Time) {
	zlog.Infof("finished %s, elapsed=%v", name, time.Since(start))
}
