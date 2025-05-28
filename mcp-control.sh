#!/bin/bash

function cleanup() {
    echo "Cleaning up running processes..."
    pkill -f "node.*integration" 2>/dev/null
    pkill -f "go.*run.*main.go" 2>/dev/null
    sleep 1
}

function check_ports() {
    if lsof -i :3000 > /dev/null || lsof -i :8081 > /dev/null; then
        echo "⚠️  Ports 3000 or 8081 are still in use. Cleaning up..."
        cleanup
    fi
}

function start_servers() {
    echo "Starting Go server..."
    cd tools && go run cmd/main.go &
    sleep 2
    
    echo "Starting MCP agent..."
    cd ../mcp-agent-chat && npm start &
    sleep 2
    
    echo "✅ Servers started! Use './mcp-control.sh stop' to stop them"
}

function stop_servers() {
    echo "Stopping all servers..."
    cleanup
    echo "✅ All servers stopped"
}

case "$1" in
    "start")
        check_ports
        start_servers
        ;;
    "stop")
        stop_servers
        ;;
    "restart")
        stop_servers
        sleep 1
        start_servers
        ;;
    *)
        echo "Usage: ./mcp-control.sh [start|stop|restart]"
        ;;
esac 