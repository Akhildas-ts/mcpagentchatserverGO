import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { z } from 'zod';
import axios from 'axios';
import 'dotenv/config';

process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
});
process.on('uncaughtException', (err) => {
  console.error('Uncaught Exception:', err);
});

// Create an MCP server
const serverMcp = new McpServer({
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

// Essential GitHub tools
serverMcp.tool(
  'github_create_repository',
  {
    name: z.string().describe('Repository name'),
    description: z.string().optional().describe('Repository description'),
    private: z.boolean().optional().describe('Whether the repository should be private'),
    autoInit: z.boolean().optional().describe('Initialize with README.md')
  },
  async ({ name, description, private: isPrivate, autoInit }) => {
    try {
      const response = await axios.post(`${config.GO_SERVER_URL}/github/repos`, {
        name,
        description,
        private: isPrivate,
        auto_init: autoInit
      }, axiosConfig);
      
      return {
        content: [{ type: 'text', text: JSON.stringify(response.data, null, 2) }]
      };
    } catch (error) {
      console.error('Error creating repository:', error.message);
      return {
        content: [{ type: 'text', text: `Error: ${error.message}` }],
        isError: true
      };
    }
  }
);

serverMcp.tool(
  'github_create_file',
  {
    owner: z.string().describe('Repository owner'),
    repo: z.string().describe('Repository name'),
    path: z.string().describe('File path'),
    content: z.string().describe('File content'),
    message: z.string().describe('Commit message'),
    branch: z.string().describe('Branch name')
  },
  async ({ owner, repo, path, content, message, branch }) => {
    try {
      const response = await axios.post(`${config.GO_SERVER_URL}/github/contents`, {
        owner,
        repo,
        path,
        content,
        message,
        branch
      }, axiosConfig);
      
      return {
        content: [{ type: 'text', text: JSON.stringify(response.data, null, 2) }]
      };
    } catch (error) {
      console.error('Error creating file:', error.message);
      return {
        content: [{ type: 'text', text: `Error: ${error.message}` }],
        isError: true
      };
    }
  }
);

serverMcp.tool(
  'github_create_issue',
  {
    owner: z.string(),
    repo: z.string(),
    title: z.string(),
    body: z.string().optional(),
    labels: z.array(z.string()).optional(),
    assignees: z.array(z.string()).optional()
  },
  async ({ owner, repo, title, body, labels, assignees }) => {
    try {
      const response = await axios.post(`${config.GO_SERVER_URL}/github/issues`, {
        owner,
        repo,
        title,
        body,
        labels,
        assignees
      }, axiosConfig);
      
      return {
        content: [{ type: 'text', text: JSON.stringify(response.data, null, 2) }]
      };
    } catch (error) {
      console.error('Error creating issue:', error.message);
      return {
        content: [{ type: 'text', text: `Error: ${error.message}` }],
        isError: true
      };
    }
  }
);

serverMcp.tool(
  'github_create_pull_request',
  {
    owner: z.string(),
    repo: z.string(),
    title: z.string(),
    head: z.string(),
    base: z.string(),
    body: z.string().optional(),
    draft: z.boolean().optional()
  },
  async ({ owner, repo, title, head, base, body, draft }) => {
    try {
      const response = await axios.post(`${config.GO_SERVER_URL}/github/pulls`, {
        owner,
        repo,
        title,
        head,
        base,
        body,
        draft
      }, axiosConfig);
      
      return {
        content: [{ type: 'text', text: JSON.stringify(response.data, null, 2) }]
      };
    } catch (error) {
      console.error('Error creating pull request:', error.message);
      return {
        content: [{ type: 'text', text: `Error: ${error.message}` }],
        isError: true
      };
    }
  }
);

// Vector search tool
serverMcp.tool(
  'vector_search',
  {
    query: z.string().describe('The search query'),
    repository: z.string().describe('The repository to search in'),
    limit: z.number().optional().describe('Maximum number of results to return')
  },
  async ({ query, repository, limit = 5 }) => {
    try {
      const response = await axios.post(`${config.GO_SERVER_URL}/vector-search`, {
        query,
        repository,
        limit
      }, axiosConfig);
      
      return {
        content: [{ type: 'text', text: JSON.stringify(response.data, null, 2) }]
      };
    } catch (error) {
      console.error('Error performing vector search:', error.message);
      return {
        content: [{ type: 'text', text: `Error: ${error.message}` }],
        isError: true
      };
    }
  }
);

// Add a chat tool for MCP compatibility
serverMcp.tool(
  'chat',
  {
    message: z.string().describe('The user message to process'),
    repository: z.string().optional().describe('The GitHub repository to reference'),
    context: z.record(z.any()).optional().describe('Additional context for the chat')
  },
  async ({ message, repository = '', context = {} }) => {
    return {
      content: [
        {
          type: 'text',
          text: JSON.stringify({
            message: `Echo: ${message}`,
            repository,
            context,
            timestamp: new Date().toISOString()
          }, null, 2),
        },
      ],
    };
  }
);

// Start the MCP server with stdio transport
const transport = new StdioServerTransport();
serverMcp.connect(transport).catch((error) => {
  console.error('[MCP Error]', error);
  process.exit(1);
});

// Keep the process alive for Cursor MCP stdio
setInterval(() => {}, 1000 * 60 * 60);