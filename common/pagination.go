package common

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type PaginationOption struct {
	Cursor string
	Size   int
}

func MakePagination(size int) PaginationOption {
	return PaginationOption{Size: size}
}

func NoPaginationOption() PaginationOption {
	return MakePagination(0)
}

//TODO - get pagination info for request header instead of guessing next page existence by size
func CalcNextV2Pagination(repos []string, size int) *PaginationOption {
	//assume that if response repos use all the allowed size then probably there is another page
	if len(repos) < size {
		return nil
	}
	return &PaginationOption{Cursor: repos[len(repos)-1], Size: size}
}

//GetNextV2Pagination from response header
func GetNextV2Pagination(resp *http.Response) (*PaginationOption, error) {
	link := resp.Header.Get("Link")
	if link == "" {
		return nil, nil
	}

	if link[0] != '<' {
		return nil, fmt.Errorf("failed to parse link header: missing '<' in: %s", link)
	}

	end := strings.Index(link, ">")
	if end == -1 {
		return nil, fmt.Errorf("failed to parse link header: missing '>' in: %s", link)
	}
	link = link[1:end]

	linkURL, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	queryParams := linkURL.Query()
	nextPagination := PaginationOption{}
	if sizeValues, ok := queryParams["n"]; !ok {
		return nil, fmt.Errorf("page size is missing in next page header")
	} else {
		if pageSize, err := strconv.Atoi(sizeValues[0]); err != nil {
			return nil, fmt.Errorf("page size is not an integer in next page header")
		} else {
			nextPagination.Size = pageSize
		}
	}
	if cursorValues, ok := queryParams["last"]; !ok {
		return nil, fmt.Errorf("last cursor is missing in next page header")
	} else {
		nextPagination.Cursor = cursorValues[0]
	}
	return &nextPagination, nil
}
