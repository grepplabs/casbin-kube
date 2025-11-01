.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: build
build: ## build binary
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ./casbin-kube-converter .

.PHONY: test
test: ## test
	go run main.go --input https://raw.githubusercontent.com/casbin/casbin/refs/heads/master/examples/rbac_policy.csv
	go run main.go --input https://raw.githubusercontent.com/casbin/casbin/refs/heads/master/examples/rbac_policy.csv -o ./rbac_policy.yaml

	curl -s https://raw.githubusercontent.com/casbin/casbin/refs/heads/master/examples/keymatch_policy.csv -o ./keymatch_policy.csv
	go run main.go --input ./keymatch_policy.csv --label=casbin.grepplabs.com/model=keymatch
	go run main.go --input ./keymatch_policy.csv -o ./keymatch_policy.yaml --label=casbin.grepplabs.com/model=keymatch

.PHONY: clean
clean: ## clean
	rm -f ./casbin-kube-converter
	rm -f *.yaml
	rm -f *.csv

.PHONY: fmt
fmt: ## Go format
	go fmt ./...

.PHONY: vet
vet: ## Go vet
	go vet ./...

.PHONY: tidy
tidy: ## Go tidy
	go mod tidy
