// improved-test-client.js in mcp-agent-chat directory
import axios from 'axios';

const MCP_SERVER_URL = 'http://localhost:3000';

// Configure axios to provide more detailed errors
axios.defaults.timeout = 10000; // 10 seconds timeout

async function testConnection() {
  console.log('Testing MCP server connection...');
  console.log(`Target server: ${MCP_SERVER_URL}`);
  
  // Test if server is reachable at all
  try {
    console.log('\n1. Testing basic server connectivity...');
    const healthEndpoint = `${MCP_SERVER_URL}/health`;
    console.log(`   Requesting: ${healthEndpoint}`);
    
    const healthResponse = await axios.get(healthEndpoint);
    console.log('   ✅ Health check successful');
    console.log('   Response:', healthResponse.data);
  } catch (error) {
    console.error('   ❌ Health check failed');
    logDetailedError(error, 'health');
    console.log('\nServer connectivity test failed. Please ensure the MCP server is running on port 3000.');
    return;
  }
  
  // Test offerings endpoint
  try {
    console.log('\n2. Testing offerings endpoint...');
    const offeringsEndpoint = `${MCP_SERVER_URL}/offerings`;
    console.log(`   Requesting: ${offeringsEndpoint}`);
    
    const offeringsResponse = await axios.post(offeringsEndpoint);
    console.log('   ✅ Offerings check successful');
    console.log('   Tools found:', offeringsResponse.data.tools.length);
    console.log('   Tool IDs:', offeringsResponse.data.tools.map(t => t.id).join(', '));
  } catch (error) {
    console.error('   ❌ Offerings check failed');
    logDetailedError(error, 'offerings');
    // Continue testing other endpoints
  }
  
  // Test MCP registration endpoint
  try {
    console.log('\n3. Testing MCP registration endpoint...');
    const registrationEndpoint = `${MCP_SERVER_URL}/mcp-registration`;
    console.log(`   Requesting: ${registrationEndpoint}`);
    
    const registrationResponse = await axios.post(registrationEndpoint);
    console.log('   ✅ MCP registration successful');
    console.log('   Server name:', registrationResponse.data.serverInfo?.name || 'Not provided');
  } catch (error) {
    console.error('   ❌ MCP registration check failed');
    logDetailedError(error, 'registration');
    // Continue testing other endpoints
  }
  
  // Test chat endpoint
  try {
    console.log('\n4. Testing chat endpoint...');
    const chatEndpoint = `${MCP_SERVER_URL}/chat`;
    console.log(`   Requesting: ${chatEndpoint}`);
    
    const chatPayload = {
      message: 'Hello, this is a test message'
    };
    console.log('   Payload:', JSON.stringify(chatPayload));
    
    const chatResponse = await axios.post(chatEndpoint, chatPayload);
    console.log('   ✅ Chat endpoint successful');
    console.log('   Response:', chatResponse.data);
  } catch (error) {
    console.error('   ❌ Chat endpoint failed');
    logDetailedError(error, 'chat');
  }
  
  console.log('\nTest summary complete');
}

function logDetailedError(error, endpoint) {
  console.error(`   Error details for ${endpoint} endpoint:`);
  
  if (error.code === 'ECONNREFUSED') {
    console.error(`   Connection refused - Is the server running at ${MCP_SERVER_URL}?`);
  } else if (error.code === 'ETIMEDOUT') {
    console.error(`   Connection timed out - Server at ${MCP_SERVER_URL} is not responding`);
  }
  
  if (error.response) {
    // The request was made and the server responded with a status code
    // that falls out of the range of 2xx
    console.error(`   Status: ${error.response.status}`);
    console.error(`   Status Text: ${error.response.statusText}`);
    console.error(`   Headers:`, error.response.headers);
    console.error(`   Response data:`, error.response.data);
  } else if (error.request) {
    // The request was made but no response was received
    console.error(`   No response received from server`);
    console.error(`   Request details:`, error.request._currentUrl || error.request.path);
  } else {
    // Something happened in setting up the request that triggered an Error
    console.error(`   Error message:`, error.message);
  }
  
  if (error.stack) {
    console.error(`   Stack trace (first line):`, error.stack.split('\n')[0]);
  }
}

// Execute the test
testConnection().catch(err => {
  console.error('Unhandled error during testing:', err.message);
});