package common

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

const (
	DEFAULT_REGISTRY = "index.docker.io"
	DEFAULT_TAG      = "latest"
)

type RegistryKind string

const (
	Generic RegistryKind = ""
	Harbor  RegistryKind = "harbor"
	Quay    RegistryKind = "quay.io"
)

type RegistryOptions struct {
	strict          bool         // weak by default
	insecure        bool         // secure by default
	defaultRegistry string       // index.docker.io
	defaultTag      string       // latest
	project         string       // empty
	skipTLSVerify   bool         // default: do not skip
	kind            RegistryKind //registry provider (e.g. harbor) default is "Generic"
}

func GetRegistryKind(kindStr string) (RegistryKind, error) {
	switch RegistryKind(strings.ToLower(kindStr)) {
	case Harbor:
		return Harbor, nil
	case Quay:
		return Quay, nil
	case Generic:
		return Generic, nil
	default:
		return Generic, fmt.Errorf("unsupported registry kind %s, defaulting to generic kind ", kindStr)

	}
}

func MakeRegistryOptions(isStrict, isInsecure, skipTLSVerify bool, defaultRegistry, defaultTag, project string, kind RegistryKind) *RegistryOptions {
	if defaultRegistry == "" {
		defaultRegistry = DEFAULT_REGISTRY
	}

	if defaultTag == "" {
		defaultTag = DEFAULT_TAG
	}

	if kind == "" {
		kind = Generic
	}
	return &RegistryOptions{strict: isStrict, insecure: isInsecure, defaultRegistry: defaultRegistry, defaultTag: defaultTag, project: project, skipTLSVerify: skipTLSVerify, kind: kind}
}

// MakeRepoWithRegistry builds a name.Repository by prepending the target registry
// before parsing, preserving repo paths that contain registry-like prefixes
// (e.g. "docker.io/prom/blackbox-exporter" stored inside an ACR registry).
func MakeRepoWithRegistry(repoName string, registry *name.Registry) (*name.Repository, error) {
	refName := repoName
	if registry != nil {
		if parsedRepo, err := name.NewRepository(repoName); err == nil && parsedRepo.RegistryStr() == registry.RegistryStr() {
			refName = repoName
		} else {
			refName = registry.RegistryStr() + "/" + repoName
		}
	}

	repo, err := name.NewRepository(refName)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (r *RegistryOptions) Kind() RegistryKind {
	return r.kind
}

func (r *RegistryOptions) DefaultRegistry() string {
	return r.defaultRegistry
}

func (r *RegistryOptions) DefaultTag() string {
	return r.defaultTag
}

func (r *RegistryOptions) Project() string {
	return r.project
}

func (r *RegistryOptions) Strict() bool {
	return r.strict
}

func (r *RegistryOptions) Insecure() bool {
	return r.insecure
}

func (r *RegistryOptions) SkipTLSVerify() bool {
	return r.skipTLSVerify
}

func (r *RegistryOptions) WithInsecure(insecure bool) *RegistryOptions {
	r.insecure = insecure
	return r
}

func (r *RegistryOptions) WithSkipTLSVerify(skipTLSVerify bool) *RegistryOptions {
	r.skipTLSVerify = skipTLSVerify
	return r
}
