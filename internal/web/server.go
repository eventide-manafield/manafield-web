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
	"strings"
	"time"

	"github.com/eventide-manafield/manafield-web/internal/coreclient"
	webassets "github.com/eventide-manafield/manafield-web/web"
)

type Server struct {
	core     *coreclient.Client
	version  string
	basePath string
	template *template.Template
	static   http.Handler
}

type pageData struct {
	Version       string
	HomePath      string
	StaticPath    string
	CoreAvailable bool
	Modules       []coreclient.Module
	ModuleCount   int
}

func New(core *coreclient.Client, version, basePath string) (http.Handler, error) {
	if core == nil {
		return nil, fmt.Errorf("Core client is required")
	}

	basePath, err := normalizeBasePath(basePath)
	if err != nil {
		return nil, err
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
		basePath: basePath,
		template: tmpl,
		static:   http.FileServer(http.FS(staticFS)),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/manafield/health", server.health)

	if basePath == "/" {
		mux.Handle("/static/", http.StripPrefix("/static/", server.static))
		mux.HandleFunc("/", server.home)
	} else {
		staticPrefix := basePath + "/static/"
		mux.Handle(staticPrefix, http.StripPrefix(staticPrefix, server.static))
		mux.HandleFunc(basePath+"/manafield/health", server.health)
		mux.HandleFunc(basePath, server.home)
		mux.HandleFunc(basePath+"/", server.home)
	}

	return mux, nil
}

func normalizeBasePath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "/" {
		return "/", nil
	}

	if !strings.HasPrefix(value, "/") {
		return "", fmt.Errorf("Web base path must start with '/'")
	}

	value = strings.TrimRight(value, "/")
	if strings.Contains(value, "//") || strings.ContainsAny(value, "?#") {
		return "", fmt.Errorf("Web base path must be a canonical URL path prefix")
	}

	return value, nil
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
	if !s.isHomePath(r.URL.Path) {
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

	homePath := "/"
	staticPath := "/static"
	if s.basePath != "/" {
		homePath = s.basePath + "/"
		staticPath = s.basePath + "/static"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.template.ExecuteTemplate(w, "index.html", pageData{
		Version:       s.version,
		HomePath:      homePath,
		StaticPath:    staticPath,
		CoreAvailable: coreAvailable,
		Modules:       modules,
		ModuleCount:   len(modules),
	}); err != nil {
		slog.Error("render home", "error", err)
	}
}

func (s *Server) isHomePath(path string) bool {
	if s.basePath == "/" {
		return path == "/"
	}

	return path == s.basePath || path == s.basePath+"/"
}
