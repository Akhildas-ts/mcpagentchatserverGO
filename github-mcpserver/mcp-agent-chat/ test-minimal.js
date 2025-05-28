// test-minimal.js
const { spawn } = require('child_process');
const path = require('path');

console.log('Testing minimal STDIO handler...');

// Start the minimal handler
const child = spawn('node', ['index.js'], {
  stdio: ['pipe', 'pipe', 'inherit']
});

// Listen for data from the child process
child.stdout.on('data', (data) => {
  const message = data.toString();
  console.log('Received:', message);
  
  try {
    // For each line in the output
    const lines = message.split('\n').filter(Boolean);
    for (const line of lines) {
      const parsed = JSON.parse(line);
      
      // If we got server info, send a listOfferings request
      if (parsed.method === 'serverInfo') {
        console.log('✅ Got server info notification!');
        
        // Send a listOfferings request
        const request = {
          jsonrpc: '2.0',
          method: 'listOfferings',
          id: 1
        };
        
        console.log('Sending listOfferings request...');
        child.stdin.write(JSON.stringify(request) + '\n');
      }
      
      // If we got offerings, we're done
      if (parsed.id === 1 && parsed.result && parsed.result.tools) {
        console.log('✅ Got offerings response!');
        console.log(`Found ${parsed.result.tools.length} tools`);
        console.log('Test successful!');
        
        // Exit the test
        setTimeout(() => {
          child.kill();
          process.exit(0);
        }, 1000);
      }
    }
  } catch (err) {
    console.error('Error parsing message:', err.message);
  }
});

// Handle child process exit
child.on('close', (code) => {
  console.log(`Child process exited with code ${code}`);
  process.exit(code);
});

// Handle errors
child.on('error', (err) => {
  console.error('Error spawning child process:', err.message);
  process.exit(1);
});