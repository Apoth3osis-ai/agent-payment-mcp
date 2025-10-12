# JSON Schema 2020-12 Compliance Fix

## Problem
Claude Code CLI was returning the following error:
```
API Error: 400 {"type":"error","error":{"type":"invalid_request_error","message":"tools.17.custom.input_schema: JSON schema is invalid. It must match JSON Schema draft 2020-12 (https://json-schema.org/draft/2020-12). Learn more about tool use at https://docs.claude.com/en/docs/tool-use."},"request_id":"req_011CU47w21SGhoMzzC2sSvyx"}
```

## Root Cause
The Agent Payment API was returning tool schemas with **non-standard** JSON Schema format:

### ❌ Invalid Schema (from API)
```json
{
  "type": "object",
  "properties": {
    "field_name": {
      "type": "string",
      "required": true    // ❌ NOT VALID in JSON Schema 2020-12!
    }
  }
}
```

In JSON Schema draft 2020-12, the `required` property cannot be inside individual properties. It must be an array at the object level.

### ✅ Valid Schema (JSON Schema 2020-12)
```json
{
  "type": "object",
  "properties": {
    "field_name": {
      "type": "string"
    }
  },
  "required": ["field_name"]    // ✅ CORRECT
}
```

## Solution
Added `sanitizeJSONSchema()` function in `/mcp-server/internal/mcp/server.go` that:

1. **Detects** `"required": true` or `"required": false` inside property definitions
2. **Removes** the invalid `required` field from each property
3. **Collects** properties with `"required": true` into a list
4. **Creates/updates** a top-level `"required"` array with those property names
5. **Preserves** any existing required fields from the original schema

## Changes Made

### File: `mcp-server/internal/mcp/server.go`

1. Added `sanitizeJSONSchema()` function (lines ~188-260)
2. Updated `registerTool()` to call sanitization before sentence case fixing:
   ```go
   // Sanitize schema to be JSON Schema 2020-12 compliant
   sanitizedParams := sanitizeJSONSchema(toolDef.Function.Parameters)
   
   // Fix sentence case in parameter descriptions/examples
   fixedParams := fixSentenceCaseInSchema(sanitizedParams)
   ```

### File: `mcp-server/internal/mcp/schema_test.go` (NEW)
Added comprehensive unit tests to verify:
- Moving `required: true` from properties to required array
- Preserving existing required arrays
- Handling schemas without properties
- Handling empty schemas
- Ensuring no `required` field remains in properties

## Testing
```bash
cd mcp-server
go test -v ./internal/mcp -run TestSanitizeJSONSchema
```

All tests pass ✅

## Deployment
Rebuilt all platform binaries with the fix:
```bash
bash scripts/build-all.sh
```

Binaries updated:
- `distribution/binaries/darwin-amd64/agent-payment-server`
- `distribution/binaries/darwin-arm64/agent-payment-server`
- `distribution/binaries/linux-amd64/agent-payment-server`
- `distribution/binaries/windows-amd64/agent-payment-server.exe`

## How to Apply
If you've already installed the MCP server, reinstall using:
```bash
# Linux/macOS
bash distribution/templates/install-linux.sh
# or
bash distribution/templates/install-macos.sh

# Windows
.\distribution\templates\install-windows.ps1
```

Or rebuild locally:
```bash
cd mcp-server
go build -o agent-payment-server ./cmd/agent-payment-server/
```

## Verification
After updating, the Claude Code CLI should no longer show the JSON schema validation error. All tools should be properly registered and usable.

## Technical Details
The sanitization happens during tool registration in the MCP server, before tools are exposed to Claude. This ensures:
1. **Backward compatibility** - Works with existing API responses
2. **Standards compliance** - Outputs valid JSON Schema 2020-12
3. **No API changes needed** - Fixes the issue on the MCP server side
4. **Transparent to users** - Tools work exactly as expected

## Related Files
- `/mcp-server/internal/mcp/server.go` - Main server logic with sanitization
- `/mcp-server/internal/mcp/schema_test.go` - Unit tests for sanitization
- `/mcp-server/internal/mcp/rpc.go` - RPC handler (uses sanitized schemas)
- `/mcp-server/internal/api/client.go` - API client (receives raw schemas)
