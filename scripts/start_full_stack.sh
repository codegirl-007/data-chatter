#!/bin/bash

echo "🚀 Starting Data Chatter Full Stack"
echo "===================================="

# Check if API server is already running
if curl -s http://localhost:8081/health > /dev/null 2>&1; then
    echo "✅ API server already running on port 8081"
else
    echo "🔧 Starting API server..."
    cd /Users/stephaniegredell/data-chatter
    ./bin/server &
    API_PID=$!
    echo "API server PID: $API_PID"
    
    # Wait for API server to start
    echo "⏳ Waiting for API server to start..."
    for i in {1..10}; do
        if curl -s http://localhost:8081/health > /dev/null 2>&1; then
            echo "✅ API server is ready!"
            break
        fi
        echo "   Attempt $i/10..."
        sleep 2
    done
fi

# Start web server
echo "🌐 Starting web interface..."
cd /Users/stephaniegredell/data-chatter/web
go run server.go &
WEB_PID=$!
echo "Web server PID: $WEB_PID"

# Wait for web server to start
echo "⏳ Waiting for web server to start..."
sleep 3

echo ""
echo "🎉 Data Chatter is ready!"
echo "========================="
echo "📊 API Server: http://localhost:8081"
echo "🌐 Web Interface: http://localhost:3000"
echo ""
echo "💡 Open http://localhost:3000 in your browser"
echo "🔧 Press Ctrl+C to stop all servers"
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Shutting down servers..."
    kill $API_PID 2>/dev/null
    kill $WEB_PID 2>/dev/null
    echo "✅ All servers stopped"
    exit 0
}

# Trap Ctrl+C
trap cleanup INT

# Wait for user to stop
wait
