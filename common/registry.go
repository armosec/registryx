package common

const (
	DEFAULT_REGISTRY = "index.docker.io"
	DEFAULT_TAG      = "latest"
)

type RegistryOptions struct {
	strict          bool   // weak by default
	insecure        bool   // secure by default
	defaultRegistry string // index.docker.io
	defaultTag      string // latest
	project         string // empty
}

func MakeRegistryOptions(isStrict, isInsecure bool, defaultRegistry, defaultTag, project string) *RegistryOptions {
	if defaultRegistry == "" {
		defaultRegistry = DEFAULT_REGISTRY
	}

	if defaultTag == "" {
		defaultTag = DEFAULT_TAG
	}
	return &RegistryOptions{strict: isStrict, insecure: isInsecure, defaultRegistry: defaultRegistry, defaultTag: defaultTag, project: project}
}

func (r *RegistryOptions) Project() string {
	return r.project
}

func (r *RegistryOptions) Strict(strict bool) {
	r.strict = strict
}
