package common

const (
	DEFAULT_REGISTRY = "index.docker.io"
	DEFAULT_TAG      = "latest"
)

type RegistryKind string

const (
	Generic RegistryKind = "generic"
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

func (r *RegistryOptions) Kind() RegistryKind {
	return r.kind
}

func (r *RegistryOptions) Project() string {
	return r.project
}

func (r *RegistryOptions) Strict(strict bool) {
	r.strict = strict
}

func (r *RegistryOptions) Insecure() bool {
	return r.insecure
}

func (r *RegistryOptions) SkipTLSVerify() bool {
	return r.skipTLSVerify
}
