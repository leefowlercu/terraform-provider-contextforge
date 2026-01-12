#!/bin/bash
set -e

echo "🧹 Cleaning up Context Forge test environment..."

# Get the project root directory (one level up from scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Stop MCP time server process
if [ -f "$PROJECT_ROOT/tmp/time-server.pid" ]; then
    TIME_SERVER_PID=$(cat "$PROJECT_ROOT/tmp/time-server.pid")
    if ps -p $TIME_SERVER_PID > /dev/null 2>&1; then
        echo "🛑 Stopping MCP time server (PID: $TIME_SERVER_PID)..."
        kill $TIME_SERVER_PID 2>/dev/null || true
        sleep 1
        # Force kill if still running
        kill -9 $TIME_SERVER_PID 2>/dev/null || true
    fi
fi

# Stop additional MCP time servers
for port in 8003 8004; do
  if [ -f "$PROJECT_ROOT/tmp/time-server-$port.pid" ]; then
    SERVER_PID=$(cat "$PROJECT_ROOT/tmp/time-server-$port.pid")
    if ps -p $SERVER_PID > /dev/null 2>&1; then
      echo "🛑 Stopping MCP time server on port $port (PID: $SERVER_PID)..."
      kill $SERVER_PID 2>/dev/null || true
      sleep 1
      kill -9 $SERVER_PID 2>/dev/null || true
    fi
  fi
done

# Stop ContextForge gateway process
if [ -f "$PROJECT_ROOT/tmp/contextforge-test.pid" ]; then
    PID=$(cat "$PROJECT_ROOT/tmp/contextforge-test.pid")
    if ps -p $PID > /dev/null 2>&1; then
        echo "🛑 Stopping ContextForge gateway (PID: $PID)..."
        kill $PID 2>/dev/null || true
        sleep 2
        # Force kill if still running
        kill -9 $PID 2>/dev/null || true
    fi
fi

# Clean up test artifacts directory
echo "🗑️  Removing test artifacts..."
rm -rf "$PROJECT_ROOT/tmp"

echo "✅ Test environment cleaned up!"
