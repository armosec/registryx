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

func MakeRepoWithRegistry(repoName string, registry *name.Registry) (*name.Repository, error) {
	repo, err := name.NewRepository(repoName)
	if err != nil {
		return nil, err
	}
	if registry != nil {
		repo.Registry = *registry
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
