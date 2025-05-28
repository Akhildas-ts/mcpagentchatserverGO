# Go Vector Search Server for MCP

This is the Go backend server that provides vector search capabilities for the Agent Chat MCP (Model Context Protocol) implementation. This server works in conjunction with the Node.js MCP server to provide AI-powered code search and chat functionality.

## Overview

This Go server provides the vector search backend infrastructure that:
- Handles code embedding and vector storage
- Processes search queries
- Returns relevant code snippets based on vector similarity

## Prerequisites

- Go 1.19 or higher
- PostgreSQL (for vector storage)
- Node.js (for the MCP server integration)

## Setup

1. Install Go dependencies:
```bash
go mod tidy
```

2. Configure environment variables:
```bash
export DB_CONNECTION="postgresql://user:password@localhost:5432/dbname"
export SERVER_PORT=8081
```

## Running the Server

Start the Go server:
```bash
go run main.go
```

The server will start on `http://localhost:8081` by default.

## API Endpoints

- `POST /search` - Vector search endpoint
  - Request body:
    ```json
    {
      "query": "search query string",
      "repository": "repository name",
      "limit": 10
    }
    ```

## Integration with MCP Server

This Go server is designed to work with the Node.js MCP server. To set up the complete system:

1. Start this Go server first
2. Start the Node.js MCP server with:
   ```bash
   npm run integration
   ```
3. The Node.js server will connect to this Go server at `http://localhost:8081`

## Architecture
