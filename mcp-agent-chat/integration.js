import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { z } from 'zod';
import axios from 'axios';
import 'dotenv/config';
import express from 'express';
import fetch from 'node-fetch';  // If you need fetch

const app = express();
const port = process.env.MCP_PORT || 3000;

// Debug logging
process.on('message', (message) => {
  console.error('DEBUG - MESSAGE RECEIVED:', JSON.stringify(message));
});

// Create an MCP server with stdio transport for integration with Go server
const server = new McpServer({
  name: 'agent-chat-mcp',
  version: '1.0.0',
});

// Configuration for your Go server
const config = {
  GO_SERVER_URL: process.env.GO_SERVER_URL || 'http://localhost:8081',
  MCP_SECRET_TOKEN: process.env.MCP_SECRET_TOKEN
};

console.error('DEBUG - CONFIG:', {
  GO_SERVER_URL: config.GO_SERVER_URL,
  MCP_SECRET_TOKEN: config.MCP_SECRET_TOKEN ? `${config.MCP_SECRET_TOKEN.slice(0, 5)}...` : 'not set'
});

// Add authentication header to requests if token is available
const axiosConfig = {};
if (config.MCP_SECRET_TOKEN) {
  axiosConfig.headers = {
    'X-MCP-Token': config.MCP_SECRET_TOKEN
  };
}

/**
 * Process a chat message by delegating to the Go server's vector search
 */
async function processChat(message, repository = '', context = {}) {
  try {
    console.error(`DEBUG - Processing chat: "${message}" for repo: ${repository}`);
    
    // First, search for relevant code using the Go server's vector search
    const searchResponse = await axios.post(`${config.GO_SERVER_URL}/vector-search`, {
      query: message,
      repository: repository,
      limit: 5
    }, axiosConfig);

    console.error('DEBUG - Search response status:', searchResponse.status);
    
    // Format the chat response with the search results
    const searchResults = searchResponse.data.success ? searchResponse.data.data : [];
    
    return {
      message: `I processed your message: "${message}"`,
      repository: repository,
      codeContext: searchResults,
      timestamp: new Date().toISOString()
    };
  } catch (error) {
    console.error('Error processing chat:', error.message);
    if (error.response) {
      console.error('Response data:', error.response.data);
      console.error('Response status:', error.response.status);
    }
    throw error;
  }
}

// Add chat tool that integrates with Go server
server.tool(
  'chat',
  {
    message: z.string().describe('The user message to process'),
    repository: z.string().optional().describe('The GitHub repository to reference'),
    context: z.record(z.any()).optional().describe('Additional context for the chat')
  },
  async ({ message, repository = '', context = {} }) => {
    try {
      console.error('DEBUG - Chat tool called with:', { message, repository });
      const result = await processChat(message, repository, context);
      console.error('DEBUG - Chat result:', result);
      return {
        content: [
          {
            type: 'text',
            text: JSON.stringify(result, null, 2),
          },
        ],
      };
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      return {
        content: [
          {
            type: 'text',
            text: JSON.stringify({
              status: 'error',
              message: `Error processing chat: ${errorMessage}`
            }, null, 2),
          },
        ],
        isError: true,
      };
    }
  }
);

// Pass through vector search to Go server
server.tool(
  'vectorSearch',
  {
    query: z.string().describe('The search query'),
    repository: z.string().describe('The repository to search in'),
    limit: z.number().optional().describe('Maximum number of results to return')
  },
  async ({ query, repository, limit = 5 }) => {
    try {
      console.error('DEBUG - Vector search tool called with:', { query, repository, limit });
      
      // Call the Go server's vector search endpoint
      const response = await axios.post(`${config.GO_SERVER_URL}/vector-search`, {
        query,
        repository,
        limit
      }, axiosConfig);

      console.error('DEBUG - Vector search result status:', response.status);
      
      // Return the results
      return {
        content: [
          {
            type: 'text',
            text: JSON.stringify(response.data, null, 2),
          },
        ],
      };
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      console.error('DEBUG - Vector search error:', errorMessage);
      if (error.response) {
        console.error('Response data:', error.response.data);
        console.error('Response status:', error.response.status);
      }
      return {
        content: [
          {
            type: 'text',
            text: JSON.stringify({
              status: 'error',
              message: `Search failed: ${errorMessage}`
            }, null, 2),
          },
        ],
        isError: true,
      };
    }
  }
);

