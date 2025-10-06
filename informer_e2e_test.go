//nolint:goconst
package casbinkube

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"
	"github.com/grepplabs/loggo/zlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestE2E(t *testing.T) {
	ctrl.SetLogger(zlog.Logger)

	adapter, err := NewAdapter(&AdapterConfig{})
	require.NoError(t, err)

	model := "examples/rbac_model.conf"

	admin, err := casbin.NewSyncedEnforcer(model, adapter)
	require.NoError(t, err)

	reader, err := casbin.NewSyncedEnforcer(model, adapter)
	require.NoError(t, err)

	informer, err := NewInformer(&InformerConfig{}, reader)
	require.NoError(t, err)
	defer informer.Close()
	err = informer.Start(context.Background())
	require.NoError(t, err)

	sub := "sub-" + uuid.NewString()
	obj := "obj-" + uuid.NewString()
	act := "read"

	ok, err := admin.Enforce(sub, obj, act)
	requireFalse(t, ok, err)

	ok, err = reader.Enforce(sub, obj, act)
	requireFalse(t, ok, err)

	added, err := admin.AddPolicy(sub, obj, act)
	requireTrue(t, added, err)

	assert.Eventually(t, func() bool {
		ok, err := admin.Enforce(sub, obj, act)
		return err == nil && ok
	}, 3*time.Second, 500*time.Millisecond, "admin enforce true")

	assert.Eventually(t, func() bool {
		ok, err := reader.Enforce(sub, obj, act)
		return err == nil && ok
	}, 3*time.Second, 500*time.Millisecond, "reader enforce true")

	removed, err := admin.RemovePolicy(sub, obj, act)
	require.NoError(t, err)
	require.True(t, removed)

	assert.Eventually(t, func() bool {
		ok, err := admin.Enforce(sub, obj, act)
		return err == nil && !ok
	}, 3*time.Second, 500*time.Millisecond, "admin enforce false")

	assert.Eventually(t, func() bool {
		ok, err := reader.Enforce(sub, obj, act)
		return err == nil && !ok
	}, 3*time.Second, 500*time.Millisecond, "reader enforce false")
}

func TestE2EDifferentLabels(t *testing.T) { //nolint:funlen
	ctrl.SetLogger(zlog.Logger)

	kubeConfig1 := KubeConfig{
		Labels: map[string]string{
			"label-selector": "value1",
		},
	}
	adapter1, err := NewAdapter(&AdapterConfig{
		KubeConfig: kubeConfig1,
	})
	require.NoError(t, err)

	kubeConfig2 := KubeConfig{
		Labels: map[string]string{
			"label-selector": "value2",
		},
	}
	adapter2, err := NewAdapter(&AdapterConfig{
		KubeConfig: kubeConfig2,
	})
	require.NoError(t, err)

	model := "examples/rbac_model.conf"

	admin1, err := casbin.NewSyncedEnforcer(model, adapter1)
	require.NoError(t, err)
	reader1, err := casbin.NewSyncedEnforcer(model, adapter1)
	require.NoError(t, err)
	reader2, err := casbin.NewSyncedEnforcer(model, adapter2)
	require.NoError(t, err)

	informer1, err := NewInformer(&InformerConfig{KubeConfig: kubeConfig1}, reader1)
	require.NoError(t, err)
	defer informer1.Close()
	err = informer1.Start(context.Background())
	require.NoError(t, err)

	informer2, err := NewInformer(&InformerConfig{KubeConfig: kubeConfig2}, reader2)
	require.NoError(t, err)
	defer informer2.Close()
	err = informer2.Start(context.Background())
	require.NoError(t, err)

	sub := "sub-" + uuid.NewString()
	obj := "obj-" + uuid.NewString()
	act := "read"

	ok, err := admin1.Enforce(sub, obj, act)
	requireFalse(t, ok, err)

	ok, err = reader1.Enforce(sub, obj, act)
	requireFalse(t, ok, err)

	ok, err = reader2.Enforce(sub, obj, act)
	requireFalse(t, ok, err)

	added, err := admin1.AddPolicy(sub, obj, act)
	requireTrue(t, added, err)

	assert.Eventually(t, func() bool {
		ok, err := admin1.Enforce(sub, obj, act)
		return err == nil && ok
	}, 3*time.Second, 500*time.Millisecond, "admin1 enforce true")

	assert.Eventually(t, func() bool {
		ok, err := reader1.Enforce(sub, obj, act)
		return err == nil && ok
	}, 3*time.Second, 500*time.Millisecond, "reader1 enforce true")

	assert.Never(t, func() bool {
		ok, err := reader2.Enforce(sub, obj, act)
		return err == nil && ok
	}, 2*time.Second, 500*time.Millisecond, "reader1 enforce false")

	removed, err := admin1.RemovePolicy(sub, obj, act)
	requireTrue(t, removed, err)

	assert.Eventually(t, func() bool {
		ok, err := admin1.Enforce(sub, obj, act)
		return err == nil && !ok
	}, 3*time.Second, 500*time.Millisecond, "admin1 enforce false")

	assert.Eventually(t, func() bool {
		ok, err := reader1.Enforce(sub, obj, act)
		return err == nil && !ok
	}, 3*time.Second, 500*time.Millisecond, "reader1 enforce false")

	assert.Never(t, func() bool {
		ok, err := reader2.Enforce(sub, obj, act)
		return err == nil && ok
	}, 2*time.Second, 500*time.Millisecond, "reader2 enforce false")
}

