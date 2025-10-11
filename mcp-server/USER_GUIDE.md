# Agent Payment MCP Server - User Guide

A comprehensive guide for end users on how to install, configure, and use the Agent Payment MCP Server with Claude Desktop, Cursor, and other MCP clients.

## What is This?

The Agent Payment MCP Server connects your MCP-compatible AI assistants (like Claude Desktop or Cursor) to the Agent Payment platform, giving them access to 50+ powerful tools.

**What you can do**:
- Access specialized tools from within Claude Desktop or Cursor
- Execute tools directly in your AI conversations
- Pay-as-you-go for tool usage
- Discover new tools automatically as they're added to the platform

## Installation

### Step 1: Download the Server

Download the appropriate binary for your operating system:

**macOS (Apple Silicon - M1/M2/M3)**:
```bash
curl -L -o agent-payment-server https://github.com/agentpmt/mcp-server/releases/latest/download/agent-payment-server-darwin-arm64
chmod +x agent-payment-server
```

**macOS (Intel)**:
```bash
curl -L -o agent-payment-server https://github.com/agentpmt/mcp-server/releases/latest/download/agent-payment-server-darwin-amd64
chmod +x agent-payment-server
```

**Linux**:
```bash
curl -L -o agent-payment-server https://github.com/agentpmt/mcp-server/releases/latest/download/agent-payment-server-linux-amd64
chmod +x agent-payment-server
```

**Windows**:
Download from: https://github.com/agentpmt/mcp-server/releases/latest/download/agent-payment-server-windows-amd64.exe

### Step 2: Move to a Permanent Location

**macOS/Linux**:
```bash
sudo mv agent-payment-server /usr/local/bin/
```

**Windows**:
Move `agent-payment-server.exe` to `C:\Program Files\AgentPayment\`

### Step 3: Get Your API Keys

1. Visit https://agentpmt.com
2. Sign up or log in
3. Go to Settings → API Keys
4. Copy your API Key and Budget Key

## Configuration

### Claude Desktop

**macOS**:
Edit `~/Library/Application Support/Claude/claude_desktop_config.json`

**Windows**:
Edit `%APPDATA%\Claude\claude_desktop_config.json`

**Linux**:
Edit `~/.config/Claude/claude_desktop_config.json`

Add this configuration:

```json
{
  "mcpServers": {
    "agent-payment": {
      "command": "/usr/local/bin/agent-payment-server",
      "env": {
        "AGENT_PAYMENT_API_KEY": "your-api-key-here",
        "AGENT_PAYMENT_BUDGET_KEY": "your-budget-key-here"
      }
    }
  }
}
```

**Windows**: Use full path like `C:\\Program Files\\AgentPayment\\agent-payment-server.exe`

**Restart Claude Desktop** for changes to take effect.

### Cursor

1. Open Cursor
2. Go to Settings (Cmd+, on Mac, Ctrl+, on Windows)
3. Search for "MCP"
4. Click "Edit MCP Settings"
5. Add this configuration:

```json
{
  "mcpServers": {
    "agent-payment": {
      "command": "/usr/local/bin/agent-payment-server",
      "env": {
        "AGENT_PAYMENT_API_KEY": "your-api-key-here",
        "AGENT_PAYMENT_BUDGET_KEY": "your-budget-key-here"
      }
    }
  }
}
```

**Restart Cursor** for changes to take effect.

### VS Code (with MCP Extension)

1. Install the MCP extension
2. Open VS Code settings (JSON)
3. Add to `mcp.servers`:

```json
{
  "mcp.servers": {
    "agent-payment": {
      "command": "/usr/local/bin/agent-payment-server",
      "env": {
        "AGENT_PAYMENT_API_KEY": "your-api-key-here",
        "AGENT_PAYMENT_BUDGET_KEY": "your-budget-key-here"
      }
    }
  }
}
```

**Restart VS Code** for changes to take effect.

## Usage

### Discovering Available Tools

Once configured, ask your AI assistant:

```
What tools do I have available?
```

Claude/Cursor will respond with a list of all Agent Payment tools, including:
- Tool names
- Descriptions
- What they do

### Using a Tool

Simply describe what you want to do in natural language:

**Example 1: Weather**
```
User: What's the weather like in New York?

Claude: I'll check the weather in New York for you.
[Uses weather_check tool]
The weather in New York is currently sunny with a temperature of 72°F.
```

**Example 2: Stock Price**
```
User: What's the current price of Apple stock?

Claude: I'll look up the current Apple stock price.
[Uses stock_price tool]
Apple (AAPL) is currently trading at $182.45.
```

**Example 3: Currency Conversion**
```
User: Convert 100 USD to EUR

