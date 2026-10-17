// Package catlhttp provides a mountable HTTP daemon for serving catalog queries
// and atomic reloads following strict RESTful conventions and RFC 9457 Problem Details.
package catlhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ju4n97/catl"
)

// MediaTypeProblemJSON is the RFC 9457 media type for problem details payloads.
const MediaTypeProblemJSON = "application/problem+json"

// Route documents a registered HTTP route for self-introspection and CLI help formatting.
type Route struct {
	Method      string
	Path        string
	Description string
}

// Routes returns the canonical list of RESTful routes supported across single and registry modes.
func Routes() []Route {
	return []Route{
		{Method: "GET", Path: "/v1/catalog", Description: "Get catalog identity, namespace, and tree hash"},
		{Method: "POST", Path: "/v1/catalog/reload", Description: "Atomically reload catalog from disk"},
		{Method: "GET", Path: "/v1/catalog/entities", Description: "List all entities across all collections"},
		{Method: "GET", Path: "/v1/collections", Description: "List all collections with entity counts"},
		{Method: "GET", Path: "/v1/collections/{collection}", Description: "Get collection details"},
		{Method: "GET", Path: "/v1/collections/{collection}/entities", Description: "List entities in a collection"},
		{
			Method:      "GET",
			Path:        "/v1/collections/{collection}/entities/{id}",
			Description: "Get entity resource (metadata and parsed manifest)",
		},
		{
			Method:      "GET",
			Path:        "/v1/collections/{collection}/entities/{id}/manifest",
			Description: "Stream raw manifest file as stored on disk",
		},
		{Method: "GET", Path: "/v1/collections/{collection}/entities/{id}/files", Description: "List companion files in bundle entity"},
		{Method: "GET", Path: "/v1/collections/{collection}/entities/{id}/files/{path...}", Description: "Stream bundle companion file"},
		{Method: "GET", Path: "/v1/catalogs", Description: "List all tenant catalogs (registry mode)"},
		{Method: "POST", Path: "/v1/catalogs/reload", Description: "Atomically reload all catalogs (registry mode)"},
		{Method: "ANY", Path: "/v1/catalogs/{catalog}/*", Description: "Scoped access to tenant catalog routes"},
	}
}

// EndpointsSummary formats the registered routes into an aligned plain text table for CLI help.
func EndpointsSummary() string {
	var sb strings.Builder
	for _, r := range Routes() {
		fmt.Fprintf(&sb, "  %-6s %-64s # %s\n", r.Method, r.Path, r.Description)
	}
	return strings.TrimRight(sb.String(), "\n")
}

// Problem represents an RFC 9457 Problem Details for HTTP APIs payload.
type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance,omitzero"`
}

// CatalogSummary represents the top-level metadata of a catalog.
type CatalogSummary struct {
	Name        string   `json:"name"`
	TreeHash    string   `json:"tree_hash"`
	Collections []string `json:"collections"`
	EntityCount int      `json:"entity_count"`
}

// CollectionSummary represents metadata and entity count for a collection.
type CollectionSummary struct {
	Name        string `json:"name"`
	EntityCount int    `json:"entity_count"`
}

// EntityResponse represents the complete, canonical Entity resource payload.
type EntityResponse struct {
	ID             string   `json:"id"`
	Collection     string   `json:"collection"`
	Namespace      string   `json:"namespace"`
	QualifiedID    string   `json:"qualified_id"`
	Path           string   `json:"path"`
	IsBundle       bool     `json:"is_bundle"`
	CompanionFiles []string `json:"companion_files,omitzero"`
	Manifest       any      `json:"manifest"`
}

// Server serves catalog entities over HTTP and implements http.Handler.
type Server struct {
	handler http.Handler
}

// ServeHTTP implements http.Handler, allowing Server to be mounted into any Go HTTP router.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

