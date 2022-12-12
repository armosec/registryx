package harbor

import (
	"context"
	_ "embed"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/armosec/registryx/common"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/assert"
)

//go:embed fixtures/repoResponse.json
var repoResponseBytes []byte

//go:embed fixtures/tagsResponse.json
var tagsResponseBytes []byte

func TestCatalogAndList(t *testing.T) {
	registry, err := name.NewRegistry("localhost:9111")
	if err != nil {
		t.Errorf("err1: %v", err.Error())
	}
	//prepare mock registry server for repositories request
	testServer, err := startTestClientServer("127.0.0.1:9111", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2.0/repositories", r.URL.String(), "request path does not match")
		assert.Equal(t, "Basic YWRtaW46SGFyYm9yMTIzNDU=", r.Header["Authorization"][0])
		w.Write(repoResponseBytes)
	}))
	if err != nil {
		t.Error(err)
	}

	//create harbor registry
	regOptions := common.MakeRegistryOptions(true, false, true, "", "", "", common.Harbor)
	harbor, err := NewHarborRegistry(&authn.AuthConfig{Username: "admin", Password: "Harbor12345"}, &registry, regOptions)
	assert.Nil(t, err)
	ctx := context.Background()
	//test catalog
	repos, nextPage, statusCode, err := harbor.Catalog(ctx, common.NoPaginationOption(), common.CatalogOption{}, nil)
	assert.Nil(t, nextPage)
	assert.Equal(t, 200, statusCode)
	assert.Nil(t, err)
	assert.Equal(t, []string{"my-project/ca-ws", "user2private/kibana", "user-project/kibana", "my-project/kibana", "my-project/postgres"}, repos)
	testServer.Close()

	//prepare mock registry server for list tags request
	testServer, err = startTestClientServer("127.0.0.1:9111", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/my-project/ca-ws/tags/list", r.URL.String(), "request path does not match")
		assert.Equal(t, "Basic YWRtaW46SGFyYm9yMTIzNDU=", r.Header["Authorization"][0])
		w.Write(tagsResponseBytes)
	}))
	if err != nil {
		t.Error(err)
	}

	//test list tags
	tags, nextPage, err := harbor.List(repos[0], common.NoPaginationOption())
	assert.Nil(t, err)
	assert.Nil(t, nextPage)
	assert.Equal(t, []string{"latest", "v0.26", "v0.27", "v0.28", "v22"}, tags)
	testServer.Close()
}

//go:embed fixtures/projectRepos.json
var projectReposResponseBytes []byte

//go:embed fixtures/tagsPaginatedResponse.json
var tagsPaginatedResponseBytes []byte

