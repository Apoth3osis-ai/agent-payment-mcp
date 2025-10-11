# Architecture Decision Document

## Executive Summary

This document explains the design decisions for the Agent Payment MCP Server, which dynamically loads tools from a REST API and proxies execution requests back to that API.

## Problem Statement

We need an MCP server that:
1. Fetches tool definitions from a REST API at runtime (unknown at compile time)
2. Registers any number of dynamic tools with MCP clients
3. Proxies tool execution requests back to the REST API
4. Provides excellent user experience in Claude/Cursor/VS Code
5. Remains simple and maintainable (< 500 lines)
6. Compiles to a small, standalone executable

## Design Approach Selected: Individual Dynamic Tool Registration

### The Chosen Solution

We register each API tool as a separate MCP tool using the `Server.AddTool()` method with dynamic schemas. Each tool:
- Has its own name, description, and parameter schema
- Appears as an individual tool in MCP clients
- Has a dedicated handler that proxies to the API
- Is indistinguishable from compile-time tools

### Why This Approach?

#### Pros
✅ **Best User Experience**: Each tool appears individually in Claude's tool list with full descriptions and schemas
✅ **Native MCP Integration**: Uses the official Go SDK's intended pattern
✅ **Familiar to Users**: Tools work exactly like built-in Claude tools
✅ **Automatic Validation**: MCP SDK validates parameters against schemas
✅ **Discovery**: Users can browse tools, see descriptions, and understand parameters
✅ **Flexibility**: Supports any number of tools with any schemas
✅ **Simplicity**: Clean, straightforward code (~200 lines for MCP server)

#### Cons
❌ **Startup Time**: Must fetch and register all tools before accepting connections (~1-2 seconds)
❌ **Static After Startup**: Can't add new tools without restart (acceptable tradeoff)
❌ **Memory Per Tool**: Each tool has its own handler closure (~1KB per tool, negligible for 100s of tools)

### Alternatives Considered

#### Option A: Single Generic "Execute Tool" Tool

**Concept**: Register one MCP tool named `agent_payment_execute` that takes `tool_name` and `parameters`.

```go
Tool{
    Name: "agent_payment_execute",
    Description: "Execute any Agent Payment tool",
    Parameters: {
        tool_name: string (required),
        parameters: object (optional)
    }
}
```

**User Experience**:
```
User: Check weather in NYC
Claude: I'll use the agent_payment_execute tool
[Calls: agent_payment_execute(tool_name="weather_check", parameters={city: "NYC"})]
```

**Pros**:
- Simple implementation (one tool registration)
- Can add new tools without changing server
- Minimal startup time

**Cons**:
- ❌ **Poor UX**: User must know exact tool names
- ❌ **No Discovery**: Can't browse available tools in UI
- ❌ **No Validation**: Parameters not validated against schemas
- ❌ **Confusing**: Not obvious what tools are available
- ❌ **Un-ergonomic**: Extra layer of indirection

**Verdict**: Rejected. User experience is paramount, and this fails that test.

#### Option B: Tools as MCP Resources

**Concept**: Expose tools via MCP Resources, execute via generic tool.

```go
Resources: [
    "tool://weather_check",
    "tool://stock_price",
    ...
]
Tool: {
    Name: "execute_agent_payment_tool",
    Parameters: {resource_uri: string, parameters: object}
}
```

**User Experience**:
```
User: What tools do I have?
Claude: I see resources: tool://weather_check, tool://stock_price...
```

**Pros**:
- Tools discoverable via `listResources`
- Metadata can be rich (descriptions, schemas in resource content)
- Separation of concerns (discovery vs execution)

**Cons**:
- ❌ **Awkward UX**: Resources aren't tools - users expect tools
- ❌ **Two-Step Process**: List resources, then call tool with resource name
- ❌ **Misuse of Resources**: Resources are for documents/data, not tool definitions
- ❌ **Complex**: More code, more concepts
- ❌ **Non-Standard**: Not how MCP tools are meant to work

**Verdict**: Rejected. Clever but confusing. Misuses MCP primitives.

#### Option C: Tools as MCP Prompts

**Concept**: Expose each tool as an MCP Prompt, execute via generic tool.

```go
Prompts: [
    {name: "weather_check", description: "...", arguments: [...]},
    ...
]
Tool: {
    Name: "execute_tool",
    Parameters: {prompt_name: string, parameters: object}
}
```

**User Experience**:
```
User: What tools do I have?
Claude: I see prompts: weather_check, stock_price...
[User confused: are these prompts or tools?]
```

**Pros**:
- Prompts have descriptions and argument schemas
- Discoverable via `listPrompts`

