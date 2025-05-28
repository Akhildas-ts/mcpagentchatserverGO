// middleware.js in mcp-agent-chat directory
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
  
  module.exports = {
    connectionTracking
  };