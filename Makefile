SHELL := /usr/bin/env bash
.SHELLFLAGS += -o pipefail -O extglob
.DEFAULT_GOAL := prepare

CONTROLLER_TOOLS_VERSION := v0.19.0
GOLANGCI_LINT_VERSION := v2.4.0

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


## Tool Binaries
GO_RUN := go run
CONTROLLER_GEN ?= $(GO_RUN) sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
GOLANGCI_LINT ?= $(GO_RUN) github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: manifests
manifests: ## Generate CustomResourceDefinition objects.
	$(CONTROLLER_GEN) crd paths="./api/..." output:crd:dir=config/crds
	$(CONTROLLER_GEN) crd paths="./api/..." output:crd:dir=charts/casbin-kube/crds

.PHONY: generate
generate: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object paths="./api/..."

.PHONY: prepare
prepare: generate manifests ## Run generic prepare steps


.PHONY: lint
lint: ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

.PHONY: lint-config
lint-config: ## Verify golangci-lint linter configuration
	$(GOLANGCI_LINT) config verify


##@ Test targets

.PHONY: test
test: ## run tests
	go test -v -race -count=1 ./...

.PHONY: benchmark
benchmark: ## run benchmarks
	go test -bench=. -benchmem ./...

.PHONY: init-kind-cluster
init-kind-cluster: export KUBECONFIG=tmp/casbin-kube-kubeconfig.yaml
init-kind-cluster:
	mkdir -p tmp
	kind get clusters | grep '^casbin-kube$$' && kind export kubeconfig --name casbin-kube --kubeconfig tmp/casbin-kube-kubeconfig.yaml || \
		kind create cluster --name casbin-kube --config scripts/kind/kubernetes-1.32.yaml --kubeconfig tmp/casbin-kube-kubeconfig.yaml
	kubectl apply -k config/crds --force-conflicts --server-side=true
	kubectl apply -k config/rbac --force-conflicts --server-side=true

.PHONY: delete-kind-cluster
delete-kind-cluster: ## deletes the kind cluster
	rm -f tmp/casbin-kube-kubeconfig.yaml
	kind delete cluster --name casbin-kube


##@ Examples targets
.PHONY: example-docker-build
example-docker-build: ## Build docker build with examples/main.go
	docker build -f examples/Dockerfile . -t grepplabs/casbin-kube-example:latest

.PHONY: example-run
example-run: ## run examples/main.go
	go run 	examples/main.go

.PHONY: example-deploy
example-deploy: ## deploy example to  kind cluster
	kind load docker-image grepplabs/casbin-kube-example:latest --name casbin-kube
	kubectl apply -f scripts/examples/deploy/
	kubectl apply -k config/samples
	kubectl wait --for=condition=Available -l app=casbin-kube-example --timeout=2m deployment

.PHONY: example-run-in-cluster
example-run-in-cluster: example-docker-build
example-run-in-cluster: init-kind-cluster
example-run-in-cluster: export KUBECONFIG=tmp/casbin-kube-kubeconfig.yaml
example-run-in-cluster: example-deploy
example-run-in-cluster: ## run example in kind cluster
