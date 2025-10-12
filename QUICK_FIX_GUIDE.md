# How to Apply the JSON Schema Fix

## Problem Summary
Your Claude Code CLI was showing this error:
```
API Error: 400 - tools.17.custom.input_schema: JSON schema is invalid
```

This was caused by the Agent Payment API returning JSON schemas with `"required": true` inside property definitions, which violates JSON Schema 2020-12 specification.

## Solution Applied ✅

I've added automatic schema sanitization to the MCP server that:
1. Detects `"required": true` inside properties
2. Removes it from properties  
3. Moves required fields to a proper top-level `"required"` array

## What Was Changed

### Files Modified:
1. **`mcp-server/internal/mcp/server.go`**
   - Added `sanitizeJSONSchema()` function
   - Updated `registerTool()` to sanitize schemas before registration
   - Added logging to show when schemas are sanitized

### Files Created:
2. **`mcp-server/internal/mcp/schema_test.go`** - Unit tests
3. **`SCHEMA_FIX.md`** - Detailed documentation
4. **`scripts/verify-schema-fix.sh`** - Verification script

## How to Use the Fixed Version

### Option 1: Use Pre-built Binaries (Easiest)
The binaries have already been rebuilt with the fix. Just reinstall:

```bash
# For Linux
cd /home/richard/Documents/agentpmt/local_mcp
bash distribution/templates/install-linux.sh

# For macOS (if you use that)
bash distribution/templates/install-macos.sh
```

### Option 2: Build from Source
If you prefer to build yourself:

```bash
cd /home/richard/Documents/agentpmt/local_mcp/mcp-server
go build -o agent-payment-server ./cmd/agent-payment-server/
```

Then copy the binary to your installation location.

## Testing the Fix

1. **Run the verification script:**
   ```bash
   cd /home/richard/Documents/agentpmt/local_mcp
   bash scripts/verify-schema-fix.sh
   ```
   You should see all green checkmarks ✅

2. **Test with Claude Code CLI:**
   After reinstalling, try using Claude Code CLI again. The schema error should be gone.

3. **Check logs:**
   When the MCP server starts, it will log messages like:
   ```
   Sanitized schema for tool <tool-name> (fixed 'required' fields)
   ```
   This confirms the fix is working.

## What to Expect

- ✅ No more "JSON schema is invalid" errors
- ✅ All tools properly registered with Claude
- ✅ Tools work exactly as before (functionality unchanged)
- ✅ Transparent fix - you won't notice any difference except the error is gone

## Verification Checklist

- [x] Schema sanitization function added
- [x] Function integrated into tool registration
- [x] Unit tests created and passing
- [x] All platform binaries rebuilt (Linux, macOS Intel, macOS ARM, Windows)
- [x] Documentation created
- [x] Verification script created

## Need Help?

If you still see the error after applying the fix:
1. Check that you're using the newly built binary
2. Look for "Sanitized schema" log messages when server starts
3. Share the server logs for debugging

## Technical Details

See `SCHEMA_FIX.md` for complete technical documentation including:
- Root cause analysis
- Before/after schema examples
- Implementation details
- Testing strategy

---

**Status**: ✅ **READY TO USE**

All changes have been implemented, tested, and built. Just reinstall the MCP server and you're good to go!
