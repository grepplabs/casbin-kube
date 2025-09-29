package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	ctrl "sigs.k8s.io/controller-runtime"

	casbinkube "github.com/grepplabs/casbin-kube"
	"github.com/grepplabs/casbin-kube/pkg/logger"
)

//go:embed *.conf
var FS embed.FS

func main() {
	logger.Init(logger.LogConfig{Level: "debug", Format: "text"})
	ctrl.SetLogger(logger.Logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	kubeconfig := casbinkube.KubeConfig{}
	adapter := noError(casbinkube.NewAdapter(&casbinkube.AdapterConfig{KubeConfig: kubeconfig}))

	// casbin.NewEnforcer("rbac_model.conf", adapter)
	m := noError(loadModelFromFS("rbac_model.conf"))
	enforcer := noError(casbin.NewSyncedEnforcer(m, adapter))

	//// See also InformerConfig.SkipDisableAuto
	// enforcer.EnableAutoSave(false) // must be set for readonly i.e. when it is used with informer
	// enforcer.EnableAutoNotifyWatcher(false)
	// enforcer.EnableAutoNotifyDispatcher(false)

	// enforcer.StartAutoLoadPolicy(15 * time.Minute)

	informer := noError(casbinkube.NewInformer(&casbinkube.InformerConfig{KubeConfig: kubeconfig}, enforcer))
	defer informer.Close()
	checkNoError(informer.Start(ctx))

	<-ctx.Done()
}

func noError[T any](t T, err error) T {
	checkNoError(err)
	return t
}

func checkNoError(err error) {
	if err != nil {
		logger.Fatalf("unexpected error: %v", err)
	}
}

func loadModelFromFS(name string) (model.Model, error) {
	data, err := FS.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("error reading model %s: %w", name, err)
	}
	m, err := model.NewModelFromString(string(data))
	if err != nil {
		return nil, fmt.Errorf("error parsing model %s: %w", name, err)
	}
	return m, nil
}