// Serve starts listening on addr and serves requests until ctx is cancelled.
func (s *Server) Serve(ctx context.Context, addr string) error {
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           s.handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	errCh := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// New constructs an HTTP Server for a single catalog instance.
func New(cat *catl.Catalog) *Server {
	sh := newSingleHandler(cat)
	mux := http.NewServeMux()
	mux.Handle("/v1/", http.StripPrefix("/v1", sh))
	return &Server{handler: mux}
}

// NewRegistry constructs an HTTP Server for a multi-tenant Registry.
func NewRegistry(reg *catl.Registry) *Server {
	rh := newRegistryHandler(reg)
	return &Server{handler: rh}
}

// singleHandler manages routes and state for a single catalog relative to its mount root.
type singleHandler struct {
	catalog atomic.Pointer[catl.Catalog]
	mux     *http.ServeMux
}

// newSingleHandler initializes a singleHandler with routes declared relative to root.
func newSingleHandler(cat *catl.Catalog) *singleHandler {
	sh := &singleHandler{}
	sh.catalog.Store(cat)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalog", sh.handleGetCatalog)
	mux.HandleFunc("POST /catalog/reload", sh.handleReloadCatalog)
	mux.HandleFunc("GET /catalog/entities", sh.handleGetCatalogEntities)
	mux.HandleFunc("GET /collections", sh.handleListCollections)
	mux.HandleFunc("GET /collections/{collection}", sh.handleGetCollection)
	mux.HandleFunc("GET /collections/{collection}/entities", sh.handleListCollectionEntities)
	mux.HandleFunc("GET /collections/{collection}/entities/{id}", sh.handleGetEntity)
	mux.HandleFunc("GET /collections/{collection}/entities/{id}/manifest", sh.handleGetManifest)
	mux.HandleFunc("GET /collections/{collection}/entities/{id}/files", sh.handleListCompanionFiles)
	mux.HandleFunc("GET /collections/{collection}/entities/{id}/files/{path...}", sh.handleGetCompanionFile)

	sh.mux = mux
	return sh
}

// ServeHTTP dispatches requests to the relative single-catalog router.
func (sh *singleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sh.mux.ServeHTTP(w, r)
}

// handleGetCatalog handles GET /catalog by returning catalog metadata and tree hash.
func (sh *singleHandler) handleGetCatalog(w http.ResponseWriter, r *http.Request) {
	cat := sh.catalog.Load()
	treeHash, err := cat.TreeHash()
	if err != nil {
		writeProblem(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, CatalogSummary{
		Name:        cat.Name(),
		TreeHash:    treeHash,
		Collections: cat.Collections(),
		EntityCount: len(cat.Entities()),
	})
}

// handleReloadCatalog handles POST /catalog/reload by re-reading the catalog from disk.
func (sh *singleHandler) handleReloadCatalog(w http.ResponseWriter, r *http.Request) {
	cat := sh.catalog.Load()
	rootPath := cat.RootPath()
	if rootPath == "" {
		writeProblem(w, r, http.StatusBadRequest, "catalog was opened without a backing directory; disk reload unavailable")
		return
	}

	next, err := catl.Open(r.Context(), rootPath)
	if err != nil {
		writeProblem(w, r, http.StatusUnprocessableEntity, fmt.Sprintf("reload failed: %v", err))
		return
	}

	sh.catalog.Store(next)

	treeHash, err := next.TreeHash()
	if err != nil {
		writeProblem(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to compute tree hash: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":    "ok",
		"catalog":   next.Name(),
		"tree_hash": treeHash,
	})
}

// handleGetCatalogEntities handles GET /catalog/entities by returning all entities across all collections.
func (sh *singleHandler) handleGetCatalogEntities(w http.ResponseWriter, r *http.Request) {
	cat := sh.catalog.Load()
	writeJSON(w, http.StatusOK, cat.Entities())
}

// handleListCollections handles GET /collections by returning summary counts for all collections.
func (sh *singleHandler) handleListCollections(w http.ResponseWriter, r *http.Request) {
	cat := sh.catalog.Load()
	colls := cat.Collections()

	res := make([]CollectionSummary, 0, len(colls))
	for _, name := range colls {
		res = append(res, CollectionSummary{
			Name:        name,
			EntityCount: len(cat.Entities(name)),
		})
	}

	writeJSON(w, http.StatusOK, res)
}

// handleGetCollection handles GET /collections/{collection} by returning collection metadata.
func (sh *singleHandler) handleGetCollection(w http.ResponseWriter, r *http.Request) {
	collection := r.PathValue("collection")
	cat := sh.catalog.Load()

	if !slices.Contains(cat.Collections(), collection) {
		writeProblem(w, r, http.StatusNotFound, fmt.Sprintf("collection %q not found", collection))
		return
	}

	writeJSON(w, http.StatusOK, CollectionSummary{
		Name:        collection,
		EntityCount: len(cat.Entities(collection)),
	})
}

// handleListCollectionEntities handles GET /collections/{collection}/entities by listing entities in the collection.
func (sh *singleHandler) handleListCollectionEntities(w http.ResponseWriter, r *http.Request) {
	collection := r.PathValue("collection")
	cat := sh.catalog.Load()

	if !slices.Contains(cat.Collections(), collection) {
		writeProblem(w, r, http.StatusNotFound, fmt.Sprintf("collection %q not found", collection))
		return
	}

	writeJSON(w, http.StatusOK, cat.Entities(collection))
}

// handleGetEntity handles GET /collections/{collection}/entities/{id}.
// Returns the full Entity resource envelope including metadata and parsed manifest content.
func (sh *singleHandler) handleGetEntity(w http.ResponseWriter, r *http.Request) {
	collection := r.PathValue("collection")
	id := r.PathValue("id")
	cat := sh.catalog.Load()

	ent, err := cat.Resolve(collection, id)
	if err != nil {
		if errors.Is(err, catl.ErrNotQualified) {
			writeProblem(
				w,
				r,
				http.StatusBadRequest,
				fmt.Sprintf("%q is not a qualified ID (want namespace.id); did you mean %q?", id, cat.Name()+"."+id),
			)
			return
		}
		writeProblem(w, r, http.StatusNotFound, err.Error())
		return
	}

	var parsedManifest any
	if err := ent.Unmarshal(&parsedManifest); err != nil {
		writeProblem(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to parse manifest: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, EntityResponse{
		ID:             ent.ID,
		Collection:     ent.Collection,
		Namespace:      ent.Namespace,
		QualifiedID:    ent.QualifiedID,
		Path:           ent.ManifestPath,
		IsBundle:       ent.IsBundle(),
		CompanionFiles: ent.CompanionFiles,
		Manifest:       parsedManifest,
	})
}

// handleGetManifest handles GET /collections/{collection}/entities/{id}/manifest.
// Streams the raw manifest file bytes exactly as stored on disk.
func (sh *singleHandler) handleGetManifest(w http.ResponseWriter, r *http.Request) {
	collection := r.PathValue("collection")
	id := r.PathValue("id")
	cat := sh.catalog.Load()

	ent, err := cat.Resolve(collection, id)
	if err != nil {
		if errors.Is(err, catl.ErrNotQualified) {
			writeProblem(
				w,
				r,
				http.StatusBadRequest,
				fmt.Sprintf("%q is not a qualified ID (want namespace.id); did you mean %q?", id, cat.Name()+"."+id),
			)
			return
		}
		writeProblem(w, r, http.StatusNotFound, err.Error())
		return
	}

	if path.Ext(ent.ManifestPath) == ".json" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(ent.RawManifest())
}

// handleListCompanionFiles handles GET /collections/{collection}/entities/{id}/files.
// Returns a JSON array of relative paths for all companion files in a bundle entity.
func (sh *singleHandler) handleListCompanionFiles(w http.ResponseWriter, r *http.Request) {
	collection := r.PathValue("collection")
	id := r.PathValue("id")
	cat := sh.catalog.Load()

	ent, err := cat.Resolve(collection, id)
	if err != nil {
		if errors.Is(err, catl.ErrNotQualified) {
			writeProblem(
				w,
				r,
				http.StatusBadRequest,
				fmt.Sprintf("%q is not a qualified ID (want namespace.id); did you mean %q?", id, cat.Name()+"."+id),
			)
			return
		}
		writeProblem(w, r, http.StatusNotFound, err.Error())
		return
	}

	if !ent.IsBundle() {
		writeProblem(w, r, http.StatusBadRequest, fmt.Sprintf("entity %q is a flat entity and has no companion files", ent.ID))
		return
	}

	writeJSON(w, http.StatusOK, ent.CompanionFiles)
}

// handleGetCompanionFile handles GET /collections/{collection}/entities/{id}/files/{path...}.
// Streams an individual companion file from a bundle entity with appropriate MIME headers.
func (sh *singleHandler) handleGetCompanionFile(w http.ResponseWriter, r *http.Request) {
	collection := r.PathValue("collection")
	id := r.PathValue("id")
	filePath := r.PathValue("path")
	cat := sh.catalog.Load()

	ent, err := cat.Resolve(collection, id)
	if err != nil {
		if errors.Is(err, catl.ErrNotQualified) {
			writeProblem(
				w,
				r,
				http.StatusBadRequest,
				fmt.Sprintf("%q is not a qualified ID (want namespace.id); did you mean %q?", id, cat.Name()+"."+id),
			)
			return
		}
		writeProblem(w, r, http.StatusNotFound, err.Error())
		return
	}

	if !ent.IsBundle() {
		writeProblem(w, r, http.StatusBadRequest, fmt.Sprintf("entity %q is a flat entity and has no companion files", ent.ID))
		return
	}

	bundleFS, err := ent.FS(r.Context())
	if err != nil {
		writeProblem(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	file, err := bundleFS.Open(filePath)
	if err != nil {
		writeProblem(w, r, http.StatusNotFound, fmt.Sprintf("file %q not found in bundle: %v", filePath, err))
		return
	}
	defer file.Close()

	if ctype := mime.TypeByExtension(filepath.Ext(filePath)); ctype != "" {
		w.Header().Set("Content-Type", ctype)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, file)
}

// registryHandler dispatches requests across multiple tenant catalogs in a Registry.
type registryHandler struct {
	reg     *catl.Registry
	servers atomic.Pointer[map[string]*singleHandler]
	mux     *http.ServeMux
}

// newRegistryHandler initializes a registryHandler and configures top-level routing.
func newRegistryHandler(reg *catl.Registry) *registryHandler {
	rh := &registryHandler{reg: reg}
	rh.rebuildServers()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/catalogs", rh.handleListCatalogs)
	mux.HandleFunc("POST /v1/catalogs/reload", rh.handleReloadRegistry)
	mux.HandleFunc("/v1/catalogs/{catalog}/", rh.handleTenantDispatch)

	rh.mux = mux
	return rh
}

// ServeHTTP dispatches requests to the multi-catalog registry router.
func (rh *registryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rh.mux.ServeHTTP(w, r)
}

// handleListCatalogs handles GET /v1/catalogs by listing all available catalog names.
func (rh *registryHandler) handleListCatalogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, rh.reg.Names())
}

// handleReloadRegistry handles POST /v1/catalogs/reload by rescanning the registry directory.
func (rh *registryHandler) handleReloadRegistry(w http.ResponseWriter, r *http.Request) {
	if err := rh.reg.Reload(r.Context()); err != nil {
		writeProblem(w, r, http.StatusUnprocessableEntity, fmt.Sprintf("registry reload failed: %v", err))
		return
	}
	rh.rebuildServers()

	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"catalogs": rh.reg.Names(),
	})
}

// handleTenantDispatch handles /v1/catalogs/{catalog}/* by stripping the tenant prefix
// and delegating directly to the tenant's singleHandler.
func (rh *registryHandler) handleTenantDispatch(w http.ResponseWriter, r *http.Request) {
	catName := r.PathValue("catalog")
	sh := rh.getTenant(catName)
	if sh == nil {
		writeProblem(w, r, http.StatusNotFound, fmt.Sprintf("catalog %q not found in registry", catName))
		return
	}

	prefix := "/v1/catalogs/" + catName
	http.StripPrefix(prefix, sh).ServeHTTP(w, r)
}

// getTenant returns the singleHandler instance for a given catalog name.
func (rh *registryHandler) getTenant(name string) *singleHandler {
	servers := *rh.servers.Load()
	return servers[name]
}

// rebuildServers constructs singleHandler instances for all active tenant catalogs.
func (rh *registryHandler) rebuildServers() {
	servers := make(map[string]*singleHandler)
	for _, cat := range rh.reg.Catalogs() {
		servers[cat.Name()] = newSingleHandler(cat)
	}
	rh.servers.Store(&servers)
}

// writeProblem formats and sends an RFC 9457 application/problem+json response.
func writeProblem(w http.ResponseWriter, r *http.Request, status int, detail string) {
	p := Problem{
		Type:     "about:blank",
		Title:    http.StatusText(status),
		Status:   status,
		Detail:   detail,
		Instance: r.URL.Path,
	}

	w.Header().Set("Content-Type", MediaTypeProblemJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(p)
}

// writeJSON serializes data to w with an application/json content type.
func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}
