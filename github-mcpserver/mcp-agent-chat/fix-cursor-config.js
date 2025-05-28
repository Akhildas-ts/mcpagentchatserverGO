import fs from 'fs';
import path from 'path';
import os from 'os';

const cursorConfigPath = path.join(os.homedir(), '.cursor', 'mcp.json');

try {
  const config = JSON.parse(fs.readFileSync(cursorConfigPath, 'utf8'));
  
  // Create a merged configuration
  const mergedConfig = {
    mcpServers: {
      ...config.mcpServers
    }
  };

  // Remove any existing agentchatmcp config
  delete mergedConfig.mcpServers.agentchatmcp;

  // Add the complete configuration
  mergedConfig.mcpServers.agentchatmcp = {
    command: "node",
    args: ["integration.js"],
    cwd: "/Users/akhildas/Desktop/techjays/build-with-ai-github.com/mcpserver/mcp-agent-chat",
    env: {
      GO_SERVER_URL: "http://localhost:8081",
      MCP_SECRET_TOKEN: "1823695f5eee57db3fb25d4b3fcb20b7e812174d0dbc96d2f5eb2f3b360c5e0b",
      MCP_PORT: "3001",
      NODE_ENV: "development",
      DEBUG: "*"
    }
  };

  // Remove the top-level agentchatmcp if it exists
  delete config.agentchatmcp;

  // Write the updated config back
  fs.writeFileSync(cursorConfigPath, JSON.stringify(mergedConfig, null, 2));
  console.log('Successfully fixed Cursor MCP configuration');
  console.log('New configuration:');
  console.log(JSON.stringify(mergedConfig, null, 2));
} catch (error) {
  console.error('Error fixing Cursor config:', error);
} 