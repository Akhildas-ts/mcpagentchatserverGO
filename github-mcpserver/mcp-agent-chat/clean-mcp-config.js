import fs from 'fs';
import path from 'path';
import os from 'os';

const cursorConfigPath = path.join(os.homedir(), '.cursor', 'mcp.json');

// Create a clean, minimal configuration
const cleanConfig = {
  "mcpServers": {
    "agentchatmcp": {
      "command": "node",
      "args": ["integration.js"],
      "cwd": "/Users/akhildas/Desktop/techjays/build-with-ai-github.com/mcpserver/mcp-agent-chat",
      "env": {
        "GO_SERVER_URL": "http://localhost:8081",
        "MCP_SECRET_TOKEN": "1823695f5eee57db3fb25d4b3fcb20b7e812174d0dbc96d2f5eb2f3b360c5e0b",
        "MCP_PORT": "3001",
        "NODE_ENV": "development"
      }
    }
  }
};

try {
  // Write the clean configuration
  fs.writeFileSync(cursorConfigPath, JSON.stringify(cleanConfig, null, 2));
  console.log('Successfully wrote clean MCP configuration');
  console.log('New configuration:');
  console.log(JSON.stringify(cleanConfig, null, 2));
} catch (error) {
  console.error('Error writing Cursor config:', error);
} 