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

func TestE2E(t *testing.T) { //nolint:funlen
	zlog.Init(zlog.LogConfig{
		Level:  "debug",
		Format: "text",
	})
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
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = reader.Enforce(sub, obj, act)
	require.NoError(t, err)
	require.False(t, ok)

	added, err := admin.AddPolicy(sub, obj, act)
	require.NoError(t, err)
	require.True(t, added)

	assert.Eventually(t, func() bool {
		ok, err = admin.Enforce(sub, obj, act)
		return err == nil && ok
	}, 3*time.Second, 500*time.Millisecond, "admin enforce true")

	assert.Eventually(t, func() bool {
		ok, err = reader.Enforce(sub, obj, act)
		return err == nil && ok
	}, 3*time.Second, 500*time.Millisecond, "reader enforce true")

	removed, err := admin.RemovePolicy(sub, obj, act)
	require.NoError(t, err)
	require.True(t, removed)

	assert.Eventually(t, func() bool {
		ok, err = admin.Enforce(sub, obj, act)
		return err == nil && !ok
	}, 3*time.Second, 500*time.Millisecond, "admin enforce false")

	assert.Eventually(t, func() bool {
		ok, err = reader.Enforce(sub, obj, act)
		return err == nil && !ok
	}, 3*time.Second, 500*time.Millisecond, "reader enforce false")
}

func TestK8sInformer(t *testing.T) {
	t.SkipNow()

	zlog.Init(zlog.LogConfig{
		Level:  "debug",
		Format: "",
	})
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
