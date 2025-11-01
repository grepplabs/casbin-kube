package main

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"
)

type RuleSpec struct {
	PType string `yaml:"ptype" json:"ptype"`
	V0    string `yaml:"v0,omitempty" json:"v0,omitempty"`
	V1    string `yaml:"v1,omitempty" json:"v1,omitempty"`
	V2    string `yaml:"v2,omitempty" json:"v2,omitempty"`
	V3    string `yaml:"v3,omitempty" json:"v3,omitempty"`
	V4    string `yaml:"v4,omitempty" json:"v4,omitempty"`
	V5    string `yaml:"v5,omitempty" json:"v5,omitempty"`
}

type RuleMetadataYAML struct {
	Name      string            `yaml:"name" json:"name"`
	Namespace string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Labels    map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

type RuleYAML struct {
	APIVersion string           `yaml:"apiVersion" json:"apiVersion"`
	Kind       string           `yaml:"kind" json:"kind"`
	Metadata   RuleMetadataYAML `yaml:"metadata" json:"metadata"`
	Spec       RuleSpec         `yaml:"spec" json:"spec"`
}

func readPolicyContent(src string) (string, error) {
	u, err := url.Parse(src)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		rawURL := u.String()

		resp, err := http.Get(rawURL)
		if err != nil {
			return "", fmt.Errorf("http get %q: %w", rawURL, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return "", fmt.Errorf("http get %q: unexpected status %s", rawURL, resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("read http body: %w", err)
		}

		ct := resp.Header.Get("Content-Type")
		if ct != "" && !strings.HasPrefix(ct, "text/") && !strings.Contains(ct, "csv") {
			return "", fmt.Errorf("unexpected Content-Type %q (did you use a raw URL?)", ct)
		}
		if len(body) > 0 && strings.HasPrefix(strings.TrimSpace(string(body)), "<!DOCTYPE html>") {
			return "", fmt.Errorf("fetched HTML not policy (did you pass a GitHub blob URL instead of raw?)")
		}
		return string(body), nil
	}

	b, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("read file %q: %w", src, err)
	}
	return string(b), nil
}

func parsePolicyContent(content string) ([]RuleSpec, error) {
	var specs []RuleSpec

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		r := csv.NewReader(strings.NewReader(line))
		r.Comma = ','
		r.Comment = '#'
		r.TrimLeadingSpace = true

		tokens, err := r.Read()
		if err != nil {
			return nil, fmt.Errorf("parse csv line %q: %w", line, err)
		}
		if len(tokens) == 0 {
			continue
		}

		ptype := tokens[0]
		fields := tokens[1:]

		rs := RuleSpec{PType: ptype}
		if len(fields) > 0 {
			rs.V0 = fields[0]
		}
		if len(fields) > 1 {
			rs.V1 = fields[1]
		}
		if len(fields) > 2 {
			rs.V2 = fields[2]
		}
		if len(fields) > 3 {
			rs.V3 = fields[3]
		}
		if len(fields) > 4 {
			rs.V4 = fields[4]
		}
		if len(fields) > 5 {
			rs.V5 = fields[5]
		}
		specs = append(specs, rs)
	}
	return specs, nil
}

func buildName(spec RuleSpec) string {
	const delimiter = "\x1f"
	parts := []string{spec.PType, spec.V0, spec.V1, spec.V2, spec.V3, spec.V4, spec.V5}
	base := strings.Join(parts, delimiter)
	sum := sha256.Sum256([]byte(base))
	return "rule-" + hex.EncodeToString(sum[:])
}

func buildRuleYAML(spec RuleSpec, ns string, labels map[string]string) RuleYAML {
	return RuleYAML{
		APIVersion: "casbin.grepplabs.com/v1alpha1",
		Kind:       "Rule",
		Metadata: RuleMetadataYAML{
			Name:      buildName(spec),
			Namespace: ns,
			Labels:    labels,
		},
		Spec: spec,
	}
}

func exitErr(msg string) {
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(1)
}

func main() {
	var (
		inputPath  string
		outputPath string
		namespace  string
		labels     map[string]string
	)

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
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
`)
		pflag.PrintDefaults()
	}
	pflag.StringVarP(&inputPath, "input", "i", "", "Path or URL to Casbin policy CSV (file or http/https)")
	pflag.StringVarP(&outputPath, "output", "o", "-", "Output file for generated YAML. Use '-' for stdout.")
	pflag.StringVarP(&namespace, "namespace", "n", "", "Target namespace for generated Rules (optional)")
	pflag.StringToStringVar(&labels, "label", nil, "Label to add to metadata.labels (repeatable: --label key=value)")
	pflag.Parse()

	if inputPath == "" {
		pflag.Usage()
		exitErr("missing --input/-i (must be file or URL)")
	}
	if inputPath == "-" {
		exitErr("stdin input is not supported, please provide a file path or HTTP(S) URL")
	}

	policyContent, err := readPolicyContent(inputPath)
	if err != nil {
		exitErr(err.Error())
	}

	specs, err := parsePolicyContent(policyContent)
	if err != nil {
		exitErr(err.Error())
	}

	var output io.Writer
	if outputPath == "-" {
		output = os.Stdout
	} else {
		f, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o664)
		if err != nil {
			exitErr(fmt.Sprintf("open output file: %v", err))
		}
		defer func() {
			if err := f.Close(); err != nil {
				exitErr(fmt.Sprintf("close output file: %v", err))
			}
		}()
		output = f
	}

	for _, spec := range specs {
		obj := buildRuleYAML(spec, namespace, labels)
		data, err := yaml.Marshal(&obj)
		if err != nil {
			exitErr(fmt.Sprintf("yaml marshal: %v", err))
		}
		if _, err := fmt.Fprintf(output, "---\n%s", data); err != nil {
			exitErr(fmt.Sprintf("write output: %v", err))
		}
	}
}
