# Agent Chat MCP Server

This is an MCP (Model Context Protocol) server implementation for agent chat that integrates with your existing vector search Go server.

## Setup

1. Install dependencies:
   ```
   npm install
   ```

## Available Scripts

- Start the standalone MCP server:
  ```
  npm start
  ```

- Run the test client:
  ```
  npm test
  ```

- Start the integration server (connects to your Go server):
  ```
  npm run integration
  ```

## Configuration

For the integration server, you can set these environment variables:
- `GO_SERVER_URL`: The URL of your Go vector search server (default: http://localhost:8081)
- `MCP_PORT`: The port for the HTTP MCP server (default: 3000)

## Usage in Cursor

To use this MCP server in Cursor or other AI tools that support MCP:

1. Start the integration server:
   ```
   npm run integration
   ```

2. Configure the MCP client (in Cursor or other tools) to connect to:
   ```
   http://localhost:3000
   ```

## Available Tools

1. `chat` - Process a chat message with repository context
   - Parameters:
     - `message`: The user message to process (string, required)
     - `repository`: The GitHub repository to reference (string, optional)
     - `context`: Additional context for the chat (object, optional)

2. `vectorSearch` - Search for code in a repository
   - Parameters:
     - `query`: The search query (string, required)
     - `repository`: The repository to search in (string, required)
     - `limit`: Maximum number of results to return (number, optional)

## Integration with Go Server

The integration script (`integration.js`) connects this MCP server to your existing Go vector search server, allowing you to leverage your existing functionality while exposing it through the MCP protocol. 