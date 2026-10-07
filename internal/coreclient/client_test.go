package coreclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListModules(t *testing.T) {
	description := "Official shell"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/modules" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("unexpected Accept header: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"manafield-web","name":"Manafield Web","description":"Official shell","version":"0.0.1"}]`))
	}))
	defer server.Close()

	client, err := New(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}

	modules, err := client.ListModules(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(modules))
	}
	if modules[0].ID != "manafield-web" || modules[0].Name != "Manafield Web" {
		t.Fatalf("unexpected module: %#v", modules[0])
	}
	if modules[0].Description == nil || *modules[0].Description != description {
		t.Fatalf("unexpected description: %#v", modules[0].Description)
	}
}

func TestListModulesRejectsErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client, err := New(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.ListModules(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
