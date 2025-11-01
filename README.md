Casbin Kube
====

[![Build](https://github.com/grepplabs/casbin-kube/actions/workflows/ci.yml/badge.svg)](https://github.com/grepplabs/casbin-kube/actions/workflows/ci.yml)
[![Godoc](https://godoc.org/github.com/casbin/casbin?status.svg)](https://pkg.go.dev/github.com/grepplabs/casbin-kube)

Casbin Kube is the [Kubernetes](https://kubernetes.io/) adapter for [Casbin](https://github.com/casbin/casbin). This library allows Casbin to load policies from Kubernetes and save policies back to it

The adapter integrates with the Kubernetes **Informer** mechanism to notify about policy changes.

## Kubernetes 

You need to install the `rules.casbin.grepplabs.com` custom resource and grant access to this CRD

```
kubectl apply -k config/crds
kubectl apply -k config/rbac
```
or

```
helm install casbin-kube oci://ghcr.io/grepplabs/helm/casbin-kube:0.0.1
```

## Tools
### [casbin-kube-converter](tools/cmd/casbin-kube-converter/README.md)

## Installation

    go get github.com/grepplabs/casbin-kube

## Usage Examples

### Sample data

```yaml
apiVersion: casbin.grepplabs.com/v1alpha1
kind: Rule
metadata:
  name: rule-sample
spec:
  ptype: "p"
  v0: "alice"
  v1: "data"
  v2: "read"
``` 

### Policy editor / admin 

```go
package main

import (
    "github.com/casbin/casbin/v2"
    casbinkube "github.com/grepplabs/casbin-kube"
)

func main() {
    // Initialize a casbin kube adapter and use it in a Casbin enforcer:
    kubeconfig := casbinkube.KubeConfig{}
    a, _ := casbinkube.NewAdapter(&casbinkube.AdapterConfig{KubeConfig: kubeconfig})
    e, _ := casbin.NewSyncedEnforcer("examples/rbac_model.conf", a)

    // Load the policy from Kubernetes.
    e.LoadPolicy()

    // Check the permission.
    e.Enforce("alice", "data1", "read")

    // Modify the policy.
    // e.AddPolicy(...)
    // e.RemovePolicy(...)

    // Save the policy back to Kubernetes.
    e.SavePolicy()
}
```

### Policy reader / enforcer

Casbin provides a [watcher](https://casbin.org/docs/watchers) mechanism to maintain consistency between multiple Casbin enforcer instances. 
Watchers can still be used with the adapter, but `Casbin Kube` natively supports the Kubernetes `Informer` mechanism to notify about policy changes, 
which eliminates the need for a watcher.

The Informer will automatically disable auto-save (`e.EnableAutoSave(false)`) and auto-notify watcher (`e.EnableAutoNotifyWatcher(false)`).

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/casbin/casbin/v2"
	casbinkube "github.com/grepplabs/casbin-kube"
	"github.com/grepplabs/loggo/zlog"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	ctrl.SetLogger(zlog.Logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize a casbin kube adapter and use it in a Casbin enforcer:
	kubeconfig := casbinkube.KubeConfig{}
	a, _ := casbinkube.NewAdapter(&casbinkube.AdapterConfig{KubeConfig: kubeconfig})
	e, _ := casbin.NewSyncedEnforcer("examples/rbac_model.conf", a)

	i, _ := casbinkube.NewInformer(&casbinkube.InformerConfig{KubeConfig: kubeconfig}, e)
	defer i.Close()
	i.Start(ctx)

	// Check the permission.
	e.Enforce("alice", "data1", "read")

}
```