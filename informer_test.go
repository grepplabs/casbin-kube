package casbinkube

import (
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/grepplabs/casbin-kube/api/v1alpha1"
	"github.com/stretchr/testify/require"
)

func Test_toPolicyRuleArray(t *testing.T) {
	tests := []struct {
		name string
		r    *v1alpha1.Rule
		want []string
	}{
		{
			name: "only V0",
			r:    rule("p", "alice"),
			want: []string{"alice"},
		},
		{
			name: "V0..V2",
			r:    rule("p", "alice", "data1", "read"),
			want: []string{"alice", "data1", "read"},
		},
		{
			name: "V0..V3 last empty trimmed",
			r:    rule("p", "alice", "data1", "read", ""),
			want: []string{"alice", "data1", "read"},
		},
		{
			name: "full V0..V5",
			r:    rule("p", "v0", "v1", "v2", "v3", "v4", "v5"),
			want: []string{"v0", "v1", "v2", "v3", "v4", "v5"},
		},
		{
			name: "full V1..V5",
			r:    rule("p", "", "v1", "v2", "v3", "v4", "v5"),
			want: []string{"", "v1", "v2", "v3", "v4", "v5"},
		},
		{
			name: "full V1..V4",
			r:    rule("p", "", "v1", "v2", "v3", "v4", ""),
			want: []string{"", "v1", "v2", "v3", "v4"},
		},
		{
			name: "full V1..V4",
			r:    rule("p", "", "v1", "v2", "v3", "v4"),
			want: []string{"", "v1", "v2", "v3", "v4"},
		},
		{
			name: "all empty -> empty slice",
			r:    rule("p", "", "", "", "", "", ""),
			want: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := toPolicyRuleArray(tc.r)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_toPolicyParams(t *testing.T) {
	// Standard policy
	r := rule("p", "alice", "data1", "read")
	ptype, sec, vals := toPolicyParams(r)
	require.Equal(t, "p", ptype)
	require.Equal(t, "p", sec) // your code returns the full string as sec; here it's "p"
	require.Equal(t, []string{"alice", "data1", "read"}, vals)

	// Grouping policy (often "g")
	g := rule("g", "alice", "admin")
	ptype, sec, vals = toPolicyParams(g)
	require.Equal(t, "g", ptype)
	require.Equal(t, "g", sec)
	require.Equal(t, []string{"alice", "admin"}, vals)

	// Empty ptype -> all empty
	empty := &v1alpha1.Rule{}
	ptype, sec, vals = toPolicyParams(empty)
	require.Empty(t, ptype)
	require.Empty(t, sec)
	require.Empty(t, vals)
}

func Test_Enforcer_AddRemove_WithPolicyParamsFromRule(t *testing.T) {
	e, err := casbin.NewEnforcer("examples/rbac_model.conf")
	require.NoError(t, err)

	r := rule("p", "alice", "data1", "read")
	ptype, sec, vals := toPolicyParams(r)

	added, err := e.SelfAddPolicy(ptype, sec, vals)
	require.NoError(t, err)
	require.True(t, added)

	has, err := e.HasPolicy(vals)
	require.NoError(t, err)
	require.True(t, has)

	removed, err := e.SelfRemovePolicy(ptype, sec, vals)
	require.NoError(t, err)
	require.True(t, removed)

	has, err = e.HasPolicy(vals)
	require.NoError(t, err)
	require.False(t, has)
}

func rule(ptype string, vals ...string) *v1alpha1.Rule {
	r := &v1alpha1.Rule{}
	r.Spec.PType = ptype
	// pad to 6 for V0..V5
	buf := make([]string, 6)
	for i := range buf {
		if i < len(vals) {
			buf[i] = vals[i]
		}
	}
	r.Spec.V0, r.Spec.V1, r.Spec.V2 = buf[0], buf[1], buf[2]
	r.Spec.V3, r.Spec.V4, r.Spec.V5 = buf[3], buf[4], buf[5]
	return r
}
