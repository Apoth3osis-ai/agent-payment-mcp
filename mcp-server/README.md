# Agent Payment MCP Server

A Model Context Protocol (MCP) server that dynamically fetches tools from the Agent Payment API and proxies tool execution requests back to the API.

## Overview

This MCP server enables Claude Desktop, Cursor, VS Code, and other MCP clients to access dynamically-loaded tools from the Agent Payment platform. Tools are fetched at runtime, allowing the server to support any number of tools without recompilation.

## Features

- **Dynamic Tool Loading**: Fetches available tools from REST API at startup
- **Automatic Tool Registration**: Registers all tools with MCP server automatically
- **Proxy Execution**: Routes tool calls to Agent Payment API
- **Native MCP Integration**: Works with any MCP-compatible client
- **Simple Configuration**: Just set environment variables and run
- **Graceful Shutdown**: Handles SIGINT/SIGTERM signals properly

## Architecture

### Components

1. **MCP Server** (`internal/mcp/server.go`)
   - Creates MCP server using official Go SDK
   - Fetches tools from Agent Payment API at startup
   - Registers each tool dynamically with appropriate schemas
   - Routes tool execution requests to API client

2. **API Client** (`internal/api/client.go`)
   - Handles HTTP communication with Agent Payment API
   - Fetches tool definitions (`GET /products/fetch`)
   - Executes tools (`POST /products/purchase`)
   - Manages authentication headers

3. **Main Entry Point** (`cmd/agent-payment-server/main.go`)
   - Reads configuration from environment variables
   - Initializes server
   - Sets up signal handling
   - Runs MCP server on stdio transport

### How It Works

```
┌─────────────────────┐
│  Claude Desktop /   │
│  Cursor / VS Code   │
└──────────┬──────────┘
           │ MCP Protocol (stdio)
           ▼
┌─────────────────────┐
│   MCP Server (Go)   │
│  ┌───────────────┐  │
│  │ Tool Registry │  │ ← Fetched at startup
│  └───────────────┘  │
│  ┌───────────────┐  │
│  │ Tool Handlers │  │ ← Proxy to API
│  └───────────────┘  │
└──────────┬──────────┘
           │ HTTPS
           ▼
┌─────────────────────┐
│ Agent Payment API   │
│  - Fetch Tools      │
│  - Execute Tools    │
└─────────────────────┘
```

## Installation

### Prerequisites

- Go 1.23 or later
- Agent Payment API key and budget key

### Build from Source

```bash
cd mcp-server
go mod download
go build -o agent-payment-server ./cmd/agent-payment-server
```

### Install Binary

```bash
go install github.com/agentpmt/agent-payment-mcp-server/cmd/agent-payment-server@latest
```

## Configuration

The server requires two environment variables:

- `AGENT_PAYMENT_API_KEY`: Your Agent Payment API key
- `AGENT_PAYMENT_BUDGET_KEY`: Your Agent Payment budget key

Set these before running the server:

```bash
export AGENT_PAYMENT_API_KEY=your-api-key
export AGENT_PAYMENT_BUDGET_KEY=your-budget-key
```

## Usage

### Running the Server

```bash
./agent-payment-server
```

The server will:
1. Fetch available tools from the API
2. Register all tools with the MCP server
3. Start listening on stdio for MCP client connections

### Claude Desktop Integration

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "agent-payment": {
      "command": "/path/to/agent-payment-server",
      "env": {
        "AGENT_PAYMENT_API_KEY": "your-api-key",
        "AGENT_PAYMENT_BUDGET_KEY": "your-budget-key"
      }
    }
  }
}
```

Restart Claude Desktop, and all Agent Payment tools will be available.

### Cursor Integration

Add to Cursor's MCP settings:

```json
{
  "mcpServers": {
    "agent-payment": {
      "command": "/path/to/agent-payment-server",
      "env": {
        "AGENT_PAYMENT_API_KEY": "your-api-key",
        "AGENT_PAYMENT_BUDGET_KEY": "your-budget-key"
      }
    }
  }
}
```

### VS Code Integration

Install the MCP extension and configure similarly.

## User Experience

### Discovering Tools

When you open Claude Desktop with this server configured:

1. **Ask "What tools do I have?"**
   - Claude will list all available tools
   - Each tool appears with its name and description
   - Tools are indistinguishable from built-in tools

2. **Browse in Tools UI**
   - Each tool appears individually in the tools list
   - Full descriptions and parameter schemas are visible
   - You can see what each tool does before using it

### Using Tools

**Example conversation:**

```
User: Check the weather in New York

