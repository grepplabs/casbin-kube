package casbinkube

import (
	"context"
	"log"
	"sort"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/casbin/casbin/v2/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	adapter, err := NewAdapter(&AdapterConfig{})
	require.NoError(t, err)

	_, err = casbin.NewEnforcer("examples/rbac_model.conf", adapter)
	require.NoError(t, err)
}

func TestExample(t *testing.T) {
	srcAdapter := fileadapter.NewAdapter("examples/rbac_policy.csv")
	m, err := model.NewModelFromFile("examples/rbac_model.conf")
	require.NoError(t, err)

	err = srcAdapter.LoadPolicy(m)
	require.NoError(t, err)

	adapter, err := NewAdapter(&AdapterConfig{})
	require.NoError(t, err)

	err = adapter.SavePolicy(m)
	require.NoError(t, err)

	_, err = casbin.NewEnforcer("examples/rbac_model.conf", adapter)
	require.NoError(t, err)

	m, err = model.NewModelFromFile("examples/rbac_model.conf")
	require.NoError(t, err)

	err = adapter.LoadPolicy(m)
	require.NoError(t, err)
}

func TestLabels(t *testing.T) {
	labels := map[string]string{
		"label1": "value1",
		"label2": "value2",
	}
	ac := &AdapterConfig{
		KubeConfig: KubeConfig{
			Labels: labels,
		},
	}
	adapter, err := NewAdapter(ac)
	require.NoError(t, err)
	e, err := casbin.NewEnforcer("examples/rbac_model.conf", adapter)
	require.NoError(t, err)

	userId := "user-" + uuid.NewString()

	ok, err := e.AddPolicy(userId, "data1", "read")
	require.NoError(t, err)
	require.True(t, ok)

	k8s, err := newK8sAdapter(ac)
	require.NoError(t, err)

	cr, err := k8s.k8sClient.Get(context.Background(), keyFor(CasbinRule{
		PType: "p",
		V0:    userId,
		V1:    "data1",
		V2:    "read",
	}))
	require.NoError(t, err)
	require.Equal(t, cr.GetLabels(), labels)
	require.Equal(t, "p", cr.Spec.PType)
	require.Equal(t, cr.Spec.V0, userId)
	require.Equal(t, "data1", cr.Spec.V1)
	require.Equal(t, "read", cr.Spec.V2)
	require.Empty(t, cr.Spec.V3)
	require.Empty(t, cr.Spec.V4)
	require.Empty(t, cr.Spec.V5)
}

func testGetPolicy(t *testing.T, e *casbin.Enforcer, res [][]string) {
	t.Helper()
	myRes, err := e.GetPolicy()
	if err != nil {
		panic(err)
	}

	log.Print("Policy: ", myRes)

	sort2D(res)
	sort2D(myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Policy: ", myRes, ", supposed to be ", res)
	}
}

func sort2D(arr [][]string) {
	sort.Slice(arr, func(i, j int) bool {
		for k := 0; k < len(arr[i]) && k < len(arr[j]); k++ {
			if arr[i][k] != arr[j][k] {
				return arr[i][k] < arr[j][k]
			}
		}
		return len(arr[i]) < len(arr[j])
	})
}

