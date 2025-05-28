// MCP Diagnostic Utility
// This script checks your MCP and Go server setup for issues

import axios from 'axios';
import http from 'http';
import { exec } from 'child_process';
import { promisify } from 'util';
import path from 'path';
import fs from 'fs';
import 'dotenv/config';

const execAsync = promisify(exec);

// Configuration
const config = {
  GO_SERVER_URL: process.env.GO_SERVER_URL || 'http://localhost:8081',
  MCP_PORT: parseInt(process.env.MCP_PORT || '3000', 10)
};

async function checkGoServer() {
  console.log('\n🔍 Checking Go Vector Search Server...');
  
  try {
    const response = await axios.get(`${config.GO_SERVER_URL}/health`, { 
      timeout: 5000,
      headers: { 'User-Agent': 'MCP-Diagnostic/1.0' }
    });
    console.log('✅ Go server is running!');
    console.log(`   Response: ${JSON.stringify(response.data)}`);
    return true;
  } catch (error) {
    console.error('❌ Go server is not reachable');
    if (error.code === 'ECONNREFUSED') {
      console.error(`   Connection refused at ${config.GO_SERVER_URL}`);
      console.error('   Make sure your Go server is running on the correct port');
    } else if (error.response) {
      console.error(`   Server responded with status ${error.response.status}`);
      console.error(`   Message: ${error.response.data}`);
    } else {
      console.error(`   Error: ${error.message}`);
    }
    return false;
  }
}

async function checkMcpHttpServer() {
  console.log('\n🔍 Checking MCP HTTP Server...');
  
  try {
    const response = await axios.get(`http://localhost:${config.MCP_PORT}/mcp`, { 
      timeout: 5000,
      headers: { 'User-Agent': 'MCP-Diagnostic/1.0' }
    });
    console.log('✅ MCP HTTP server is running!');
    console.log(`   Response status: ${response.status}`);
    return true;
  } catch (error) {
    if (error.response && error.response.status === 404) {
      console.log('✅ MCP HTTP server is running (404 is expected for GET /mcp)');
      return true;
    } else if (error.code === 'ECONNREFUSED') {
      console.error('❌ MCP HTTP server is not reachable');
      console.error(`   Connection refused at http://localhost:${config.MCP_PORT}`);
      console.error('   Make sure your Node.js MCP server is running on the correct port');
    } else {
      console.error('❌ MCP HTTP server check failed');
      console.error(`   Error: ${error.message}`);
    }
    return false;
  }
}

async function checkNodeJsVersion() {
  console.log('\n🔍 Checking Node.js environment...');
  
  try {
    const { stdout: nodeVersion } = await execAsync('node --version');
    console.log(`✅ Node.js version: ${nodeVersion.trim()}`);
    
    // Check npm packages
    if (fs.existsSync('package.json')) {
      const packageJson = JSON.parse(fs.readFileSync('package.json', 'utf8'));
      console.log(`✅ Package name: ${packageJson.name}, version: ${packageJson.version}`);
      
      // Check MCP SDK version
      const mcpSdkVersion = packageJson.dependencies["@modelcontextprotocol/sdk"];
      if (mcpSdkVersion) {
        console.log(`✅ MCP SDK version: ${mcpSdkVersion}`);
      } else {
        console.error('❌ MCP SDK not found in dependencies');
      }
    } else {
      console.error('❌ package.json not found in current directory');
    }
    
    return true;
  } catch (error) {
    console.error('❌ Failed to check Node.js environment');
    console.error(`   Error: ${error.message}`);
    return false;
  }
}

async function checkNetwork() {
  console.log('\n🔍 Checking network connectivity...');
  
  // Check local loopback
  try {
    await axios.get('http://localhost:8080', { 
      timeout: 2000,
      validateStatus: () => true // Accept any status code
    });
    console.log('✅ Local loopback (localhost) is working');
  } catch (error) {
    if (error.code === 'ECONNREFUSED') {
      console.log('✅ Local loopback (localhost) is working (connection refused, but reachable)');
    } else {
      console.error('❌ Local loopback (localhost) check failed');
      console.error(`   Error: ${error.message}`);
    }
  }
  
  // Check if ports are in use
  try {
    const { stdout: portsInUse } = await execAsync('netstat -an | grep LISTEN');
    console.log('✅ Ports currently in use:');
    
    // Extract the port numbers
    const goServerPort = new URL(config.GO_SERVER_URL).port || '8081';
    const goServerRegex = new RegExp(`LISTEN.*:${goServerPort}\\b`);
    const mcpServerRegex = new RegExp(`LISTEN.*:${config.MCP_PORT}\\b`);
    
    const goServerRunning = goServerRegex.test(portsInUse);
    const mcpServerRunning = mcpServerRegex.test(portsInUse);
    
    if (goServerRunning) {
      console.log(`   ✅ Go server port ${goServerPort} is active`);
    } else {
      console.error(`   ❌ Go server port ${goServerPort} does not appear to be active`);
    }
    
    if (mcpServerRunning) {
      console.log(`   ✅ MCP server port ${config.MCP_PORT} is active`);
    } else {
      console.error(`   ❌ MCP server port ${config.MCP_PORT} does not appear to be active`);
    }
  } catch (error) {
    console.error('❌ Failed to check ports in use');
    console.error(`   Error: ${error.message}`);
  }
}

async function checkServerProcesses() {
  console.log('\n🔍 Checking for running server processes...');
  
  try {
    let { stdout: processes } = await execAsync('ps aux | grep -E "node|go run" | grep -v grep');
    console.log('✅ Found running processes that might be related to MCP or Go servers:');
    
    // Format the output to be more readable
    processes = processes.split('\n')
      .filter(line => line.trim())
      .map(line => {
        const parts = line.split(/\s+/);
        const pid = parts[1];
        const cmd = parts.slice(10).join(' ');
        return `   PID ${pid}: ${cmd}`;
      })
      .join('\n');
    
    console.log(processes || '   No matching processes found');
  } catch (error) {
    if (error.stderr.includes('no matches found')) {
      console.log('   No matching server processes found running');
    } else {
      console.error('❌ Failed to check server processes');
      console.error(`   Error: ${error.message}`);
    }
  }
}

async function main() {
  console.log('🔧 MCP Server Diagnostic Tool 🔧');
  console.log('================================');
  console.log(`Current configuration:
  - Go server URL: ${config.GO_SERVER_URL}
  - MCP HTTP port: ${config.MCP_PORT}
  `);
  
  await checkNodeJsVersion();
  await checkGoServer();
  await checkMcpHttpServer();
  await checkNetwork();
  await checkServerProcesses();
  
  console.log('\n✨ Diagnostic Summary ✨');
  console.log('======================');
  console.log('If you see any ❌ errors above, those are the areas to investigate.');
  console.log('\nRecommended next steps:');
  console.log('1. Make sure your Go server is running first');
  console.log('2. Start the enhanced MCP integration script: node enhanced-integration.js');
  console.log('3. Test with the diagnostic client: node test-client.js');
  console.log('\nFor more detailed logs, set the DEBUG environment variable:');
  console.log('   DEBUG=mcp:* node enhanced-integration.js');
}

main().catch(error => {
  console.error('Fatal error during diagnostic:', error);
});