**Cons**:
- ❌ **Conceptual Mismatch**: Prompts are for prompt templates, not tools
- ❌ **Confusing UX**: Users expect tools, get prompts
- ❌ **Indirect**: Must map prompt → tool call
- ❌ **Abuse of MCP**: Wrong primitive for the job

**Verdict**: Rejected. Even more of a conceptual stretch than resources.

#### Option D: Single Tool with Enum Schema

**Concept**: One tool with dynamically-generated enum of tool names.

```go
Tool: {
    Name: "call_tool",
    Parameters: {
        tool_name: {
            type: "string",
            enum: ["weather_check", "stock_price", ...]
        },
        // Dynamic parameters based on tool_name
    }
}
```

**User Experience**:
```
User: Check weather in NYC
Claude: [Calls: call_tool(tool_name="weather_check", ...)]
```

**Pros**:
- Tool name is validated (must be in enum)
- Single tool registration
- Discovery via enum values

**Cons**:
- ❌ **Can't Handle Dynamic Parameters**: Each tool has different parameters
- ❌ **No Per-Tool Descriptions**: All tools share same description
- ❌ **Complex Schema**: Trying to combine 50+ different parameter schemas
- ❌ **Poor UX**: Tools don't appear individually in UI
- ❌ **Validation Nightmare**: Can't validate parameters correctly

**Verdict**: Rejected. Technically complex, poor UX, doesn't scale.

### Decision Matrix

| Approach | UX Score | Simplicity | Native MCP | Discovery | Validation | Selected |
|----------|----------|------------|------------|-----------|------------|----------|
| **Individual Tools (Selected)** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ✅ |
| Single Generic Tool | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐ | ⭐⭐ | ❌ |
| Resources | ⭐⭐ | ⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ❌ |
| Prompts | ⭐ | ⭐⭐ | ⭐ | ⭐⭐ | ⭐⭐ | ❌ |
| Enum Schema | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ❌ |

## Implementation Details

### Technology: Official Go MCP SDK

**Repository**: `github.com/modelcontextprotocol/go-sdk`

**Why Go SDK?**
- ✅ Official SDK maintained by MCP team + Google
- ✅ Small binary size (6-8MB static build)
- ✅ Fast startup and execution
- ✅ Strong typing and compile-time safety
- ✅ Excellent standard library (HTTP, JSON, concurrency)
- ✅ Easy cross-compilation for multiple platforms
- ✅ Simple deployment (single binary)

**Why NOT alternatives?**

**Python FastMCP**:
- ❌ Larger binaries (40MB+ with PyInstaller)
- ❌ Slower startup time
- ❌ More complex distribution
- ✅ Simpler code (would be easier)
- ✅ More examples available

**TypeScript SDK**:
- ❌ Even larger binaries (50MB+)
- ❌ Requires Node.js runtime or bundler
- ❌ More complex build process
- ✅ Excellent documentation
- ✅ More mature ecosystem

**Rust SDK**:
- ✅ Smallest binaries (2-3MB)
- ✅ Fastest execution
- ❌ Steeper learning curve
- ❌ Harder to maintain
- ❌ Less documentation

**Verdict**: Go SDK provides the best balance of simplicity, performance, and deployment ease.

### Key Implementation Patterns

#### 1. Dynamic Tool Registration

```go
func (s *Server) registerTool(toolDef api.ToolDefinition) error {
    // Parse schema from API response
    var schema map[string]interface{}
    json.Unmarshal(toolDef.Function.Parameters, &schema)

    // Create MCP tool with dynamic schema
    tool := &mcp.Tool{
        Name:        toolDef.Function.Name,
        Description: toolDef.Function.Description,
        InputSchema: convertToJSONSchema(schema),
    }

    // Register with handler closure
    s.mcpServer.AddTool(tool, s.createToolHandler(toolDef.Function.Name))
}
```

**Why This Works**:
- Uses `Server.AddTool()` (not generic `mcp.AddTool()`) for runtime schemas
- Converts API's OpenAI-format schemas to MCP's jsonschema format
- Creates a closure for each tool handler (preserves tool name)

#### 2. Handler Closure Pattern

```go
func (s *Server) createToolHandler(toolName string) mcp.ToolHandler {
    return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // Unmarshal parameters
        var params map[string]interface{}
        json.Unmarshal(req.Params.Arguments, &params)

        // Proxy to API
        result, err := s.apiClient.ExecuteTool(toolName, params)

        // Return formatted result
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: formatResult(result)}},
            StructuredContent: result.Result,
        }, nil
    }
}
```

**Why Closures?**:
- Each handler captures its specific `toolName`
- No need for tool name lookup in handler
- Clean, functional approach
- Minimal overhead (closure is just a pointer)