// Add a repository indexing tool to your MCP server
server.tool(
  'indexRepository',
  {
    repoUrl: z.string().describe('The GitHub repository URL to index'),
    branch: z.string().optional().describe('The branch to index (default: main)')
  },
  async ({ repoUrl, branch = 'main' }) => {
    try {
      console.error(`DEBUG - Indexing repository: ${repoUrl}, branch: ${branch}`);
      
      // Call the Go server's repository indexing endpoint
      const response = await axios.post(`${config.GO_SERVER_URL}/index-repository`, {
        repoUrl,
        branch
      }, axiosConfig);

      console.error('DEBUG - Repository indexing result status:', response.status);
      
      return {
        content: [
          {
            type: 'text',
            text: JSON.stringify(response.data, null, 2),
          },
        ],
      };
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      console.error('DEBUG - Repository indexing error:', errorMessage);
      if (error.response) {
        console.error('Response data:', error.response.data);
        console.error('Response status:', error.response.status);
      }
      return {
        content: [
          {
            type: 'text',
            text: JSON.stringify({
              status: 'error',
              message: `Repository indexing failed: ${errorMessage}`
            }, null, 2),
          },
        ],
        isError: true,
      };
    }
  }
);

// Start the server with stdio transport
const transport = new StdioServerTransport();

// Log connection information
console.error('Agent Chat MCP server running on stdio');
console.error(`Connected to Go server at ${config.GO_SERVER_URL}`);

// Connect the server
server.connect(transport).catch((error) => {
  console.error('[MCP Error]', error);
  process.exit(1);
});

// Enable JSON parsing
app.use(express.json());

// Add your MCP routes
app.post('/mcp', async (req, res) => {
  try {
    console.log('Received request body:', req.body);

    // Verify token
    const authHeader = req.headers.authorization;
    if (!authHeader || `Bearer ${config.MCP_SECRET_TOKEN}` !== authHeader) {
      return res.status(401).json({ error: 'Unauthorized' });
    }

    let goServerRequest;
    
    if (req.body.tool === 'vectorSearch') {
      // Format for vector search
      goServerRequest = {
        query: req.body.params.query,
        repository: req.body.params.repository,
        branch: req.body.params.branch || 'main'
      };
    } else if (req.body.tool === 'indexRepository') {
      // Format for repository indexing
      goServerRequest = {
        repository: req.body.params.repoUrl.replace('https://github.com/', ''),
        branch: req.body.params.branch || 'main'
      };
    } else {
      // Direct format (as shown in your working example)
      goServerRequest = {
        query: req.body.query,
        repository: req.body.repository,
        branch: req.body.branch || 'main'
      };
    }

    console.log('Sending request to Go server:', goServerRequest);
    
    // Choose the correct endpoint based on the tool
    const endpoint = req.body.tool === 'indexRepository' ? '/index-repository' : '/vector-search';
    
    const response = await fetch(`${config.GO_SERVER_URL}${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(goServerRequest)
    });

    const responseData = await response.text();
    console.log('Go server response:', responseData);

    if (!response.ok) {
      console.error(`Go server responded with status ${response.status}`);
      return res.status(response.status).json({ 
        error: 'Request failed',
        message: responseData,
        success: false
      });
    }

    const data = JSON.parse(responseData);
    res.json(data);
  } catch (error) {
    console.error('Error processing request:', error);
    res.status(500).json({ 
      error: 'Internal server error',
      message: error.message,
      success: false
    });
  }
});

// Start HTTP server
app.listen(port, () => {
  console.log(`MCP server listening on port ${port}`);
  console.log(`Connected to Go server at ${config.GO_SERVER_URL}`);
}); 