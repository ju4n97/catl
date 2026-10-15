package catlmcp_test

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ju4n97/catl"
	"github.com/ju4n97/catl/catlmcp"
)

// newTestCatalog scaffolds an in-memory catalog for MCP testing.
func newTestCatalog(t *testing.T, name string) *catl.Catalog {
	t.Helper()

	fsys := fstest.MapFS{
		"catl.toml": &fstest.MapFile{
			Data: []byte("[catalog]\nname = \"" + name + "\"\n[collections.services]\nschema = \"schemas/service.json\"\n"),
		},
		"schemas/service.json": &fstest.MapFile{
			Data: []byte(`{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"type": "object",
				"required": ["tier"],
				"properties": {
					"tier": { "type": "string" },
					"active": { "type": "boolean" }
				},
				"additionalProperties": false
			}`),
		},
		"services/auth.yaml": &fstest.MapFile{
			Data: []byte("tier: critical\nactive: true\n"),
		},
		"services/billing/manifest.json": &fstest.MapFile{
			Data: []byte(`{"tier": "standard", "active": false}`),
		},
		"services/billing/runbook.md": &fstest.MapFile{
			Data: []byte("# Billing Runbook Guide\n"),
		},
	}

	cat, err := catl.OpenFS(context.Background(), fsys)
	if err != nil {
		t.Fatalf("OpenFS() failed: %v", err)
	}
	return cat
}

// setupClientSession connects an in-memory mcp.Client directly to the server.
func setupClientSession(t *testing.T, srv *catlmcp.Server) *mcp.ClientSession {
	t.Helper()

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := srv.MCPServer().Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect() failed: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() failed: %v", err)
	}

	t.Cleanup(func() {
		_ = clientSession.Close()
		_ = serverSession.Close()
	})

	return clientSession
}

// TestOfficialMCPServer exercises tools and resource streaming using the official in-memory transport.
func TestOfficialMCPServer(t *testing.T) {
	t.Parallel()

	cat := newTestCatalog(t, "platform")
	srv := catlmcp.New(cat)
	session := setupClientSession(t, srv)
	ctx := context.Background()

	t.Run("list tools includes all operations", func(t *testing.T) {
		t.Parallel()

		tools, err := session.ListTools(ctx, nil)
		if err != nil {
			t.Fatalf("ListTools() failed: %v", err)
		}

		if len(tools.Tools) != 7 {
			t.Fatalf("got %d tools, want 7", len(tools.Tools))
		}
	})

	t.Run("call tool catl_check executes validation", func(t *testing.T) {
		t.Parallel()

		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "catl_check",
			Arguments: map[string]any{},
		})
		if err != nil {
			t.Fatalf("CallTool(catl_check) failed: %v", err)
		}

		if res.IsError {
			t.Errorf("tool returned error: %+v", res)
		}
	})

	t.Run("call tool catl_list_collections lists declared collections", func(t *testing.T) {
		t.Parallel()

		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "catl_list_collections",
			Arguments: map[string]any{},
		})
		if err != nil {
			t.Fatalf("CallTool(catl_list_collections) failed: %v", err)
		}
		if res.IsError {
			t.Fatalf("tool returned error: %+v", res)
		}
	})

	t.Run("call tool catl_list_entities returns all entities", func(t *testing.T) {
		t.Parallel()

		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "catl_list_entities",
			Arguments: map[string]any{},
		})
		if err != nil {
			t.Fatalf("CallTool(catl_list_entities) failed: %v", err)
		}
		if res.IsError {
			t.Fatalf("tool returned error: %+v", res)
		}
	})

	t.Run("call tool catl_get_entity returns entity payload", func(t *testing.T) {
		t.Parallel()

		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "catl_get_entity",
			Arguments: map[string]any{
				"collection":   "services",
				"qualified_id": "platform.auth",
			},
		})
		if err != nil {
			t.Fatalf("CallTool(catl_get_entity) failed: %v", err)
		}

		if res.IsError {
			t.Fatalf("expected successful resolve: %+v", res)
		}
	})

	t.Run("call tool catl_get_schema returns raw json schema", func(t *testing.T) {
		t.Parallel()

		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "catl_get_schema",
			Arguments: map[string]any{
				"collection": "services",
			},
		})
		if err != nil {
			t.Fatalf("CallTool(catl_get_schema) failed: %v", err)
		}
		if res.IsError {
			t.Fatalf("tool returned error: %+v", res)
		}
	})

	t.Run("call tool catl_validate_manifest with valid content", func(t *testing.T) {
		t.Parallel()

		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "catl_validate_manifest",
			Arguments: map[string]any{
				"collection": "services",
				"content":    "tier: critical\nactive: true\n",
			},
		})
		if err != nil {
			t.Fatalf("CallTool(catl_validate_manifest) failed: %v", err)
		}
		if res.IsError {
			t.Fatalf("tool returned error: %+v", res)
		}
	})

	t.Run("call tool catl_read_file returns companion file content", func(t *testing.T) {
		t.Parallel()

		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "catl_read_file",
			Arguments: map[string]any{
				"collection":   "services",
				"qualified_id": "platform.billing",
				"path":         "runbook.md",
			},
		})
		if err != nil {
			t.Fatalf("CallTool(catl_read_file) failed: %v", err)
		}
		if res.IsError {
			t.Fatalf("tool returned error: %+v", res)
		}
	})

	t.Run("read resource returns companion file content", func(t *testing.T) {
		t.Parallel()

		res, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
			URI: "catl://services/platform.billing/files/runbook.md",
		})
		if err != nil {
			t.Fatalf("ReadResource() failed: %v", err)
		}

		if len(res.Contents) != 1 {
			t.Fatalf("got %d contents, want 1", len(res.Contents))
		}

		if !strings.Contains(res.Contents[0].Text, "# Billing Runbook Guide") {
			t.Errorf("content mismatch: %s", res.Contents[0].Text)
		}
	})
}
