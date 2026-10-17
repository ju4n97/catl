// Package catlmcp integrates catl catalogs and registries with the official
// Model Context Protocol (MCP) Go SDK.
package catlmcp

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ju4n97/catl"
)

// defaultInstructions defines operational guidance injected into connected LLMs during initialization.
const defaultInstructions = `This server provides read-only access to declarative catalogs of structured data, schemas, and companion files managed by catl.

Rules for querying:
1. Entity identifiers strictly require qualified IDs in the form 'namespace.id' (e.g. 'platform.auth'). Bare IDs without a namespace prefix will be rejected.
2. Manifests are available as resources at 'catl://{collection}/{qualified_id}/manifest'.
3. Bundle entities contain companion files (documentation, runbooks, templates) available as resources at 'catl://{collection}/{qualified_id}/files/{path}'.
4. Use 'catl_list_collections' to discover available collections and entity counts.
5. Use 'catl_list_entities' to discover available entities across collections.
6. Use 'catl_get_entity' to read specific entity metadata and parsed manifest content.
7. Use 'catl_get_schema' to inspect a collection's JSON Schema Draft 2020-12 rules before drafting new manifests.
8. Use 'catl_validate_manifest' to dry-run validate a draft YAML/JSON manifest string against a collection's schema before saving.
9. Use 'catl_read_file' to directly query companion files (runbooks, documentation) from bundle entities.
10. Use 'catl_check' to verify that catalog manifests and schemas conform to JSON Schema Draft 2020-12 rules and content tree hashes.`

// Server wraps an official MCP server instance configured for catl.
type Server struct {
	server *mcp.Server
	cat    *catl.Catalog
	reg    *catl.Registry
}

type (
	// CheckInput defines parameters for the catl_check tool.
	CheckInput struct{}

	// CheckOutput defines the result payload of the catl_check tool.
	CheckOutput struct {
		Valid    bool     `json:"valid"`
		TreeHash string   `json:"tree_hash"`
		Summary  []string `json:"summary"`
	}
)

type (
	// ListCollectionsInput defines parameters for the catl_list_collections tool.
	ListCollectionsInput struct{}

	// CollectionInfo describes a collection's configuration and entity count.
	CollectionInfo struct {
		Catalog     string `json:"catalog"`
		Name        string `json:"name"`
		EntityCount int    `json:"entity_count"`
	}

	// ListCollectionsOutput defines the result payload of the catl_list_collections tool.
	ListCollectionsOutput struct {
		Collections []CollectionInfo `json:"collections"`
	}
)

type (
	// ListEntitiesInput defines input filters for listing catalog entities.
	ListEntitiesInput struct {
		Collection string `json:"collection,omitempty" jsonschema:"optional collection name filter"`
	}

	// ListEntitiesOutput defines the output format for entity listings.
	ListEntitiesOutput struct {
		Entities []string `json:"entities"`
	}
)

type (
	// GetEntityInput defines input parameters for resolving a catalog entity.
	GetEntityInput struct {
		Collection  string `json:"collection"   jsonschema:"name of the collection (e.g. services)"`
		QualifiedID string `json:"qualified_id" jsonschema:"canonical qualified identifier (e.g. platform.auth)"`
	}
)

type (
	// GetSchemaInput defines input parameters for the catl_get_schema tool.
	GetSchemaInput struct {
		Catalog    string `json:"catalog,omitempty" jsonschema:"optional catalog namespace (required in multi-catalog registries)"`
		Collection string `json:"collection"        jsonschema:"name of the collection to inspect (e.g. services)"`
	}

	// GetSchemaOutput defines the result payload containing the collection's JSON schema.
	GetSchemaOutput struct {
		Catalog    string `json:"catalog"`
		Collection string `json:"collection"`
		Permissive bool   `json:"permissive"`
		Schema     string `json:"schema,omitempty"`
	}
)

type (
	// ValidateManifestInput defines parameters for dry-run manifest validation.
	ValidateManifestInput struct {
		Catalog    string `json:"catalog,omitempty" jsonschema:"optional catalog namespace (required in multi-catalog registries)"`
		Collection string `json:"collection"        jsonschema:"name of the target collection"`
		Content    string `json:"content"           jsonschema:"raw YAML or JSON manifest string to validate"`
	}

	// ValidateManifestOutput defines the outcome of dry-run manifest validation.
	ValidateManifestOutput struct {
		Valid       bool              `json:"valid"`
		Diagnostics []catl.Diagnostic `json:"diagnostics,omitempty"`
	}
)