func TestCatalogAndListWithProjectAndPagination(t *testing.T) {
	registry, err := name.NewRegistry("localhost:9111")
	if err != nil {
		t.Errorf("err1: %v", err.Error())
	}
	//create harbor registry with project
	regOptions := common.MakeRegistryOptions(true, false, true, "", "", "my-project", common.Harbor)
	harbor, err := NewHarborRegistry(&authn.AuthConfig{Username: "admin", Password: "Harbor12345"}, &registry, regOptions)
	assert.Nil(t, err)
	ctx := context.Background()
	//prepare mock registry server for repositories request
	testServer, err := startTestClientServer("127.0.0.1:9111", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2.0/projects/my-project/repositories?page=1&page_size=2", r.URL.String(), "request path does not match")
		assert.Equal(t, "Basic YWRtaW46SGFyYm9yMTIzNDU=", r.Header["Authorization"][0])
		w.Header().Add("Link", "</api/v2.0/projects/my-project/repositories?page=2&page_size=2>; rel=\"next\"")
		w.Write(projectReposResponseBytes)
	}))
	if err != nil {
		t.Error(err)
	}
	//test catalog
	repos, nextReposPage, statusCode, err := harbor.Catalog(ctx, common.MakePagination(2), common.CatalogOption{}, nil)
	assert.Nil(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, []string{"my-project/ca-ws", "my-project/kibana"}, repos)
	assert.Equal(t, &common.PaginationOption{Cursor: "2", Size: 2}, nextReposPage)
	testServer.Close()

	//test catalog with with next page
	testServer, err = startTestClientServer("127.0.0.1:9111", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2.0/projects/my-project/repositories?page=2&page_size=1", r.URL.String(), "request path does not match")
		assert.Equal(t, "Basic YWRtaW46SGFyYm9yMTIzNDU=", r.Header["Authorization"][0])
		w.Header().Add("Link", "</api/v2.0/projects/my-project/repositories?page=1&page_size=1>; rel=\"prev\" , </api/v2.0/projects/my-project/repositories?page=3&page_size=1>; rel=\"next\"")
		w.Write(projectReposResponseBytes)
	}))
	if err != nil {
		t.Error(err)
	}

	_, nextReposPage, statusCode, err = harbor.Catalog(ctx, common.PaginationOption{Cursor: "2", Size: 1}, common.CatalogOption{}, nil)
	assert.Nil(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, &common.PaginationOption{Cursor: "3", Size: 1}, nextReposPage)
	testServer.Close()

	//test catalog with with last page
	testServer, err = startTestClientServer("127.0.0.1:9111", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2.0/projects/my-project/repositories?page=2&page_size=2", r.URL.String(), "request path does not match")
		assert.Equal(t, "Basic YWRtaW46SGFyYm9yMTIzNDU=", r.Header["Authorization"][0])
		w.Header().Add("Link", "</api/v2.0/projects/my-project/repositories?page=1&page_size=2>; rel=\"prev\"")
		w.Write(projectReposResponseBytes)
	}))
	if err != nil {
		t.Error(err)
	}
	_, nextReposPage, _, err = harbor.Catalog(ctx, common.PaginationOption{Cursor: "2", Size: 2}, common.CatalogOption{}, nil)
	assert.Nil(t, err)
	assert.Nil(t, nextReposPage)
	testServer.Close()

	//prepare mock registry server for list tags request
	testServer, err = startTestClientServer("127.0.0.1:9111", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/my-project/ca-ws/tags/list?n=2", r.URL.String(), "request path does not match")
		assert.Equal(t, "Basic YWRtaW46SGFyYm9yMTIzNDU=", r.Header["Authorization"][0])
		w.Header().Add("Link", "</v2/my-project/ca-ws/tags/list?last=v0.26&n=2>; rel=\"next\"")

		w.Write(tagsPaginatedResponseBytes)
	}))
	if err != nil {
		t.Error(err)
	}

	//test list tags
	tags, nextTagsPage, err := harbor.List(repos[0], common.MakePagination(2))
	assert.Nil(t, err)
	assert.Equal(t, &common.PaginationOption{Cursor: "v0.26", Size: 2}, nextTagsPage)
	assert.Equal(t, []string{"latest", "v0.26"}, tags)
	testServer.Close()

	//test next tags page
	testServer, err = startTestClientServer("127.0.0.1:9111", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/my-project/ca-ws/tags/list?last=v0.26&n=2", r.URL.String(), "request path does not match")
		assert.Equal(t, "Basic YWRtaW46SGFyYm9yMTIzNDU=", r.Header["Authorization"][0])
		w.Header().Add("Link", "</v2/my-project/ca-ws/tags/list?last=v0.28&n=2>; rel=\"next\"")
		w.Write(tagsPaginatedResponseBytes)
	}))
	if err != nil {
		t.Error(err)
	}
	_, nextTagsPage, err = harbor.List(repos[0], *nextTagsPage)
	assert.Nil(t, err)
	assert.Equal(t, &common.PaginationOption{Cursor: "v0.28", Size: 2}, nextTagsPage)
	testServer.Close()

	//test last tags page
	testServer, err = startTestClientServer("127.0.0.1:9111", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/my-project/ca-ws/tags/list?last=v0.28&n=2", r.URL.String(), "request path does not match")
		assert.Equal(t, "Basic YWRtaW46SGFyYm9yMTIzNDU=", r.Header["Authorization"][0])
		w.Write(tagsPaginatedResponseBytes)
	}))
	if err != nil {
		t.Error(err)
	}
	_, nextTagsPage, err = harbor.List(repos[0], *nextTagsPage)
	assert.Nil(t, err)
	assert.Nil(t, nextTagsPage)
	testServer.Close()

}

func TestInsecureRequest(t *testing.T) {
	//create harbor secure registry
	registry, err := name.NewRegistry("localhost:9111")
	if err != nil {
		t.Errorf("err1: %v", err.Error())
	}
	regOptions := common.MakeRegistryOptions(true, false, true, "", "", "", common.Harbor)
	iHarbor, err := NewHarborRegistry(&authn.AuthConfig{Username: "admin", Password: "Harbor12345"}, &registry, regOptions)
	assert.Nil(t, err)
	harbor := iHarbor.(*HarborRegistry)
	//test repo request scheme
	req, err := harbor.repositoriesRequest("0", "")
	assert.Nil(t, err)
	assert.Equal(t, "https", req.URL.Scheme)
	//test list tags request scheme
	repoData, err := name.NewRepository("repo")
	assert.Nil(t, err)
	repoData.Registry = registry
	req, err = harbor.listTagsRequest(repoData, "0", "")
	assert.Nil(t, err)
	assert.Equal(t, "https", req.URL.Scheme)

	//create harbor insecure registry
	registry, err = name.NewRegistry("localhost:9111", name.Insecure)
	if err != nil {
		t.Errorf("err1: %v", err.Error())
	}
	regOptions = common.MakeRegistryOptions(true, true, true, "", "", "", common.Harbor)
	iHarbor, err = NewHarborRegistry(&authn.AuthConfig{Username: "admin", Password: "Harbor12345"}, &registry, regOptions)
	assert.Nil(t, err)
	harbor = iHarbor.(*HarborRegistry)
	//test repo request scheme
	req, err = harbor.repositoriesRequest("0", "")
	assert.Nil(t, err)
	assert.Equal(t, "http", req.URL.Scheme)
	//test list tags request scheme
	repoData, err = name.NewRepository("repo")
	assert.Nil(t, err)
	repoData.Registry = registry
	req, err = harbor.listTagsRequest(repoData, "0", "")
	assert.Nil(t, err)
	assert.Equal(t, "http", req.URL.Scheme)

}

func startTestClientServer(requestUrl string, handler http.Handler) (*httptest.Server, error) {
	l, err := net.Listen("tcp", requestUrl)
	if err != nil {
		return nil, err
	}
	testServer := httptest.NewUnstartedServer(handler)
	testServer.Listener.Close()
	testServer.Listener = l
	testServer.StartTLS()
	return testServer, nil
}
