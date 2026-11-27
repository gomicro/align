package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

func TestCloneAccountsPagination(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.URL.RequestURI())
		mu.Unlock()

		var body string
		switch {
		case r.URL.Path == "/user":
			body = `{"login":"SomeUser"}`

		case r.URL.Path == "/user/orgs" && r.URL.Query().Get("page") == "":
			w.Header().Set("Link", fmt.Sprintf(`<http://%s/user/orgs?page=2&per_page=100>; rel="next"`, r.Host))
			body = `[{"login":"FirstOrg","name":"First Organization","description":"First organization"}]`

		case r.URL.Path == "/user/orgs" && r.URL.Query().Get("page") == "2":
			body = `[{"login":"SecondOrg","description":null}]`

		default:
			t.Errorf("unexpected GitHub request: %s", r.URL)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// The byte count is not relevant to the response fixture.
		_, err := fmt.Fprint(w, body)
		if err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	defer server.Close()

	baseURL, err := url.Parse(server.URL + "/")
	if !assert.NoError(t, err) {
		return
	}
	ghClient := github.NewClient(server.Client())
	ghClient.BaseURL = baseURL
	clt := &Client{ghClient: ghClient}

	accounts, err := clt.CloneAccounts(context.Background())
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []CloneAccount{
		{Login: "someuser"},
		{Login: "firstorg", Organization: true, Name: "First Organization", Description: "First organization"},
		{Login: "secondorg", Organization: true},
	}, accounts)

	logins, err := clt.GetLogins(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{"someuser", "firstorg", "secondorg"}, logins)

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, requests, 6)
	for _, request := range requests {
		assert.NotContains(t, request, "/repos")
		assert.NotContains(t, request, "/orgs/")
	}
}

func TestCloneAccountsFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		failPath string
	}{
		{name: "user", failPath: "/user"},
		{name: "organizations", failPath: "/user/orgs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == tt.failPath {
					http.Error(w, "unavailable", http.StatusServiceUnavailable)
					return
				}

				// The byte count is not relevant to the response fixture.
				_, err := fmt.Fprint(w, `{"login":"user"}`)
				if err != nil {
					t.Errorf("writing response: %v", err)
				}
			}))
			defer server.Close()

			baseURL, err := url.Parse(server.URL + "/")
			if !assert.NoError(t, err) {
				return
			}
			ghClient := github.NewClient(server.Client())
			ghClient.BaseURL = baseURL
			clt := &Client{ghClient: ghClient}

			accounts, err := clt.CloneAccounts(context.Background())
			assert.Error(t, err)
			assert.Empty(t, accounts)
		})
	}
}

func TestCloneAccountsLaterPageFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user/orgs" && r.URL.Query().Get("page") == "2" {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}

		body := `{"login":"user"}`
		if r.URL.Path == "/user/orgs" {
			w.Header().Set("Link", fmt.Sprintf(`<http://%s/user/orgs?page=2>; rel="next"`, r.Host))
			body = `[{"login":"first"}]`
		}

		// The byte count is not relevant to the response fixture.
		_, err := fmt.Fprint(w, body)
		if err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	defer server.Close()

	baseURL, err := url.Parse(server.URL + "/")
	if !assert.NoError(t, err) {
		return
	}
	ghClient := github.NewClient(server.Client())
	ghClient.BaseURL = baseURL
	clt := &Client{ghClient: ghClient}
	accounts, err := clt.CloneAccounts(context.Background())
	assert.Error(t, err)
	assert.Empty(t, accounts)
}
