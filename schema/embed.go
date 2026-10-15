package schema

import _ "embed"

// CatlJSON contains the raw bytes of the authoritative catl.schema.json.
//
//go:embed catl.schema.json
var CatlJSON []byte