#### 3. Error Handling Strategy

```go
// API errors become tool errors (not Go errors)
return &mcp.CallToolResult{
    IsError: true,
    Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
}, nil  // Return nil error so MCP doesn't disconnect
```

**Why This Pattern?**:
- Returning Go error would disconnect MCP client
- Setting `IsError: true` signals failure to client
- Client sees error message but connection stays alive
- Allows retry without restart

## Trade-offs and Limitations

### Trade-offs Made

1. **Startup Time vs Runtime Flexibility**
   - ✅ Chose: Fast runtime, slower startup
   - Tools loaded once at startup (~1-2 seconds)
   - Can't add tools without restart
   - Acceptable because tools rarely change

2. **Memory vs Simplicity**
   - ✅ Chose: Simple code, slightly more memory
   - Each tool has its own handler closure
   - ~1KB per tool (negligible for 100s of tools)
   - Total memory: ~10-20MB for full server

3. **Type Safety vs Dynamic Schemas**
   - ✅ Chose: Dynamic schemas, runtime validation
   - Can't use Go's compile-time type checking for tool parameters
   - MCP SDK validates against JSON schemas at runtime
   - More flexible, slightly slower validation

### Known Limitations

1. **Can't Hot-Reload Tools**
   - Tools are fetched once at startup
   - Adding new tools requires server restart
   - **Mitigation**: Restart is fast (~2 seconds), can be automated

2. **No Tool-Specific Middleware**
   - All tools use same authentication (API key + budget key)
   - Can't have per-tool permissions
   - **Mitigation**: API handles permissions, not server

3. **No Streaming Responses**
   - Tool results are returned all at once
   - Long-running tools may timeout
   - **Mitigation**: API should complete tools quickly or use webhooks

4. **No Local Caching**
   - Each tool call hits the API
   - No caching of results
   - **Mitigation**: Could add optional cache layer if needed

### Future Extensibility

**Easy to Add**:
- ✅ Caching layer (add between handler and API client)
- ✅ Metrics/logging (add middleware to handlers)
- ✅ Rate limiting (add to API client)
- ✅ Retry logic (add to API client)
- ✅ Tool filtering (filter API response before registration)

**Hard to Add**:
- ❌ Hot-reload of tools (requires MCP protocol support)
- ❌ Streaming responses (requires API changes)
- ❌ Tool composition (complex, may need different architecture)

## Security Considerations

1. **Credential Management**
   - API keys passed via environment variables
   - Never logged or exposed to clients
   - Not visible in process list

2. **Input Validation**
   - MCP SDK validates parameters against schemas
   - API performs additional validation
   - No direct code execution

3. **Output Sanitization**
   - API responses returned as-is to client
   - No XSS risk (MCP is not web-based)
   - Client responsible for display safety

4. **Network Security**
   - All API calls use HTTPS
   - No certificate pinning (trusts system CA)
   - Standard Go HTTP client security

## Performance Characteristics

**Startup**:
- API fetch: ~500ms - 1.5s (network dependent)
- Tool registration: ~1-10ms per tool
- Total startup: ~1-2 seconds for 100 tools

**Runtime**:
- Tool execution: API latency + ~1-5ms overhead
- Memory: ~10-20MB total
- Concurrent calls: Limited by API rate limits, not server

**Scalability**:
- Can handle 1000s of tools (limited by API, not server)
- Can handle 100s of concurrent tool calls
- Bottleneck is API, not MCP server

## Deployment Strategy

**Binary Distribution**:
```bash
# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o agent-payment-server-linux
GOOS=darwin GOARCH=amd64 go build -o agent-payment-server-mac-intel
GOOS=darwin GOARCH=arm64 go build -o agent-payment-server-mac-arm
GOOS=windows GOARCH=amd64 go build -o agent-payment-server.exe
```

**Size**: 6-8MB per binary (static build)

**Installation**: Download binary, set env vars, run. No dependencies.

## Conclusion

The chosen architecture - **individual dynamic tool registration** using the **official Go MCP SDK** - provides:

1. ✅ Excellent user experience (tools appear natively in Claude/Cursor/VS Code)
2. ✅ Simple, maintainable code (~200 lines for core server)
3. ✅ Small, standalone binaries (6-8MB)
4. ✅ Dynamic tool support (any number, unknown at compile time)
5. ✅ Native MCP integration (uses SDK as intended)
6. ✅ Easy deployment (single binary + env vars)

This architecture successfully meets all project requirements while remaining pragmatic and maintainable.

## References

- [Official Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Specification](https://modelcontextprotocol.io/specification)
- [Go SDK Documentation](https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp)
- [Agent Payment API](https://api.agentpmt.com)
