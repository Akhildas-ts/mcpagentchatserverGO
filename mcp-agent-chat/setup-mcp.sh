#!/bin/bash

# This script sets up the MCP directory for both HTTP and STDIO communication

# Go to the mcp-agent-chat directory
cd "$(dirname "$0")" || exit 1

echo "Setting up MCP Agent Chat for STDIO communication..."

# Make the index.js file executable
chmod +x index.js
echo "Made index.js executable"

# Install required dependencies if needed
if ! command -v npm &> /dev/null; then
    echo "npm not found, please install Node.js"
    exit 1
fi

echo "Installing dependencies..."
npm install express axios socket.io ws

# Update package.json
cat > package.json << 'EOL'
{
  "name": "mcp-agent-chat",
  "version": "1.0.0",
  "description": "MCP Agent Chat Server with STDIO Support",
  "type": "module",
  "main": "index.js",
  "bin": {
    "mcp-agent-chat": "./index.js"
  },
  "scripts": {
    "start": "node server.js",
    "stdio": "node index.js"
  },
  "dependencies": {
    "axios": "^1.6.2",
    "express": "^4.18.2",
    "socket.io": "^4.7.2",
    "ws": "^8.14.2"
  }
}
EOL
echo "Updated package.json"

# Create a test script to verify STDIO communication
cat > test-stdio.js << 'EOL'
#!/usr/bin/env node

import { spawn } from 'child_process';
import path from 'path';

// Start the MCP handler as a child process
const child = spawn('node', ['index.js'], {
  stdio: ['pipe', 'pipe', 'pipe']
});

// Handle data from the child process
child.stdout.on('data', (data) => {
  const message = data.toString().trim();
  console.log(`Received: ${message}`);
  
  try {
    const response = JSON.parse(message);
    
    // Check if we got server info
    if (response.method === 'serverInfo') {
      console.log('✅ Successfully received server info notification');
      
      // Send a listOfferings request
      sendRequest(child, 'listOfferings', {}, 1);
    }
    
    // Check if we got offerings
    if (response.id === 1 && response.result && response.result.tools) {
      console.log(`✅ Successfully received offerings with ${response.result.tools.length} tools`);
      console.log('Test successful! Your MCP STDIO communication is working!');
      
      // Test complete, exit
      setTimeout(() => {
        child.kill();
        process.exit(0);
      }, 1000);
    }
  } catch (error) {
    console.error('Error parsing response:', error.message);
  }
});

// Handle errors
child.stderr.on('data', (data) => {
  console.error(`Error from MCP: ${data.toString()}`);
});

// Handle process exit
child.on('close', (code) => {
  console.log(`MCP process exited with code ${code}`);
});

// Send a JSON-RPC request to the child process
function sendRequest(child, method, params, id) {
  const request = {
    jsonrpc: '2.0',
    method,
    params,
    id
  };
  
  console.log(`Sending request: ${JSON.stringify(request)}`);
  child.stdin.write(JSON.stringify(request) + '\n');
}

console.log('Testing MCP STDIO communication...');
EOL
chmod +x test-stdio.js
echo "Created test-stdio.js"

echo "Setup complete!"
echo ""
echo "To test STDIO communication, run: node test-stdio.js"
echo "To start the HTTP server, run: node server.js"
echo ""
echo "Your MCP client should now work with: node /path/to/mcp-agent-chat"