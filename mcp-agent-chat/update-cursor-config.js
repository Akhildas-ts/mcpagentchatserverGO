import fs from 'fs';
import path from 'path';
import os from 'os';

// Read the new config
const newConfig = JSON.parse(fs.readFileSync('./cursor-mcp-config.json', 'utf8'));

// Get Cursor config path
const cursorConfigPath = path.join(os.homedir(), '.cursor', 'mcp.json');

// Read existing config
let cursorConfig = {};
if (fs.existsSync(cursorConfigPath)) {
  cursorConfig = JSON.parse(fs.readFileSync(cursorConfigPath, 'utf8'));
}

// Merge configs
cursorConfig = { ...cursorConfig, ...newConfig };

// Write back to Cursor config
fs.writeFileSync(cursorConfigPath, JSON.stringify(cursorConfig, null, 2));

console.log('Updated Cursor MCP config successfully!'); 