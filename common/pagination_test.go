package common

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcNextV2Pagination(t *testing.T) {
	repos := []string{"first", "second", "last"}
	nextPage := CalcNextV2Pagination(repos, 3)
	assert.Equal(t, nextPage.Cursor, "last")
	assert.Equal(t, nextPage.Size, 3)

	nextPage = CalcNextV2Pagination(repos, 2)
	assert.Equal(t, nextPage.Cursor, "last")
	assert.Equal(t, nextPage.Size, 2)

	nextPage = CalcNextV2Pagination(repos, 4)
	assert.Nil(t, nextPage)

}

func TestGetNextV2Pagination(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	nextPage, err := GetNextV2Pagination(resp)
	assert.Nil(t, nextPage)
	assert.Nil(t, err)

	resp.Header.Set("Link", `</v2/_catalog?last=user-project%2Fkibana&n=3>; rel="next"`)
	nextPage, err = GetNextV2Pagination(resp)
	assert.Nil(t, err)
	assert.Equal(t, nextPage.Cursor, "user-project/kibana")
	assert.Equal(t, nextPage.Size, 3)

	resp.Header.Set("Link", `/v2/_catalog?last=user-project%2Fkibana&n=3>; rel="next"`)
	nextPage, err = GetNextV2Pagination(resp)
	assert.Nil(t, nextPage)
	assert.EqualError(t, err, "failed to parse link header: missing '<' in: /v2/_catalog?last=user-project%2Fkibana&n=3>; rel=\"next\"")

	resp.Header.Set("Link", `</v2/_catalog?last=user-project%2Fkibana&n=3; rel="next"`)
	nextPage, err = GetNextV2Pagination(resp)
	assert.Nil(t, nextPage)
	assert.EqualError(t, err, "failed to parse link header: missing '>' in: </v2/_catalog?last=user-project%2Fkibana&n=3; rel=\"next\"")

	resp.Header.Set("Link", `</v2/_catalog?last=user-project%2Fkibana&NOPAGE=3>; rel="next"`)
	nextPage, err = GetNextV2Pagination(resp)
	assert.Nil(t, nextPage)
	assert.EqualError(t, err, "page size is missing in next page header")

	resp.Header.Set("Link", `</v2/_catalog?last=user-project%2Fkibana&n=string>; rel="next"`)
	nextPage, err = GetNextV2Pagination(resp)
	assert.Nil(t, nextPage)
	assert.EqualError(t, err, "page size is not an integer in next page header")

	resp.Header.Set("Link", `</v2/_catalog?nolast=user-project%2Fkibana&n=3>; rel="next"`)
	nextPage, err = GetNextV2Pagination(resp)
	assert.Nil(t, nextPage)
	assert.EqualError(t, err, "last cursor is missing in next page header")

}