type (
	// ReadFileInput defines parameters for reading companion files from bundle entities.
	ReadFileInput struct {
		Collection  string `json:"collection"   jsonschema:"name of the collection (e.g. services)"`
		QualifiedID string `json:"qualified_id" jsonschema:"canonical qualified entity identifier (e.g. platform.billing)"`
		Path        string `json:"path"         jsonschema:"relative file path inside the bundle (e.g. runbook.md)"`
	}

	// ReadFileOutput defines the result payload containing companion file content and MIME type.
	ReadFileOutput struct {
		Path     string `json:"path"`
		MIMEType string `json:"mime_type"`
		Text     string `json:"text"`
	}
)

// New constructs an MCP Server backed by an individual catalog.
func New(cat *catl.Catalog) *Server {
	s := &Server{cat: cat}
	s.initServer("catl: " + cat.Name())
	return s
}

// NewRegistry constructs an MCP Server backed by a multi-tenant Registry.
func NewRegistry(reg *catl.Registry) *Server {
	s := &Server{reg: reg}
	s.initServer(fmt.Sprintf("catl registry (%d catalogs)", len(reg.Names())))
	return s
}

// MCPServer returns the underlying official mcp.Server pointer.
func (s *Server) MCPServer() *mcp.Server {
	return s.server
}

// Run executes the MCP server over a persistent transport (e.g. &mcp.StdioTransport{}).
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
	return s.server.Run(ctx, transport)
}

// initServer initializes capabilities, registers tools, and maps catalog resources.
func (s *Server) initServer(title string) {
	impl := &mcp.Implementation{
		Name:        "catl",
		Title:       title,
		Version:     "1.0.0",
		Description: "Verified declarative data substrate. Exposes schema-validated YAML and JSON manifests, bundle companion files, and integrity verification tools for catalogs and registries.",
		WebsiteURL:  "https://ju4n97.github.io/catl",
	}

	s.server = mcp.NewServer(impl, &mcp.ServerOptions{
		Instructions: defaultInstructions,
		Logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		})),
	})

	s.registerTools()
	s.registerResources()
}

// registerTools binds validation, resolution, and inspection operations to MCP tools.
func (s *Server) registerTools() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "catl_check",
		Description: "Validate catalog schemas, entity manifests, and content integrity",
	}, s.handleCheckTool)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "catl_list_collections",
		Description: "List all declared collections across catalogs with entity counts",
	}, s.handleListCollectionsTool)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "catl_list_entities",
		Description: "List all entities across collections or within a specified collection",
	}, s.handleListEntitiesTool)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "catl_get_entity",
		Description: "Get the complete metadata and parsed manifest for a qualified entity",
	}, s.handleGetEntityTool)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "catl_get_schema",
		Description: "Get the JSON Schema Draft 2020-12 definition for a collection",
	}, s.handleGetSchemaTool)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "catl_validate_manifest",
		Description: "Dry-run validate an uncommitted YAML or JSON manifest string against a collection's schema",
	}, s.handleValidateManifestTool)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "catl_read_file",
		Description: "Read a companion file (documentation, runbook, template) from a bundle entity",
	}, s.handleReadFileTool)
}

// handleCheckTool executes content verification and returns validation status and tree hashes.
func (s *Server) handleCheckTool(ctx context.Context, req *mcp.CallToolRequest, in CheckInput) (*mcp.CallToolResult, CheckOutput, error) {
	var summary []string
	var lastHash string

	for _, c := range s.catalogs() {
		h, err := c.TreeHash()
		if err != nil {
			return nil, CheckOutput{Valid: false}, fmt.Errorf("catalog %s check failed: %w", c.Name(), err)
		}
		lastHash = h
		summary = append(summary, fmt.Sprintf("Catalog %q valid. Content hash: %s (%d entities)", c.Name(), h, len(c.Entities())))
	}

	return nil, CheckOutput{
		Valid:    true,
		TreeHash: lastHash,
		Summary:  summary,
	}, nil
}

// handleListCollectionsTool lists all collections across single and multi-catalog modes.
func (s *Server) handleListCollectionsTool(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in ListCollectionsInput,
) (*mcp.CallToolResult, ListCollectionsOutput, error) {
	var res []CollectionInfo

	for _, c := range s.catalogs() {
		for _, name := range c.Collections() {
			res = append(res, CollectionInfo{
				Catalog:     c.Name(),
				Name:        name,
				EntityCount: len(c.Entities(name)),
			})
		}
	}

	return nil, ListCollectionsOutput{Collections: res}, nil
}

