import axios from 'axios';
import 'dotenv/config';

// Configuration
const GO_SERVER_URL = process.env.GO_SERVER_URL || 'http://localhost:8081';
const MCP_SECRET_TOKEN = process.env.MCP_SECRET_TOKEN;

// Create axios config with headers
const axiosConfig = {};
if (MCP_SECRET_TOKEN) {
  axiosConfig.headers = {
    'X-MCP-Token': MCP_SECRET_TOKEN
  };
  console.log('Using MCP_SECRET_TOKEN from .env file');
} else {
  console.log('WARNING: MCP_SECRET_TOKEN not set in .env file');
}

async function main() {
  console.log('========== MCP-AGENT-CHAT DEBUG SCRIPT ==========');
  console.log('Testing connection to Go server at:', GO_SERVER_URL);
  
  try {
    // 1. Test health endpoint
    console.log('\n1. Testing Go server health endpoint...');
    const healthResponse = await axios.get(`${GO_SERVER_URL}/health`);
    console.log(`✅ Health endpoint response (${healthResponse.status}):`, healthResponse.data);
    
    // 2. Test MCP info endpoint
    console.log('\n2. Testing Go server MCP info endpoint...');
    try {
      const mcpInfoResponse = await axios.get(`${GO_SERVER_URL}/mcp-info`);
      console.log(`✅ MCP info endpoint response (${mcpInfoResponse.status}):`, mcpInfoResponse.data);
    } catch (error) {
      console.log('❌ MCP info endpoint error:', error.message);
      if (error.response) {
        console.log('   Response:', error.response.status, error.response.data);
      }
    }
    
    // 3. Test vector search without token
    console.log('\n3. Testing vector search WITHOUT token...');
    try {
      const searchRequestNoToken = {
        query: 'test search',
        repository: 'example/repo',
        limit: 2
      };
      const searchResponseNoToken = await axios.post(
        `${GO_SERVER_URL}/vector-search`, 
        searchRequestNoToken
      );
      console.log(`✅ Vector search response without token (${searchResponseNoToken.status}):`, 
        searchResponseNoToken.data);
    } catch (error) {
      console.log('❌ Vector search without token error:', error.message);
      if (error.response) {
        console.log('   Response:', error.response.status, error.response.data);
      }
    }
    
    // 4. Test vector search with token
    console.log('\n4. Testing vector search WITH token...');
    try {
      const searchRequestWithToken = {
        query: 'test search',
        repository: 'example/repo',
        limit: 2
      };
      const searchResponseWithToken = await axios.post(
        `${GO_SERVER_URL}/vector-search`, 
        searchRequestWithToken,
        axiosConfig
      );
      console.log(`✅ Vector search response with token (${searchResponseWithToken.status}):`, 
        searchResponseWithToken.data);
    } catch (error) {
      console.log('❌ Vector search with token error:', error.message);
      if (error.response) {
        console.log('   Response:', error.response.status, error.response.data);
      }
    }
    
    console.log('\n========== DEBUG COMPLETED ==========');
    console.log('\nRecommendations:');
    console.log('1. If vector search fails with both with/without token, check if your Go server is properly set up for vector search');
    console.log('2. If it only works with the token, make sure your MCP integration is sending the token correctly');
    console.log('3. Check that Pinecone and OpenAI API keys are valid in your Go server');
    
  } catch (error) {
    console.error('Unexpected error during debug:', error);
  }
}

main(); 