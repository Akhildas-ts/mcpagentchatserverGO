// debug.js
// Save this file in your project and run with: node debug.js
import axios from 'axios';

// Configuration
const GO_SERVER_URL = 'http://localhost:8081';
const NODE_SERVER_URL = 'http://localhost:3000';
const GITHUB_REPO = 'owner/repo-name'; // Replace with your actual repo

// Utility functions
async function testEndpoint(name, url, method, data = null) {
  console.log(`\n${'-'.repeat(50)}`);
  console.log(`Testing: ${name}`);
  console.log(`URL: ${url}`);
  console.log(`Method: ${method}`);
  if (data) {
    console.log(`Data: ${JSON.stringify(data, null, 2)}`);
  }
  
  try {
    let response;
    if (method === 'GET') {
      response = await axios.get(url);
    } else {
      response = await axios.post(url, data);
    }
    
    console.log(`Status: ${response.status}`);
    console.log(`Response: ${JSON.stringify(response.data, null, 2)}`);
    return { success: true, data: response.data };
  } catch (error) {
    console.error(`Error: ${error.message}`);
    if (error.response) {
      console.error(`Status: ${error.response.status}`);
      console.error(`Response: ${JSON.stringify(error.response.data, null, 2)}`);
    }
    return { success: false, error };
  }
}

// Main function
async function runTests() {
  console.log('Starting MCP Server Tests');
  console.log('='.repeat(50));
  
  // Test Go Server health
  await testEndpoint(
    'Go Server Health Check', 
    `${GO_SERVER_URL}/health`, 
    'GET'
  );
  
  // Test Node.js Server health
  await testEndpoint(
    'Node.js Server Health Check', 
    `${NODE_SERVER_URL}/health`, 
    'GET'
  );
  
  // Test GitHub config
  await testEndpoint(
    'GitHub Config', 
    `${GO_SERVER_URL}/github-config`, 
    'POST', 
    {
      github_token: 'your_github_token', // Replace with actual token
      github_username: 'your_github_username' // Replace with actual username
    }
  );
  
  // Test repository indexing
  await testEndpoint(
    'Repository Indexing', 
    `${GO_SERVER_URL}/index-repository`, 
    'POST', 
    {
      repository: GITHUB_REPO
    }
  );
  
  // Test vector search
  await testEndpoint(
    'Vector Search', 
    `${GO_SERVER_URL}/vector-search`, 
    'POST', 
    {
      query: 'vector search implementation',
      repository: GITHUB_REPO,
      limit: 5
    }
  );
  
  // Test chat
  await testEndpoint(
    'Chat', 
    `${GO_SERVER_URL}/chat`, 
    'POST', 
    {
      message: 'How does the vector search work?',
      repository: GITHUB_REPO,
      context: {}
    }
  );
  
  // Test MCP protocol
  await testEndpoint(
    'MCP Protocol - Vector Search', 
    `${NODE_SERVER_URL}/mcp`, 
    'POST', 
    {
      jsonrpc: '2.0',
      method: 'vectorSearch',
      params: {
        query: 'vector search implementation',
        repository: GITHUB_REPO,
        limit: 5
      },
      id: 1
    }
  );
  
  // Test MCP capabilities registration
  await testEndpoint(
    'MCP Capabilities Registration', 
    `${NODE_SERVER_URL}/mcp`, 
    'POST', 
    {
      jsonrpc: '2.0',
      method: 'registerCapabilities',
      params: {},
      id: 2
    }
  );
  
  console.log('\nTests completed!');
}

// Run tests
runTests().catch(err => {
  console.error('Error running tests:', err);
});