func TestE2ESameLabels(t *testing.T) { //nolint:funlen
	ctrl.SetLogger(zlog.Logger)

	kubeConfig := KubeConfig{
		Labels: map[string]string{
			"label-selector": "value",
		},
	}
	adapter1, err := NewAdapter(&AdapterConfig{
		KubeConfig: kubeConfig,
	})
	require.NoError(t, err)

	adapter2, err := NewAdapter(&AdapterConfig{
		KubeConfig: kubeConfig,
	})
	require.NoError(t, err)

	model := "examples/rbac_model.conf"

	admin1, err := casbin.NewSyncedEnforcer(model, adapter1)
	require.NoError(t, err)
	reader1, err := casbin.NewSyncedEnforcer(model, adapter1)
	require.NoError(t, err)
	reader2, err := casbin.NewSyncedEnforcer(model, adapter2)
	require.NoError(t, err)

	informer1, err := NewInformer(&InformerConfig{KubeConfig: kubeConfig}, reader1)
	require.NoError(t, err)
	defer informer1.Close()
	err = informer1.Start(context.Background())
	require.NoError(t, err)

	informer2, err := NewInformer(&InformerConfig{KubeConfig: kubeConfig}, reader2)
	require.NoError(t, err)
	defer informer2.Close()
	err = informer2.Start(context.Background())
	require.NoError(t, err)

	sub := "sub-" + uuid.NewString()
	obj := "obj-" + uuid.NewString()
	act := "read"

	ok, err := admin1.Enforce(sub, obj, act)
	requireFalse(t, ok, err)

	ok, err = reader1.Enforce(sub, obj, act)
	requireFalse(t, ok, err)

	ok, err = reader2.Enforce(sub, obj, act)
	requireFalse(t, ok, err)

	added, err := admin1.AddPolicy(sub, obj, act)
	requireTrue(t, added, err)

	assert.Eventually(t, func() bool {
		ok, err := admin1.Enforce(sub, obj, act)
		return err == nil && ok
	}, 3*time.Second, 500*time.Millisecond, "admin1 enforce true")

	assert.Eventually(t, func() bool {
		ok, err := reader1.Enforce(sub, obj, act)
		return err == nil && ok
	}, 3*time.Second, 500*time.Millisecond, "reader1 enforce true")

	assert.Eventually(t, func() bool {
		ok, err := reader2.Enforce(sub, obj, act)
		return err == nil && ok
	}, 3*time.Second, 500*time.Millisecond, "reader2 enforce true")

	removed, err := admin1.RemovePolicy(sub, obj, act)
	require.NoError(t, err)
	require.True(t, removed)

	assert.Eventually(t, func() bool {
		ok, err := admin1.Enforce(sub, obj, act)
		return err == nil && !ok
	}, 3*time.Second, 500*time.Millisecond, "admin1 enforce false")

	assert.Eventually(t, func() bool {
		ok, err := reader1.Enforce(sub, obj, act)
		return err == nil && !ok
	}, 3*time.Second, 500*time.Millisecond, "reader1 enforce false")

	assert.Eventually(t, func() bool {
		ok, err := reader2.Enforce(sub, obj, act)
		return err == nil && !ok
	}, 3*time.Second, 500*time.Millisecond, "reader2 enforce false")
}

func TestK8sInformer(t *testing.T) {
	t.SkipNow()
	ctrl.SetLogger(zlog.Logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	adapter, err := NewAdapter(&AdapterConfig{})
	require.NoError(t, err)

	e, err := casbin.NewSyncedEnforcer("examples/rbac_model.conf", adapter)
	require.NoError(t, err)

	i, err := NewInformer(&InformerConfig{}, e)
	require.NoError(t, err)
	defer i.Close()

	err = i.Start(ctx)
	require.NoError(t, err)

	time.Sleep(5 * time.Second)
	// <-ctx.Done()
}
