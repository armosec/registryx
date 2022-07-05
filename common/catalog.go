package common

type CatalogOption struct {
	IncludeLastModified bool
	IsPublic            bool
	Namespaces          string //scope for e.g cataloging just "armosec" images
}
