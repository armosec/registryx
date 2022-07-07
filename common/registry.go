package common

type RegistryOptions struct {
	strict          bool // weak by default
	insecure        bool // secure by default
	defaultRegistry string
	defaultTag      string
	project         string
}

func (r *RegistryOptions) Project() string {
	return r.project
}

func (r *RegistryOptions) Strict(strict bool) {
	r.strict = strict
}
