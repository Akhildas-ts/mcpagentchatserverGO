// server.js in mcp-agent-chat directory
import express from 'express';
import axios from 'axios';
import http from 'http';
import { Server } from 'socket.io';

const app = express();
const server = http.createServer(app);
const io = new Server(server);

app.use(express.json());

// Configure server info for MCP protocol
const serverInfo = {
  name: "MCP Agent Chat",
  version: "1.0.0",
  description: "Chat agent with vector search capabilities",
  vendor: {
    name: "Your Organization"
  },
  capabilities: {
    chat: true,
    vectorSearch: true,
    indexRepository: true
  }
};

// Increase default timeout for connections
const GO_SERVER_URL = process.env.GO_SERVER_URL || 'http://localhost:8081';
const MCP_PORT = process.env.MCP_PORT || 3000;

// Create a custom axios instance with longer timeout
const goServerClient = axios.create({
  baseURL: GO_SERVER_URL,
  timeout: 60000, // 60 seconds timeout
  headers: {'Content-Type': 'application/json'}
});

// Keep-alive configuration
server.keepAliveTimeout = 65000; // 65 seconds
server.headersTimeout = 66000; // 66 seconds

// Connection tracking middleware
const connectionTracking = (req, res, next) => {
  // Log incoming connections
  const clientIp = req.headers['x-forwarded-for'] || req.socket.remoteAddress;
  console.log(`[${new Date().toISOString()}] Connection from ${clientIp} to ${req.path}`);
  
  // Track open connections
  const originalEnd = res.end;
  
  // Set timeout for the request
  req.setTimeout(120000); // 2 minute timeout
  
  // Handle connection closure
  res.on('close', () => {
    console.log(`[${new Date().toISOString()}] Connection closed for ${req.path}`);
  });
  
  // Handle connection errors
  req.on('error', (err) => {
    console.error(`[${new Date().toISOString()}] Connection error for ${req.path}:`, err.message);
  });
  
  // Override end method to log responses
  res.end = function() {
    console.log(`[${new Date().toISOString()}] Response sent for ${req.path} with status ${res.statusCode}`);
    originalEnd.apply(res, arguments);
  };
  
  next();
};

app.use(connectionTracking);

console.log('[' + new Date().toISOString() + '] Initializing MCP server info');
console.log('Adding keep-alive to prevent client closed errors');

// MCP registration endpoint
app.post('/mcp-registration', (req, res) => {
  console.log('Received MCP registration request');
  res.json({
    serverInfo: serverInfo,
    status: 'success'
  });
});

// Handle connection errors gracefully
io.on('connection', (socket) => {
  console.log('Client connected:', socket.id);
  
  socket.on('disconnect', (reason) => {
    console.log('Client disconnected:', socket.id, 'Reason:', reason);
  });
  
  socket.on('error', (error) => {
    console.error('Socket error:', error);
  });
});

// Chat endpoint
app.post('/chat', async (req, res) => {
  try {
    const response = await goServerClient.post('/chat', req.body);
    res.json(response.data);
  } catch (error) {
    console.error('Error forwarding chat request:', error.message);
    res.status(500).json({ error: 'Failed to process chat request' });
  }
});

// Vector search endpoint
app.post('/vector-search', async (req, res) => {
  try {
    const response = await goServerClient.post('/vector-search', req.body);
    res.json(response.data);
  } catch (error) {
    console.error('Error forwarding vector search request:', error.message);
    res.status(500).json({ error: 'Failed to process vector search request' });
  }
});

// Index repository endpoint
app.post('/index-repository', async (req, res) => {
  try {
    const response = await goServerClient.post('/index-repository', req.body);
    res.json(response.data);
  } catch (error) {
    console.error('Error forwarding index repository request:', error.message);
    res.status(500).json({ error: 'Failed to process index repository request' });
  }
});

// Health check endpoint
app.get('/health', async (req, res) => {
  try {
    const response = await goServerClient.get('/health');
    res.json({ 
      mcp: 'healthy',
      goServer: response.data.status === 'ok' ? 'healthy' : 'unhealthy' 
    });
  } catch (error) {
    console.error('Error checking Go server health:', error.message);
    res.json({ 
      mcp: 'healthy',
      goServer: 'unhealthy' 
    });
  }
});

// Required MCP protocol endpoints
app.post('/offerings', (req, res) => {
  res.json({
    tools: [
      {
        id: "chat",
        name: "Chat",
        description: "Process a chat message with repository context",
        parameters: {
          type: "object",
          required: ["message"],
          properties: {
            message: { type: "string", description: "The message to process" },
            repository: { type: "string", description: "Repository to use for context" },
            context: { type: "array", items: { type: "string" }, description: "Additional context" }
          }
        }
      },
      {
        id: "vectorSearch",
        name: "Vector Search",
        description: "Search for code in a repository",
        parameters: {
          type: "object",
          required: ["query", "repository"],
          properties: {
            query: { type: "string", description: "The search query" },
            repository: { type: "string", description: "Repository to search in" },
            limit: { type: "number", description: "Maximum results to return" }
          }
        }
      },
      {
        id: "indexRepository",
        name: "Index Repository",
        description: "Index a repository for search",
        parameters: {
          type: "object",
          required: ["repository"],
          properties: {
            repository: { type: "string", description: "Repository URL to index" },
            branch: { type: "string", description: "Branch to index" }
          }
        }
      }
    ],
    resources: [],
    resourceTemplates: []
  });
});

server.listen(MCP_PORT, () => {
  console.log(`MCP Agent Chat Server listening on port ${MCP_PORT}`);
  
  // Verify connection to Go server on startup
  goServerClient.get('/health')
    .then(response => {
      console.log(`Connected to Go Vector Search Server at ${GO_SERVER_URL}`);
    })
    .catch(error => {
      console.error(`Failed to connect to Go Vector Search Server: ${error.message}`);
    });
});