Claude: I'll convert that for you.
[Uses currency_convert tool]
100 USD = 92.34 EUR at the current exchange rate.
```

### Browsing Tools in the UI

**Claude Desktop**:
1. Click the hammer icon (🔨) in the chat interface
2. Browse available tools
3. See descriptions and parameters
4. Click to insert tool calls into your conversation

**Cursor**:
1. Open the AI chat
2. Type "@" to see available tools
3. Select a tool to see its description
4. Continue typing to provide parameters

### Understanding Tool Results

Tool results include:
- **Result**: The actual output from the tool
- **Cost**: How much the tool execution cost (e.g., $0.01)
- **Balance**: Your remaining Agent Payment balance (e.g., $9.99)

Example:
```
Result: Weather is sunny, 72°F
Cost: $0.01
Remaining Balance: $9.99
```

### Managing Your Budget

**Check Balance**:
```
User: What's my Agent Payment balance?

Claude: [Uses balance_check tool]
Your current balance is $9.99
```

**Add Funds**:
Visit https://agentpmt.com/billing to add funds to your account.

**Set Budget Alerts**:
Configure alerts at https://agentpmt.com/settings to be notified when your balance runs low.

## Common Scenarios

### Scenario 1: Research Assistant

```
User: Research the latest AI developments and summarize them

Claude: I'll search for recent AI news and summarize the findings.
[Uses web_search tool]
[Uses summarize tool]

Here are the latest AI developments:
1. New GPT-5 announcement...
2. Breakthrough in robotics...
...
```

### Scenario 2: Data Analysis

```
User: Analyze this CSV file and show me trends

Claude: I'll analyze the data for you.
[Uses csv_analyzer tool]
[Uses chart_generator tool]

The data shows an upward trend of 15% over the past quarter...
[Shows generated chart]
```

### Scenario 3: Code Generation with Tools

```
User: Generate a Python script that fetches stock prices

Claude: I'll create that script for you.
[Uses code_generator tool]
[Uses test_generator tool]

Here's the script:
```python
# Generated code...
```
```

## Troubleshooting

### "No tools available"

**Problem**: Claude/Cursor doesn't show any Agent Payment tools.

**Solutions**:
1. Check that you restarted the application after configuration
2. Verify your API keys are correct in the config file
3. Check logs for connection errors

**Check Logs**:

**macOS (Claude Desktop)**:
```bash
tail -f ~/Library/Logs/Claude/mcp*.log
```

**Linux (Claude Desktop)**:
```bash
tail -f ~/.config/Claude/logs/mcp*.log
```

Look for error messages about Agent Payment server.

### "Tool execution failed"

**Problem**: Tool calls fail with errors.

**Causes**:
- Insufficient balance
- Invalid parameters
- Network issues
- API downtime

**Solutions**:
1. Check your balance at https://agentpmt.com
2. Verify parameters match tool requirements
3. Check internet connection
4. Visit https://status.agentpmt.com for API status

### "Server disconnected"

**Problem**: Agent Payment server keeps disconnecting.

**Solutions**:
1. Verify binary has execute permissions
2. Check API keys are valid
3. Ensure network connectivity
4. Update to latest version

### "Invalid API key"

**Problem**: Error about invalid API credentials.

**Solutions**:
1. Verify API key is correct (no extra spaces)
2. Regenerate keys at https://agentpmt.com/settings
3. Check key hasn't been revoked
4. Ensure both API key and budget key are set

## Tips and Best Practices

### Tip 1: Ask Before Expensive Operations

Some tools may cost more than others. Ask Claude:
```
How much will this operation cost before you run it?
```

### Tip 2: Batch Operations

If doing multiple similar tasks:
```
Use the X tool to process all these items in one batch
```

This may be more efficient and cost-effective.

### Tip 3: Explore New Tools

New tools are added regularly. Periodically ask:
```
What new tools have been added to Agent Payment?
```

### Tip 4: Check Tool Documentation

For detailed information about a specific tool:
```
Show me the documentation for the [tool_name] tool
```

### Tip 5: Set Budget Limits

Configure budget limits in your Agent Payment account to avoid unexpected charges.

## Privacy and Security

### What Data is Shared?

When you use a tool:
- **Sent to Agent Payment**: Tool parameters (e.g., city name for weather)
- **NOT sent**: Your entire conversation history
- **NOT sent**: Other files or context in your chat

### API Key Security

- **Never share** your API keys
- **Rotate keys** periodically at https://agentpmt.com/settings
- **Revoke keys** immediately if compromised
- **Use budget keys** to limit maximum spend

### Data Retention

Agent Payment retains:
- Tool execution logs (30 days)
- Billing records (7 years)
- Usage analytics (90 days)

See full privacy policy at https://agentpmt.com/privacy

## Updating

### Check for Updates

```bash
agent-payment-server --version
```

Compare with latest version at https://github.com/agentpmt/mcp-server/releases

### Update Process

1. Download new binary (same as installation)
2. Replace old binary
3. Restart Claude/Cursor
4. Verify tools still work

No configuration changes needed for updates.

## Uninstalling

### Remove Server Binary

**macOS/Linux**:
```bash
sudo rm /usr/local/bin/agent-payment-server
```

**Windows**:
Delete `C:\Program Files\AgentPayment\agent-payment-server.exe`

### Remove Configuration

**Claude Desktop**:
Edit `claude_desktop_config.json` and remove the `agent-payment` entry.

**Cursor**:
Edit MCP settings and remove the `agent-payment` entry.

### Delete Account Data

Visit https://agentpmt.com/settings → Delete Account

This removes all your data from Agent Payment servers.

## Getting Help

### Documentation

- **MCP Server**: https://github.com/agentpmt/mcp-server
- **Agent Payment**: https://docs.agentpmt.com
- **MCP Protocol**: https://modelcontextprotocol.io

### Support

- **Email**: support@agentpmt.com
- **Discord**: https://discord.gg/agentpmt
- **Issues**: https://github.com/agentpmt/mcp-server/issues

### Community

- **Forum**: https://community.agentpmt.com
- **Examples**: https://github.com/agentpmt/examples
- **Templates**: https://github.com/agentpmt/templates

## FAQ

**Q: How much do tools cost?**
A: Pricing varies by tool. Most tools cost $0.01-$0.10 per use. Check https://agentpmt.com/pricing for details.

**Q: Can I use this with other AI assistants?**
A: Yes! Any MCP-compatible client works. Tested with Claude Desktop, Cursor, and VS Code.

**Q: Are there free tools?**
A: Some tools offer free tiers or trials. Check individual tool pages at https://agentpmt.com/tools

**Q: Can I build my own tools?**
A: Yes! Visit https://developers.agentpmt.com to create custom tools.

**Q: Is my data private?**
A: Yes. We only process data needed for tool execution. See https://agentpmt.com/privacy

**Q: What if a tool fails?**
A: Failed executions are not charged. Contact support for persistent issues.

**Q: Can I use this commercially?**
A: Yes. Commercial use is permitted. See https://agentpmt.com/terms

**Q: How do I report a bug?**
A: Open an issue at https://github.com/agentpmt/mcp-server/issues

**Q: Can I contribute?**
A: Yes! PRs welcome at https://github.com/agentpmt/mcp-server

**Q: Is source code available?**
A: Yes, open source under MIT license at https://github.com/agentpmt/mcp-server

## Example Workflows

### Workflow 1: Daily Briefing

```
User: Give me my daily briefing

