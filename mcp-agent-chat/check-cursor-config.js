import fs from 'fs';
import path from 'path';
import os from 'os';

const cursorConfigPath = path.join(os.homedir(), '.cursor', 'mcp.json');

try {
  const config = JSON.parse(fs.readFileSync(cursorConfigPath, 'utf8'));
  console.log('Current Cursor MCP Configuration:');
  console.log(JSON.stringify(config, null, 2));
} catch (error) {
  console.error('Error reading Cursor config:', error);
} 