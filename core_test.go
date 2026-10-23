package mwclientmini

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("método = %s; expected GET", r.Method)
		}

		q := r.URL.Query()

		if q.Get("format") != "json" {
			t.Errorf("format = %q; expected json", q.Get("format"))
		}

		if q.Get("formatversion") != "2" {
			t.Errorf("formatversion = %q; expected 2", q.Get("formatversion"))
		}

		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := New("https://test.wikipedia.org/w/api.php", server.URL, "test-bot", "test-password")

	params := url.Values{}
	params.Set("action", "query")

	resp, err := client.Get(params)
	if err != nil {
		t.Fatalf("Get() returned an error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d; expected 200", resp.StatusCode)
	}
}
