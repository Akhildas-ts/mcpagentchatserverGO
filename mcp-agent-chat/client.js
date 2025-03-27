import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";

async function main() {
  try {
    console.log("Starting MCP client test...");
    
    // Create and connect client
    const client = new Client({name: "agent-chat-client", version: "1.0.0"});
    await client.connect(new StdioClientTransport({
      command: "node", 
      args: ["server.js"]
    }));
    
    console.log("Connected to MCP server");
    
    // Test chat tool
    console.log("\n--- Testing chat tool ---");
    const chatResult = await client.callTool({
      name: "chat", 
      arguments: { 
        message: "Hello, can you help me search for code?",
        repository: "example/repo"
      }
    });
    
    console.log("Chat result:");
    console.log(chatResult.content[0].text);
    
    // Test vector search tool
    console.log("\n--- Testing vectorSearch tool ---");
    const searchResult = await client.callTool({
      name: "vectorSearch", 
      arguments: { 
        query: "function search",
        repository: "example/repo",
        limit: 2
      }
    });
    
    console.log("Search result:");
    console.log(searchResult.content[0].text);
    
    // No need to disconnect - process will exit naturally
    console.log("\nTest completed successfully");
    
  } catch (error) {
    console.error("Error during MCP client test:", error);
    process.exit(1);
  }
}

main(); 