Claude: I'll use the weather_check tool to get that information.
[Calls: weather_check with city="New York"]

Tool Result: Sunny, 72°F
Cost: $0.01
Remaining Balance: $9.99

Claude: The weather in New York is currently sunny with a temperature of 72°F.
```

The experience is seamless - users don't need to know they're using a proxy server.

## Testing

### 1. Compile Test

```bash
cd mcp-server
go mod tidy
go build ./cmd/agent-payment-server
```

Should compile with no errors.

### 2. Startup Test

```bash
export AGENT_PAYMENT_API_KEY=your-api-key
export AGENT_PAYMENT_BUDGET_KEY=your-budget-key
./agent-payment-server
```

Expected output:
```
Fetching tools from Agent Payment API...
Fetched 50 tools from API
Successfully registered 50 tools
Starting MCP server on stdio transport...
```

### 3. MCP Inspector Test

```bash
npx @modelcontextprotocol/inspector ./agent-payment-server
```

Opens a web UI where you can:
- View all registered tools
- See tool schemas
- Test tool calls interactively

### 4. Integration Test

Configure in Claude Desktop and test:
```
User: What tools are available?
[Claude lists all Agent Payment tools]

User: Use the [tool_name] tool
[Tool executes successfully]
```

## API Integration

### Endpoints

**GET /products/fetch**
- Fetches available tools
- Headers: `x-api-key`, `x-budget-key`
- Query params: `page`, `page_size`
- Returns: List of tool definitions with schemas

**POST /products/purchase**
- Executes a tool
- Headers: `x-api-key`, `x-budget-key`
- Body: `product_id`, `parameters`
- Returns: Execution result, cost, balance

### Tool Definition Format

Tools from the API are expected in this format:

```json
{
  "type": "function",
  "function": {
    "name": "tool_name",
    "description": "Tool description",
    "parameters": {
      "type": "object",
      "properties": {
        "param1": {
          "type": "string",
          "description": "Parameter description"
        }
      },
      "required": ["param1"]
    }
  }
}
```

This is the standard OpenAI function calling format, which maps directly to MCP tool schemas.

## Development

### Project Structure

```
mcp-server/
├── cmd/
│   └── agent-payment-server/
│       └── main.go              # Entry point
├── internal/
│   ├── api/
│   │   └── client.go            # API client
│   └── mcp/
│       └── server.go            # MCP server
├── go.mod
├── go.sum
└── README.md
```

### Adding Features

To modify tool handling:
1. Edit `internal/mcp/server.go`
2. Modify `createToolHandler()` function
3. Rebuild and test

To change API integration:
1. Edit `internal/api/client.go`
2. Update request/response types
3. Rebuild and test

### Error Handling

The server handles errors gracefully:
- API errors are returned to the client with clear messages
- Failed tool registrations are logged but don't stop startup
- Network errors include retry information
- All errors preserve context for debugging

## Troubleshooting

### "Failed to fetch tools"

- Check API key and budget key are correct
- Verify network connectivity to api.agentpmt.com
- Check API endpoint is accessible

### "No tools registered"

- API may have returned empty tool list
- Check API response format matches expected structure
- Enable debug logging to see raw API responses

### "Tool execution failed"

- Check tool parameters match schema
- Verify budget has sufficient balance
- Check API logs for execution errors

## Performance

- Startup time: ~1-2 seconds (depends on API response time)
- Tool execution: ~100-500ms (depends on API latency)
- Memory usage: ~10-20MB
- Binary size: ~6-8MB (static build)

## Security

- API keys are passed via environment variables (not command line)
- No sensitive data is logged
- All API communication uses HTTPS
- Credentials are never exposed to MCP clients

## License

MIT

## Support

For issues or questions:
- GitHub Issues: [repository URL]
- Documentation: [docs URL]
- Email: support@agentpmt.com
