# null-mcp

null-mcp allows any MCP-supporting external AI agents to access your financial data, allowing for limitless analytic and other neat use cases. For example, you could ask the agent how much you spent on groceries last month, and it could tell you. Or you could ask it to analyze your spending habits and give you insights on how to save money. 

Of course, since it's an AI that's doing it might just make a mistake or make up an answer, but to be honest, it's been quite reliable in my experience with sonnet-4.6.

> [!IMPORTANT]
> **null-mcp is not secured**. It *must* be deployed locally or behind a VPN. It *must not* be exposed to the public internet.

If your instance of null has multiple users wanting to use mcp, each one needs their own instance of null-mcp, since it's scoped per user via the `NULL_MCP_USER_ID` env.

## usage

in Claude Desktop, you can confgure mcp servers in

settings > developer > edit config

### nix

```json
{
  "mcpServers": {
    "null": {
      "command": "nix",
      "args": [
        "run",
        "nixpkgs#mcp-proxy",
        "--",
        "https://null-mcp.example.com/sse"
      ]
    }
  }
}
```

### npx 

if you have node installed

```json
{
  "mcpServers": {
    "null": {
      "command": "npx",
      "args": [
        "-y",
        "mcp-proxy",
        "https://null-mcp.example.com/sse"
      ]
    }
  }
}
```

### uvx (python/uv)

if you have uv installed

```json
{
  "mcpServers": {
    "null": {
      "command": "uvx",
      "args": [
        "mcp-proxy",
        "https://null-mcp.example.com/sse"
      ]
    }
  }
}
```

### binary

or if you have `mcp-proxy` installed globally, you can just run the binary directly

```json
{
  "mcpServers": {
    "null": {
      "command": "mcp-proxy",
      "args": [
        "https://null-mcp.example.com/sse"
      ]
    }
  }
}
```
