package catlhttp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ju4n97/catl"
	"github.com/ju4n97/catl/catlhttp"
)

// newTestCatalog scaffolds an in-memory catalog for server testing.
func newTestCatalog(t *testing.T, name string) *catl.Catalog {
	t.Helper()

	fsys := fstest.MapFS{
		"catl.toml": &fstest.MapFile{
			Data: []byte("[catalog]\nname = \"" + name + "\"\n[collections.services]\n"),
		},
		"services/auth.yaml": &fstest.MapFile{
			Data: []byte("tier: critical\nactive: true\n"),
		},
		"services/billing/manifest.json": &fstest.MapFile{
			Data: []byte(`{"tier": "standard", "active": false}`),
		},
		"services/billing/notes.md": &fstest.MapFile{
			Data: []byte("# Billing Notes\n"),
		},
	}

	cat, err := catl.OpenFS(context.Background(), fsys, catl.Options{})
	if err != nil {
		t.Fatalf("OpenFS() failed: %v", err)
	}
	return cat
}

// decodeProblem validates the Content-Type header and unmarshals an RFC 9457 problem payload.
func decodeProblem(t *testing.T, rec *httptest.ResponseRecorder) catlhttp.Problem {
	t.Helper()

	if ct := rec.Header().Get("Content-Type"); ct != catlhttp.MediaTypeProblemJSON {
		t.Fatalf("Content-Type = %q, want %q", ct, catlhttp.MediaTypeProblemJSON)
	}

	var p catlhttp.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("failed to decode RFC 9457 problem JSON: %v", err)
	}
	return p
}

// TestEndpointsSync verifies that EndpointsSummary contains all registered routes.
func TestEndpointsSync(t *testing.T) {
	t.Parallel()

	routes := catlhttp.Routes()
	summary := catlhttp.EndpointsSummary()

	for _, r := range routes {
		if !strings.Contains(summary, r.Method) {
			t.Errorf("EndpointsSummary() missing method %s", r.Method)
		}
		if !strings.Contains(summary, r.Path) {
			t.Errorf("EndpointsSummary() missing path %s", r.Path)
		}
	}
}

// TestSingleCatalog verifies all RESTful endpoints for single-catalog servers.
func TestSingleCatalog(t *testing.T) {
	t.Parallel()

	cat := newTestCatalog(t, "platform")
	srv := catlhttp.New(cat)

	t.Run("get catalog metadata", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/catalog", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var summary catlhttp.CatalogSummary
		if err := json.Unmarshal(rec.Body.Bytes(), &summary); err != nil {
			t.Fatalf("decode json: %v", err)
		}
		if summary.Name != "platform" || len(summary.Collections) != 1 {
			t.Errorf("unexpected summary: %+v", summary)
		}
	})

	t.Run("get catalog entities inventory", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/catalog/entities", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var ents []catl.Entity
		if err := json.Unmarshal(rec.Body.Bytes(), &ents); err != nil {
			t.Fatalf("decode json: %v", err)
		}
		if len(ents) != 2 {
			t.Fatalf("got %d entities, want 2", len(ents))
		}
	})

	t.Run("bare id returns 400 bad request with suggestion", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/collections/services/entities/auth", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}

		p := decodeProblem(t, rec)
		if !strings.Contains(p.Detail, `did you mean "platform.auth"?`) {
			t.Errorf("problem.Detail = %q, want hint to platform.auth", p.Detail)
		}
	})

	t.Run("get entity resource returns metadata and parsed manifest", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/collections/services/entities/platform.auth", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var res catlhttp.EntityResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("decode json: %v", err)
		}

		if res.ID != "auth" || res.QualifiedID != "platform.auth" || res.IsBundle {
			t.Errorf("unexpected entity response fields: %+v", res)
		}

		manifestMap, ok := res.Manifest.(map[string]any)
		if !ok || manifestMap["tier"] != "critical" {
			t.Errorf("unexpected parsed manifest: %+v", res.Manifest)
		}
	})

	t.Run("get entity manifest stream returns raw bytes", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/collections/services/entities/platform.auth/manifest", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if !strings.Contains(rec.Header().Get("Content-Type"), "application/yaml") {
			t.Errorf("Content-Type = %s, want application/yaml", rec.Header().Get("Content-Type"))
		}
		if !strings.Contains(rec.Body.String(), "tier: critical") {
			t.Errorf("unexpected body: %s", rec.Body.String())
		}
	})

	t.Run("list companion files in bundle entity", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/collections/services/entities/platform.billing/files", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var files []string
		if err := json.Unmarshal(rec.Body.Bytes(), &files); err != nil {
			t.Fatalf("decode json: %v", err)
		}
		if len(files) != 1 || files[0] != "notes.md" {
			t.Errorf("got files %v, want ['notes.md']", files)
		}
	})

	t.Run("list companion files on flat entity returns 400 bad request", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/collections/services/entities/platform.auth/files", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}

		p := decodeProblem(t, rec)
		if !strings.Contains(p.Detail, "flat entity and has no companion files") {
			t.Errorf("unexpected problem detail: %s", p.Detail)
		}
	})

	t.Run("stream companion file from bundle", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/collections/services/entities/platform.billing/files/notes.md", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if !strings.Contains(rec.Body.String(), "# Billing Notes") {
			t.Errorf("unexpected content: %s", rec.Body.String())
		}
	})

	t.Run("missing entity returns 404 problem", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/collections/services/entities/platform.nonexistent", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}

		p := decodeProblem(t, rec)
		if p.Status != http.StatusNotFound || p.Title != "Not Found" {
			t.Errorf("unexpected problem: %+v", p)
		}
	})
}

