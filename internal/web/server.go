package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"sort"
	"time"

	"github.com/eventide-manafield/manafield-web/internal/coreclient"
	webassets "github.com/eventide-manafield/manafield-web/web"
)

type Server struct {
	core     *coreclient.Client
	version  string
	template *template.Template
	static   http.Handler
}

type pageData struct {
	Version       string
	CoreAvailable bool
	Modules       []coreclient.Module
	ModuleCount   int
}

func New(core *coreclient.Client, version string) (http.Handler, error) {
	if core == nil {
		return nil, fmt.Errorf("Core client is required")
	}

	tmpl, err := template.ParseFS(webassets.FS, "templates/index.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	staticFS, err := fs.Sub(webassets.FS, "static")
	if err != nil {
		return nil, fmt.Errorf("open static assets: %w", err)
	}

	server := &Server{
		core:     core,
		version:  version,
		template: tmpl,
		static:   http.FileServer(http.FS(staticFS)),
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", server.static))
	mux.HandleFunc("/manafield/health", server.health)
	mux.HandleFunc("/", server.home)
	return mux, nil
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "manafield-web",
		"version": s.version,
	})
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	modules, err := s.core.ListModules(ctx)
	coreAvailable := err == nil
	if err != nil {
		slog.Warn("Core registry unavailable", "error", err)
		modules = nil
	}

	sort.Slice(modules, func(i, j int) bool {
		if modules[i].Name == modules[j].Name {
			return modules[i].ID < modules[j].ID
		}
		return modules[i].Name < modules[j].Name
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.template.ExecuteTemplate(w, "index.html", pageData{
		Version:       s.version,
		CoreAvailable: coreAvailable,
		Modules:       modules,
		ModuleCount:   len(modules),
	}); err != nil {
		slog.Error("render home", "error", err)
	}
}
