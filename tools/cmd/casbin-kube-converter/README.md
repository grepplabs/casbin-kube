# casbin-kube-converter

`casbin-kube-converter` converts **Casbin policy CSV files** (from local files or remote URLs) into **Kubernetes CRD YAML** objects of kind [`Rule`](https://github.com/grepplabs/casbin-kube) for the [casbin-kube](https://github.com/grepplabs/casbin-kube) adapter.

This allows you to take existing Casbin policy definitions and apply them directly to Kubernetes as CRDs managed by `casbin-kube`.

---

## Features

- Converts Casbin policy CSV → `Rule` CRD YAML
- Supports both **local files** and **HTTP(S)** URLs (e.g. GitHub raw URLs)
- Adds optional **namespace** and **labels**
- Outputs YAML to **stdout** or to a **file**
- Produces deterministic resource names based on policy content

---

## Installation

```bash
go install github.com/grepplabs/casbin-kube/tools/cmd/casbin-kube-converter@latest
```

## Usage

```
$ casbin-kube-converter --help

casbin-kube-converter
----------------------
Convert Casbin policy CSV files (local or remote) into Kubernetes CRD YAML 
objects of kind 'Rule' for the casbin-kube adapter.

Examples:
  casbin-kube-converter -i policy.csv
  casbin-kube-converter -i https://raw.githubusercontent.com/casbin/casbin/refs/heads/master/examples/rbac_policy.csv
  casbin-kube-converter -i https://raw.githubusercontent.com/casbin/casbin/refs/heads/master/examples/rbac_policy.csv -o rbac_policy.yaml

  casbin-kube-converter -i keymatch_policy.csv --label=casbin.grepplabs.com/model=keymatch
  casbin-kube-converter -i ./keymatch_policy.csv -o ./keymatch_policy.yaml --label=casbin.grepplabs.com/model=keymatch

Options:
  -i, --input string           Path or URL to Casbin policy CSV (file or http/https)
      --label stringToString   Label to add to metadata.labels (repeatable: --label key=value) (default [])
  -n, --namespace string       Target namespace for generated Rules (optional)
  -o, --output string          Output file for generated YAML. Use '-' for stdout. (default "-")

``` 
