package schema

import _ "embed"

// CatlBytes contains the raw bytes of the authoritative catl.schema.json.
//
//go:embed catl.schema.json
var CatlBytes []byte
