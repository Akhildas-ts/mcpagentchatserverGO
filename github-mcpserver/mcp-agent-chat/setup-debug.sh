#!/bin/bash
# setup-debug.sh - Script to set up the debug environment

echo "Setting up MCP Debug Environment..."

# Backup existing files
mkdir -p backup
cp index.js backup/index.js.bak 2>/dev/null || echo "No index.js to backup"
cp package.json backup/package.json.bak 2>/dev/null || echo "No package.json to backup"

# Create index.js that forwards to our debug tracer
cat > index.js << 'EOL'
#!/usr/bin/env node
require('./debug-tracer.js');
EOL

# Make it executable
chmod +x index.js

# Create a minimal package.json
cat > package.json << 'EOL'
{
  "name": "mcp-agent-chat",
  "version": "1.0.0",
  "description": "MCP Agent with Debug Tracer",
  "main": "index.js"
}
EOL

echo "Setup complete! The debug tracer will generate a detailed log file at mcp-full-debug.log"
echo "Restart your client application to see the debug output."