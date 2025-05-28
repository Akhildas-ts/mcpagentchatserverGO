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
