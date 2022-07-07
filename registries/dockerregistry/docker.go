package dockerregistry

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

type DockerTokenResponse struct {
	Token   string    `json:"token"`
	Expires int       `json:"expires_in"`
	Issued  time.Time `json:"issued_at"`
}

func Token(auth *authn.AuthConfig, regisry *name.Registry) (*DockerTokenResponse, error) {
	client := http.Client{}
	uri := &url.URL{
		Scheme: regisry.Scheme(),
		Host:   "auth.docker.io",
		Path:   "/token",
	}
	q := uri.Query()
	q.Add("service", "registry.docker.io")
	// scope registry
	uri.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	token := &DockerTokenResponse{}
	err = json.Unmarshal(body, token)
	return token, err
}
