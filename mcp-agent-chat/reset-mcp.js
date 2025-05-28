import fs from 'fs';
import path from 'path';
import os from 'os';

const cursorConfigPath = path.join(os.homedir(), '.cursor', 'mcp.json');

// Read existing config
const config = JSON.parse(fs.readFileSync(cursorConfigPath, 'utf8'));

// Remove the agentchatmcp configuration
if (config.mcpServers && config.mcpServers.agentchatmcp) {
    delete config.mcpServers.agentchatmcp;
    console.log('Removed agentchatmcp configuration');
}

// Write back the configuration
fs.writeFileSync(cursorConfigPath, JSON.stringify(config, null, 2));
console.log('Updated MCP configuration'); 