// handleListEntitiesTool returns a formatted list of active entities.
func (s *Server) handleListEntitiesTool(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in ListEntitiesInput,
) (*mcp.CallToolResult, ListEntitiesOutput, error) {
	var lines []string

	for _, c := range s.catalogs() {
		ents := c.Entities()
		if in.Collection != "" {
			ents = c.Entities(in.Collection)
		}
		for _, e := range ents {
			lines = append(lines, fmt.Sprintf("%s\t%s\t%s", e.Collection, e.QualifiedID, e.ManifestPath))
		}
	}

	return nil, ListEntitiesOutput{Entities: lines}, nil
}

// handleGetEntityTool resolves a qualified entity and returns its complete representation.
func (s *Server) handleGetEntityTool(ctx context.Context, req *mcp.CallToolRequest, in GetEntityInput) (*mcp.CallToolResult, any, error) {
	_, ent, err := s.resolveEntity(in.Collection, in.QualifiedID)
	if err != nil {
		return nil, nil, err
	}
	return nil, ent, nil
}

// handleGetSchemaTool retrieves the raw JSON schema text for a collection.
func (s *Server) handleGetSchemaTool(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in GetSchemaInput,
) (*mcp.CallToolResult, GetSchemaOutput, error) {
	c, err := s.selectCatalog(in.Catalog)
	if err != nil {
		return nil, GetSchemaOutput{}, err
	}

	schemaBytes, err := c.Schema(in.Collection)
	if err != nil {
		return nil, GetSchemaOutput{}, err
	}

	if schemaBytes == nil {
		return nil, GetSchemaOutput{
			Catalog:    c.Name(),
			Collection: in.Collection,
			Permissive: true,
		}, nil
	}

	return nil, GetSchemaOutput{
		Catalog:    c.Name(),
		Collection: in.Collection,
		Permissive: false,
		Schema:     string(schemaBytes),
	}, nil
}

// handleValidateManifestTool dry-run validates a raw manifest string against a collection's schema.
func (s *Server) handleValidateManifestTool(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in ValidateManifestInput,
) (*mcp.CallToolResult, ValidateManifestOutput, error) {
	c, err := s.selectCatalog(in.Catalog)
	if err != nil {
		return nil, ValidateManifestOutput{Valid: false}, err
	}

	err = c.ValidateManifest(in.Collection, []byte(in.Content))
	if err == nil {
		return nil, ValidateManifestOutput{Valid: true}, nil
	}

	if valErr, ok := errors.AsType[*catl.ValidationError](err); ok {
		return nil, ValidateManifestOutput{
			Valid:       false,
			Diagnostics: valErr.Diagnostics,
		}, nil
	}

	return nil, ValidateManifestOutput{
		Valid: false,
		Diagnostics: []catl.Diagnostic{
			{Message: err.Error()},
		},
	}, nil
}

// handleReadFileTool reads an individual companion file from a bundle entity.
func (s *Server) handleReadFileTool(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	in ReadFileInput,
) (*mcp.CallToolResult, ReadFileOutput, error) {
	_, ent, err := s.resolveEntity(in.Collection, in.QualifiedID)
	if err != nil {
		return nil, ReadFileOutput{}, err
	}

	if !ent.IsBundle() {
		return nil, ReadFileOutput{}, fmt.Errorf("entity %q is a flat entity and has no companion files", ent.ID)
	}

	bfs, err := ent.FS()
	if err != nil {
		return nil, ReadFileOutput{}, err
	}

	data, err := fs.ReadFile(bfs, in.Path)
	if err != nil {
		return nil, ReadFileOutput{}, fmt.Errorf("read file %q from bundle: %w", in.Path, err)
	}

	mType := mime.TypeByExtension(filepath.Ext(in.Path))
	if mType == "" {
		mType = "text/plain"
	}

	return nil, ReadFileOutput{
		Path:     in.Path,
		MIMEType: mType,
		Text:     string(data),
	}, nil
}