Claude: I'll gather information for your briefing.
[Uses weather_check for current location]
[Uses news_summary for top headlines]
[Uses calendar_check for today's events]
[Uses stock_check for portfolio]

Here's your daily briefing:
- Weather: Sunny, 72°F
- Top News: [headlines]
- Today's Schedule: [events]
- Portfolio: +1.2% today
```

### Workflow 2: Content Creation

```
User: Help me write a blog post about AI ethics

Claude: I'll help with that.
[Uses research_topic tool]
[Uses outline_generator tool]
[Uses content_writer tool]
[Uses seo_optimizer tool]

I've created a comprehensive blog post on AI ethics...
```

### Workflow 3: Data Processing

```
User: Process these 100 customer reviews and extract sentiment

Claude: I'll analyze those reviews for you.
[Uses batch_sentiment_analyzer tool]
[Uses data_visualizer tool]

Analysis complete:
- Positive: 65%
- Neutral: 20%
- Negative: 15%
[Shows visualization]
```

## Advanced Usage

### Custom Tool Configurations

Some tools support custom configurations. Set via environment variables:

```json
{
  "mcpServers": {
    "agent-payment": {
      "command": "/usr/local/bin/agent-payment-server",
      "env": {
        "AGENT_PAYMENT_API_KEY": "your-api-key",
        "AGENT_PAYMENT_BUDGET_KEY": "your-budget-key",
        "AGENT_PAYMENT_TIMEOUT": "30",
        "AGENT_PAYMENT_RETRY": "3"
      }
    }
  }
}
```

### Tool Filtering

If you only want certain tools, set:

```json
"env": {
  "AGENT_PAYMENT_TOOLS_FILTER": "weather,stock,news"
}
```

Only specified tools will be loaded.

### Debug Mode

Enable detailed logging:

```json
"env": {
  "AGENT_PAYMENT_DEBUG": "true"
}
```

Logs will show detailed tool execution information.

---

**Happy tool using!** 🛠️

For questions or feedback, contact support@agentpmt.com