func initPolicy(t *testing.T, a *Adapter) {
	t.Helper()
	// Because the DB is empty at first,
	// so we need to load the policy from the file adapter (.CSV) first.
	e, err := casbin.NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	require.NoError(t, err)

	// This is a trick to save the current policy to the DB.
	// We can't call e.SavePolicy() because the adapter in the enforcer is still the file adapter.
	// The current policy means the policy in the Casbin enforcer (aka in memory).
	err = a.SavePolicy(e.GetModel())
	require.NoError(t, err)

	// Clear the current policy.
	e.ClearPolicy()
	testGetPolicy(t, e, [][]string{})

	// Load the policy from DB.
	err = a.LoadPolicy(e.GetModel())
	require.NoError(t, err)
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func testSaveLoad(t *testing.T, a *Adapter) {
	t.Helper()
	// Initialize some policy in DB.
	initPolicy(t, a)
	// Note: you don't need to look at the above code
	// if you already have a working DB with policy inside.

	// Now the DB has policy, so we can provide a normal use case.
	// Create an adapter and an enforcer.
	// NewEnforcer() will load the policy automatically.
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}
func testAutoSave(t *testing.T, a *Adapter) {
	t.Helper()
	// NewEnforcer() will load the policy automatically.
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)
	// AutoSave is enabled by default.
	// Now we disable it.
	e.EnableAutoSave(false)

	// Because AutoSave is disabled, the policy change only affects the policy in Casbin enforcer,
	// it doesn't affect the policy in the storage.
	e.AddPolicy("alice", "data1", "write")
	// Reload the policy from the storage to see the effect.
	e.LoadPolicy()
	// This is still the original policy.
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	// Now we enable the AutoSave.
	e.EnableAutoSave(true)

	// Because AutoSave is enabled, the policy change not only affects the policy in Casbin enforcer,
	// but also affects the policy in the storage.
	e.AddPolicy("alice", "data1", "write")
	// Reload the policy from the storage to see the effect.
	e.LoadPolicy()
	// The policy has a new rule: {"alice", "data1", "write"}.
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"alice", "data1", "write"}})

	// Remove the added rule.
	e.RemovePolicy("alice", "data1", "write")
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	// Remove "data2_admin" related policy rules via a filter.
	// Two rules: {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"} are deleted.
	e.RemoveFilteredPolicy(0, "data2_admin")
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}})
}

func TestNilField(t *testing.T) {
	a, err := NewAdapter(&AdapterConfig{})
	require.NoError(t, err)

	e, err := casbin.NewEnforcer("examples/rbac_model.conf", a)
	require.NoError(t, err)
	e.EnableAutoSave(false)

	_, err = e.AddPolicy("", "data1", "write")
	require.NoError(t, err)
	e.SavePolicy()
	require.NoError(t, e.LoadPolicy())

	ok, err := e.Enforce("", "data1", "write")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestAdapter(t *testing.T) {
	a, err := NewAdapter(&AdapterConfig{})
	require.NoError(t, err)
	testSaveLoad(t, a)
	testAutoSave(t, a)
}

func TestRemoveFilteredPolicy0(t *testing.T) {
	a, err := NewAdapter(&AdapterConfig{})
	require.NoError(t, err)
	m := model.NewModel()
	require.NoError(t, err)

	// empty model delete all policies
	err = a.SavePolicy(m)
	require.NoError(t, err)

	e, err := casbin.NewEnforcer("examples/rbac_model.conf", a)
	require.NoError(t, err)

	_, err = e.AddPolicy("alice", "data1", "write")
	require.NoError(t, err)
	_, err = e.AddPolicy("alice", "data2", "write")
	require.NoError(t, err)
	_, err = e.AddPolicy("bob", "data1", "write")
	require.NoError(t, err)
	_, err = e.AddPolicy("bob", "data2", "write")
	require.NoError(t, err)

	_, err = e.RemoveFilteredNamedPolicy("p", 0, "alice")
	require.NoError(t, err)
	sub, err := e.GetAllSubjects()
	require.NoError(t, err)
	obj, err := e.GetAllObjects()
	require.NoError(t, err)

	require.Equal(t, []string{"bob"}, sub)
	require.Equal(t, []string{"data1", "data2"}, obj)
}

func TestRemoveFilteredPolicy1(t *testing.T) {
	a, err := NewAdapter(&AdapterConfig{})
	require.NoError(t, err)
	m := model.NewModel()
	require.NoError(t, err)

	// empty model delete all policies
	err = a.SavePolicy(m)
	require.NoError(t, err)

	e, err := casbin.NewEnforcer("examples/rbac_model.conf", a)
	require.NoError(t, err)

	_, err = e.AddPolicy("alice", "data1", "write")
	require.NoError(t, err)
	_, err = e.AddPolicy("alice", "data2", "write")
	require.NoError(t, err)
	_, err = e.AddPolicy("bob", "data1", "write")
	require.NoError(t, err)
	_, err = e.AddPolicy("bob", "data2", "write")
	require.NoError(t, err)

	_, err = e.RemoveFilteredNamedPolicy("p", 1, "data1")
	require.NoError(t, err)
	sub, err := e.GetAllSubjects()
	require.NoError(t, err)
	obj, err := e.GetAllObjects()
	require.NoError(t, err)
	require.Equal(t, []string{"alice", "bob"}, sub)
	require.Equal(t, []string{"data2"}, obj)
}