// registerResources iterates over all entities and companion files, registering catl:// URIs.
func (s *Server) registerResources() {
	for _, c := range s.catalogs() {
		for _, ent := range c.Entities() {
			capturedEnt := ent

			manifestMime := "application/yaml"
			if path.Ext(capturedEnt.ManifestPath) == ".json" {
				manifestMime = "application/json"
			}

			manifestURI := fmt.Sprintf("catl://%s/%s/manifest", capturedEnt.Collection, capturedEnt.QualifiedID)

			s.server.AddResource(
				&mcp.Resource{
					URI:         manifestURI,
					Name:        capturedEnt.QualifiedID + " manifest",
					Description: fmt.Sprintf("Entity manifest for %s in %s", capturedEnt.QualifiedID, capturedEnt.Collection),
					MIMEType:    manifestMime,
				},
				func(_ context.Context, rrr *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
					return &mcp.ReadResourceResult{
						Contents: []*mcp.ResourceContents{
							{
								URI:      rrr.Params.URI,
								MIMEType: manifestMime,
								Text:     string(capturedEnt.RawManifest()),
							},
						},
					}, nil
				},
			)

			if capturedEnt.IsBundle() {
				for _, f := range capturedEnt.CompanionFiles {
					capturedFile := f
					companionMime := mime.TypeByExtension(filepath.Ext(capturedFile))
					if companionMime == "" {
						companionMime = "text/plain"
					}

					fileURI := fmt.Sprintf("catl://%s/%s/files/%s", capturedEnt.Collection, capturedEnt.QualifiedID, capturedFile)

					s.server.AddResource(
						&mcp.Resource{
							URI:         fileURI,
							Name:        fmt.Sprintf("%s %s", capturedEnt.QualifiedID, capturedFile),
							Description: fmt.Sprintf("Companion file %s for %s", capturedFile, capturedEnt.QualifiedID),
							MIMEType:    companionMime,
						},
						func(ctx context.Context, rrr *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
							bfs, err := capturedEnt.FS()
							if err != nil {
								return nil, err
							}

							data, err := fs.ReadFile(bfs, capturedFile)
							if err != nil {
								return nil, err
							}

							return &mcp.ReadResourceResult{
								Contents: []*mcp.ResourceContents{
									{
										URI:      rrr.Params.URI,
										MIMEType: companionMime,
										Text:     string(data),
									},
								},
							}, nil
						},
					)
				}
			}
		}
	}
}

// resolveEntity queries an entity across single-catalog and multi-catalog modes.
func (s *Server) resolveEntity(collection, qualifiedID string) (*catl.Catalog, *catl.Entity, error) {
	ns, _, hasDot := strings.Cut(qualifiedID, ".")
	if !hasDot {
		return nil, nil, fmt.Errorf("%w: %q (want namespace.id)", catl.ErrNotQualified, qualifiedID)
	}

	if s.cat != nil {
		if ns != s.cat.Name() {
			return nil, nil, fmt.Errorf("%w: namespace %q", catl.ErrNotFound, ns)
		}
		ent, err := s.cat.Resolve(collection, qualifiedID)
		return s.cat, ent, err
	}

	if s.reg != nil {
		target, ok := s.reg.Catalog(ns)
		if !ok {
			return nil, nil, fmt.Errorf("%w: catalog %q not in registry", catl.ErrNotFound, ns)
		}
		ent, err := target.Resolve(collection, qualifiedID)
		return target, ent, err
	}

	return nil, nil, errors.New("mcp server has no loaded catalogs")
}

// selectCatalog resolves the target catalog from single or multi-catalog configurations.
func (s *Server) selectCatalog(name string) (*catl.Catalog, error) {
	if s.cat != nil {
		if name != "" && name != s.cat.Name() {
			return nil, fmt.Errorf("catalog %q does not match active catalog %q", name, s.cat.Name())
		}
		return s.cat, nil
	}

	if s.reg != nil {
		if name == "" {
			names := s.reg.Names()
			if len(names) == 1 {
				cat, _ := s.reg.Catalog(names[0])
				return cat, nil
			}
			return nil, fmt.Errorf("multi-catalog registry requires explicit 'catalog' parameter (available: %v)", names)
		}

		cat, ok := s.reg.Catalog(name)
		if !ok {
			return nil, fmt.Errorf("catalog %q not found in registry", name)
		}
		return cat, nil
	}

	return nil, errors.New("server has no loaded catalogs")
}

// catalogs returns the active set of catalogs being served.
func (s *Server) catalogs() []*catl.Catalog {
	if s.cat != nil {
		return []*catl.Catalog{s.cat}
	}
	if s.reg != nil {
		return s.reg.Catalogs()
	}
	return nil
}
