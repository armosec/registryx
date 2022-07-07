package quay

// put here unique datastructures related to quay.io
type QuayRepository struct {
	Namespace    string `json:"namespace,omitempty"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	IsPublic     bool   `json:"is_public,omitempty"`
	Kind         string `json:"kind,omitempty"`
	State        string `json:"state,omitempty"`
	LastModified int    `json:"last_modified,omitempty"`
}

type QuayCatalogResponse struct {
	Repositories []QuayRepository `json:"repositories"`
	Cursor       string           `json:"next_page,omitempty"`
}
