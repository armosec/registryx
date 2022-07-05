package common

type PaginationOption struct {
	Cursor string
	Size   int
}

func NoPagination(sizelimit int) PaginationOption {
	return PaginationOption{Size: sizelimit}
}
