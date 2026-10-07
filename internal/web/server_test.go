package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eventide-manafield/manafield-web/internal/coreclient"
)

func TestHealth(t *testing.T) {
	coreServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("[]"))
	}))
	defer coreServer.Close()

	client, err := coreclient.New(coreServer.URL, coreServer.Client())
	if err != nil {
		t.Fatal(err)
	}
	handler, err := New(client, "test-version", "/")
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/manafield/health", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected body: %s", response.Body.String())
	}
}

func TestHomeListsRegistryModules(t *testing.T) {
	coreServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"id":"echo","name":"Echo","description":"Transient conversations","version":"0.1.0"},
			{"id":"account","name":"Account","description":null,"version":"1.2.3"}
		]`))
	}))
	defer coreServer.Close()

	client, err := coreclient.New(coreServer.URL, coreServer.Client())
	if err != nil {
		t.Fatal(err)
	}
	handler, err := New(client, "test-version", "/")
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	response, err := server.Client().Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)
	html := string(body)
	for _, want := range []string{"Echo", "Account", "0.1.0", "1.2.3", "Registry connected"} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in body", want)
		}
	}
}

func TestHomeDegradesWhenCoreUnavailable(t *testing.T) {
	coreServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer coreServer.Close()

	client, err := coreclient.New(coreServer.URL, coreServer.Client())
	if err != nil {
		t.Fatal(err)
	}
	handler, err := New(client, "test-version", "/")
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected degraded home to stay 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "Registry unavailable") {
		t.Fatalf("expected degraded state in body: %s", response.Body.String())
	}
}

func TestPrefixBasePath(t *testing.T) {
	coreServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer coreServer.Close()

	client, err := coreclient.New(coreServer.URL, coreServer.Client())
	if err != nil {
		t.Fatal(err)
	}

	handler, err := New(client, "test-version", "/_manafield")
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/_manafield/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	body := response.Body.String()
	for _, want := range []string{
		`href="/_manafield/static/app.css"`,
		`href="/_manafield/"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected %q in body", want)
		}
	}

	healthRequest := httptest.NewRequest(http.MethodGet, "/_manafield/manafield/health", nil)
	healthResponse := httptest.NewRecorder()
	handler.ServeHTTP(healthResponse, healthRequest)

	if healthResponse.Code != http.StatusOK {
		t.Fatalf("expected prefixed health 200, got %d", healthResponse.Code)
	}
}
