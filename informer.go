package casbinkube

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/casbin/casbin/v3"
	"github.com/grepplabs/casbin-kube/api/v1alpha1"
	"github.com/grepplabs/loggo/zlog"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	crcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InformerConfig struct {
	// Kubernetes client configuration
	KubeConfig KubeConfig
	// Informer resync period. Prefer using nil and use enforcer.StartAutoLoadPolicy() if needed.
	SyncPeriod *time.Duration
	// SkipDisableAuto keeps Casbin AutoSave and AutoNotifyWatcher enabled if true.
	SkipDisableAuto bool
}

type Informer struct {
	enforcer   casbin.IEnforcer
	kubeConfig KubeConfig
	syncPeriod *time.Duration

	stop context.CancelFunc
}

func NewInformer(config *InformerConfig, e casbin.IEnforcer) (*Informer, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	kubeConfig := config.KubeConfig
	if kubeConfig.Namespace == "" {
		kubeConfig.Namespace = DefaultNamespace
	}
	if !config.SkipDisableAuto {
		e.EnableAutoSave(false) // must be set for readonly i.e. when it is used with informer
		e.EnableAutoNotifyWatcher(false)
	}
	return &Informer{
		enforcer:   e,
		kubeConfig: kubeConfig,
		syncPeriod: config.SyncPeriod,
	}, nil
}

func (w *Informer) Start(ctx context.Context) error { //nolint:cyclop,funlen
	ctx, cancel := context.WithCancel(ctx)
	w.stop = cancel

	ctrl.SetLogger(zlog.Logger)

	cfg, err := getRESTConfig(w.kubeConfig)
	if err != nil {
		return fmt.Errorf("get rest config err: %w", err)
	}
	opts := crcache.Options{
		Scheme: scheme,
		DefaultNamespaces: map[string]crcache.Config{
			w.kubeConfig.Namespace: {},
		},
		SyncPeriod: w.syncPeriod, // nil disables periodic resync (normal)
	}
	if len(w.kubeConfig.Labels) > 0 {
		opts.ByObject = map[client.Object]crcache.ByObject{
			&v1alpha1.Rule{}: {
				Label: labels.SelectorFromSet(w.kubeConfig.Labels),
			},
		}
	}
	c, err := crcache.New(cfg, opts)
	if err != nil {
		return fmt.Errorf("create cache err: %w", err)
	}
	inf, err := c.GetInformer(ctx, &v1alpha1.Rule{})
	if err != nil {
		return fmt.Errorf("get informer err: %w", err)
	}
	reg, err := inf.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInInitialList bool) {
			if r, ok := obj.(*v1alpha1.Rule); ok {
				level := 0 // info
				if isInInitialList {
					level = 1 // debug
				}
				zlog.Vf(level, "ADD(%t) %s/%s ptype=%s v0=%s", isInInitialList, r.Namespace, r.Name, r.Spec.PType, r.Spec.V0)
				_, err := w.enforcer.SelfAddPolicy(toPolicyParams(r))
				if err != nil {
					zlog.Errorf("add policy err: %s", err)
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			rNew, ok1 := newObj.(*v1alpha1.Rule)
			rOld, ok2 := oldObj.(*v1alpha1.Rule)
			if ok1 && ok2 {
				zlog.Infof("UPDATE %s/%s ptype=%s v0=%s", rNew.Namespace, rNew.Name, rNew.Spec.PType, rNew.Spec.V0)
				sec, ptype, newRule := toPolicyParams(rNew)
				oldRule := toPolicyRuleArray(rOld)
				_, err := w.enforcer.SelfUpdatePolicy(sec, ptype, oldRule, newRule)
				if err != nil {
					zlog.Errorf("update policy err: %s", err)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if r, ok := obj.(*v1alpha1.Rule); ok {
				zlog.Infof("DELETE %s/%s ptype=%s v0=%s", r.Namespace, r.Name, r.Spec.PType, r.Spec.V0)
				_, err := w.enforcer.SelfRemovePolicy(toPolicyParams(r))
				if err != nil {
					zlog.Errorf("remoove policy err: %s", err)
				}
			}
		},
	})
	if err != nil {
		return fmt.Errorf("adds an event handler err: %w", err)
	}
	go func() {
		defer zlog.Infof("informer stopped")

		if err := c.Start(ctx); err != nil {
			zlog.Fatalf("informer start failed: %v", err)
		}
	}()
	zlog.Infof("wait for the informer to sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), reg.HasSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync: %w", err)
	}
	zlog.Infof("informer started")
	return nil
}
func (w *Informer) Close() {
	if w.stop != nil {
		w.stop()
	}
}

func toPolicyParams(obj *v1alpha1.Rule) (string, string, []string) {
	if len(obj.Spec.PType) == 0 {
		return "", "", []string{}
	}
	return string(obj.Spec.PType[0]), obj.Spec.PType, toPolicyRuleArray(obj)
}

func toPolicyRuleArray(obj *v1alpha1.Rule) []string {
	spec := &obj.Spec
	var p = []string{spec.V0, spec.V1, spec.V2, spec.V3, spec.V4, spec.V5}
	index := len(p) - 1
	for index >= 0 && p[index] == "" {
		index--
	}
	if index < 0 {
		return []string{}
	}
	return p[:index+1]
}
