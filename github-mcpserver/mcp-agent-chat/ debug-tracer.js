// debug-tracer.js - Save this in your mcp-agent-chat directory
const fs = require('fs');
const path = require('path');

// Create a detailed log file with timestamps
const LOG_FILE = path.join(__dirname, 'mcp-full-debug.log');

// Initialize log file
fs.writeFileSync(LOG_FILE, `MCP Debug Tracer Started at ${new Date().toISOString()}\n\n`);
fs.appendFileSync(LOG_FILE, `Process ID: ${process.pid}\n`);
fs.appendFileSync(LOG_FILE, `Working Directory: ${process.cwd()}\n`);
fs.appendFileSync(LOG_FILE, `Node Version: ${process.version}\n`);
fs.appendFileSync(LOG_FILE, `Platform: ${process.platform}\n`);
fs.appendFileSync(LOG_FILE, `Command: ${process.argv.join(' ')}\n\n`);

// Log environment variables
fs.appendFileSync(LOG_FILE, `Environment Variables:\n${JSON.stringify(process.env, null, 2)}\n\n`);

// Detailed logging function
function log(type, message, data = null) {
  const timestamp = new Date().toISOString();
  let logMessage = `[${timestamp}] [${type}] ${message}\n`;
  
  if (data) {
    if (typeof data === 'object') {
      try {
        logMessage += `  Data: ${JSON.stringify(data, null, 2)}\n`;
      } catch (error) {
        logMessage += `  Data: [Cannot stringify data: ${error.message}]\n`;
      }
    } else {
      logMessage += `  Data: ${data}\n`;
    }
  }
  
  fs.appendFileSync(LOG_FILE, logMessage);
  
  // Also log to stderr for immediate feedback
  process.stderr.write(`${type}: ${message}\n`);
}

// Log startup
log('INFO', 'Debug tracer initialized');
log('INFO', 'Setting up STDIO interceptors');

// Hook into stdout to capture what we send to the client
const originalStdoutWrite = process.stdout.write;
process.stdout.write = function(chunk, encoding, callback) {
  // Log what we're sending to the client
  if (typeof chunk === 'string') {
    log('STDOUT', 'Data sent to client', chunk);
  } else {
    log('STDOUT', 'Binary data sent to client', '[binary data]');
  }
  
  return originalStdoutWrite.apply(this, arguments);
};

// Hook into stdin to capture what the client sends to us
process.stdin.on('data', (data) => {
  const input = data.toString();
  log('STDIN', 'Data received from client', input);
  
  // Try to parse it as JSON
  try {
    const parsed = JSON.parse(input);
    log('PARSED', 'Parsed client input as JSON', parsed);
  } catch (error) {
    log('ERROR', `Failed to parse client input as JSON: ${error.message}`);
  }
});

// Capture uncaught exceptions
process.on('uncaughtException', (error) => {
  log('EXCEPTION', `Uncaught exception: ${error.message}`, error.stack);
});

// Capture unhandled promise rejections
process.on('unhandledRejection', (reason, promise) => {
  log('REJECTION', `Unhandled rejection: ${reason}`);
});

// Capture process signals
process.on('SIGINT', () => {
  log('SIGNAL', 'Received SIGINT');
  process.exit(0);
});

process.on('SIGTERM', () => {
  log('SIGNAL', 'Received SIGTERM');
  process.exit(0);
});

// Log process exit
process.on('exit', (code) => {
  log('EXIT', `Process exiting with code ${code}`);
});

// Now send the server info notification
// This is critical - it must be the first JSON-RPC message
log('ACTION', 'Sending initial server info notification');
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

process.stdout.write(JSON.stringify({
  jsonrpc: "2.0",
  method: "serverInfo",
  params: { serverInfo }
}) + "\n");

// Listen for incoming messages
process.stdin.setEncoding('utf8');
process.stdin.resume();

// Simple mapping for outgoing requests
const pendingRequests = new Map();

// Define offerings
const offerings = {
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
};

// Handle incoming requests
process.stdin.on('data', (data) => {
  const input = data.toString().trim();
  if (!input) return;
  
  const lines = input.split('\n').filter(Boolean);
  
  for (const line of lines) {
    try {
      const request = JSON.parse(line);
      log('REQUEST', `Processing request method: ${request.method}`, request);
      
      // Handle method
      switch (request.method) {
        case 'getServerInfo':
          sendResponse({ serverInfo }, request.id);
          break;
          
        case 'listOfferings':
          sendResponse(offerings, request.id);
          break;
          
        default:
          // All other methods just get a success response
          sendResponse({ status: 'success' }, request.id);
      }
    } catch (error) {
      log('ERROR', `Error processing request: ${error.message}`);
      // Send an error response if possible
      try {
        const request = JSON.parse(line);
        sendError(-32603, `Internal error: ${error.message}`, request.id);
      } catch (e) {
        // If we can't even parse the JSON, send a generic error
        sendError(-32700, 'Parse error', null);
      }
    }
  }
});

// Send a JSON-RPC response
function sendResponse(result, id) {
  const response = {
    jsonrpc: '2.0',
    result,
    id
  };
  
  log('RESPONSE', `Sending response for id: ${id}`, response);
  process.stdout.write(JSON.stringify(response) + '\n');
}

// Send a JSON-RPC error
function sendError(code, message, id) {
  const error = {
    jsonrpc: '2.0',
    error: {
      code,
      message
    },
    id
  };
  
  log('ERROR', `Sending error response: ${message}`, error);
  process.stdout.write(JSON.stringify(error) + '\n');
}

// Keep the process alive
log('INFO', 'Starting keep-alive interval');
setInterval(() => {
  log('HEARTBEAT', 'Process still alive');
}, 5000);

// Let the user know we're ready
log('INFO', 'Debug tracer ready for client communication');