// TestMultiCatalogRegistry verifies tenant routing, namespace isolation, and clean path prefixing.
func TestMultiCatalogRegistry(t *testing.T) {
	t.Parallel()

	cat1 := newTestCatalog(t, "compliance")
	cat2 := newTestCatalog(t, "billing")
	reg := catl.NewRegistry(cat1, cat2)
	srv := catlhttp.NewRegistry(reg)

	t.Run("list all tenant catalogs", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/catalogs", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var names []string
		if err := json.Unmarshal(rec.Body.Bytes(), &names); err != nil {
			t.Fatalf("decode json: %v", err)
		}
		if len(names) != 2 || names[0] != "billing" || names[1] != "compliance" {
			t.Errorf("got catalogs %v, want ['billing', 'compliance']", names)
		}
	})

	t.Run("route tenant catalog metadata via path", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/catalogs/billing/catalog", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var summary catlhttp.CatalogSummary
		if err := json.Unmarshal(rec.Body.Bytes(), &summary); err != nil {
			t.Fatalf("decode json: %v", err)
		}
		if summary.Name != "billing" {
			t.Errorf("summary.Name = %q, want 'billing'", summary.Name)
		}
	})

	t.Run("route tenant entity resource via path", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/catalogs/billing/collections/services/entities/billing.auth", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var res catlhttp.EntityResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("decode json: %v", err)
		}
		if res.QualifiedID != "billing.auth" {
			t.Errorf("res.QualifiedID = %q, want 'billing.auth'", res.QualifiedID)
		}
	})

	t.Run("stream tenant companion file via path", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/catalogs/billing/collections/services/entities/billing.billing/files/notes.md", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if !strings.Contains(rec.Body.String(), "# Billing Notes") {
			t.Errorf("unexpected content: %s", rec.Body.String())
		}
	})

	t.Run("unknown catalog in prefix returns 404 problem", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/catalogs/unknown/catalog", nil)
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}

		p := decodeProblem(t, rec)
		if !strings.Contains(p.Detail, `catalog "unknown" not found`) {
			t.Errorf("unexpected detail: %s", p.Detail)
		}
	})
}

// TestSingleCatalogReload verifies atomic reload from disk and error handling.
func TestSingleCatalogReload(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	catlTOML := filepath.Join(tmpDir, "catl.toml")
	itemsDir := filepath.Join(tmpDir, "items")

	if err := os.MkdirAll(itemsDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	writeCat := func(ver string) {
		toml := "[catalog]\nname = \"demo\"\n[collections.items]\n"
		_ = os.WriteFile(catlTOML, []byte(toml), 0o600)
		_ = os.WriteFile(filepath.Join(itemsDir, ver+".yaml"), []byte("ok: true\n"), 0o600)
	}

	writeCat("v1")
	cat, err := catl.Open(context.Background(), tmpDir, catl.Options{Offline: true})
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}

	srv := catlhttp.New(cat)

	// Update on disk
	writeCat("v2")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/catalog/reload", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("reload status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Verify new entity is live
	recGet := httptest.NewRecorder()
	reqGet := httptest.NewRequest(http.MethodGet, "/v1/collections/items/entities/demo.v2", nil)
	srv.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Errorf("entity demo.v2 query status = %d, want %d", recGet.Code, http.StatusOK)
	}
}
