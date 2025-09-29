package casbinkube

import (
	"context"

	casbinv1alpha1 "github.com/grepplabs/casbin-kube/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(casbinv1alpha1.AddToScheme(scheme))
}

const (
	DefaultGracePeriodSeconds = 0
)

type KubeConfig struct {
	Context   string
	Namespace string
	Path      string
}

type k8sClient[T client.Object, L client.ObjectList] struct {
	Client    client.Client
	Namespace string
	New       func() T
	NewList   func() L
}

func (k *k8sClient[T, L]) Create(ctx context.Context, obj T, opts ...client.CreateOption) error {
	if obj.GetNamespace() == "" {
		obj.SetNamespace(k.Namespace)
	}
	return k.Client.Create(ctx, obj, opts...)
}

func (k *k8sClient[T, L]) Get(ctx context.Context, name string) (T, error) {
	out := k.New()
	err := k.Client.Get(ctx, types.NamespacedName{
		Namespace: k.Namespace,
		Name:      name,
	}, out)
	return out, err
}

func (k *k8sClient[T, L]) Update(ctx context.Context, obj T, opts ...client.UpdateOption) error {
	return k.Client.Update(ctx, obj, opts...)
}

func (k *k8sClient[T, L]) Delete(ctx context.Context, obj T, opts ...client.DeleteOption) error {
	if obj.GetNamespace() == "" && k.Namespace != "" {
		obj.SetNamespace(k.Namespace)
	}
	deleteOpts := append([]client.DeleteOption{client.GracePeriodSeconds(DefaultGracePeriodSeconds)}, opts...)
	return k.Client.Delete(ctx, obj, deleteOpts...)
}

func (k *k8sClient[T, L]) DeleteAllOf(ctx context.Context, obj T, opts ...client.DeleteAllOfOption) error {
	deleteOpts := []client.DeleteAllOfOption{
		client.GracePeriodSeconds(DefaultGracePeriodSeconds),
	}
	ns := obj.GetNamespace()
	if ns == "" {
		ns = k.Namespace
	}
	deleteOpts = append(deleteOpts, client.InNamespace(ns))
	deleteOpts = append(deleteOpts, opts...)
	return k.Client.DeleteAllOf(ctx, obj, deleteOpts...)
}

func (k *k8sClient[T, L]) List(ctx context.Context, opts ...client.ListOption) (L, error) {
	list := k.NewList()
	listOpts := append([]client.ListOption{client.InNamespace(k.Namespace)}, opts...)
	err := k.Client.List(ctx, list, listOpts...)
	return list, err
}

func (k *k8sClient[T, L]) Patch(ctx context.Context, obj T, p client.Patch, opts ...client.PatchOption) error {
	return k.Client.Patch(ctx, obj, p, opts...)
}

func newClient(kubeConfig KubeConfig) (client.Client, error) {
	clusterConfig, err := getRESTConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	// disable rate limiter
	clusterConfig.QPS = -1
	clusterConfig.Burst = -1

	c, err := client.New(clusterConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	return c, nil
}

func getRESTConfig(kubeConfig KubeConfig) (*rest.Config, error) {
	loading := &clientcmd.ClientConfigLoadingRules{}

	if kubeConfig.Path != "" {
		loading.ExplicitPath = kubeConfig.Path
	} else {
		loading = clientcmd.NewDefaultClientConfigLoadingRules()
	}

	overrides := &clientcmd.ConfigOverrides{}
	if kubeConfig.Context != "" {
		overrides.CurrentContext = kubeConfig.Context
	}
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, overrides).ClientConfig()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
