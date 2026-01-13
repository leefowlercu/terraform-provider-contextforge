#!/bin/bash
set -e

echo "🚀 Starting Context Forge test environment..."

# Get the project root directory (one level up from scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Check for uvx
if ! command -v uvx &> /dev/null; then
    echo "❌ uvx is required but not installed"
    echo "   Install with: brew install uv"
    exit 1
fi

echo "✓ Using uvx ($(uvx --version 2>&1))"

echo "🔧 Starting Context Forge gateway..."

# Clean up old database to ensure fresh state
echo "🗑️  Cleaning up old database files..."
rm -rfv "$PROJECT_ROOT/tmp" || true
mkdir -p "$PROJECT_ROOT/tmp"
echo "✓ Database cleanup complete"

# Change to project root to ensure database is created in the right place
cd "$PROJECT_ROOT"

# Set environment variables and start gateway in background
export MCPGATEWAY_ADMIN_API_ENABLED=true
export MCPGATEWAY_UI_ENABLED=true
export PLATFORM_ADMIN_EMAIL=admin@test.local
export PLATFORM_ADMIN_PASSWORD=testpassword123
export PLATFORM_ADMIN_FULL_NAME="Platform Administrator"
export DATABASE_URL=sqlite:///tmp/contextforge-test.db
export LOG_LEVEL=INFO
export REDIS_ENABLED=false
export OTEL_ENABLE_OBSERVABILITY=false
export AUTH_REQUIRED=false
export MCP_CLIENT_AUTH_ENABLED=false
export BASIC_AUTH_USER=admin
export BASIC_AUTH_PASSWORD=testpassword123
export JWT_SECRET_KEY="test-secret-key-for-integration-testing"
export SECURE_COOKIES=false

# Start gateway in background
uvx --from 'mcp-contextforge-gateway==1.0.0b1' mcpgateway --host 0.0.0.0 --port 8000 > "$PROJECT_ROOT/tmp/contextforge-test.log" 2>&1 &
GATEWAY_PID=$!
echo $GATEWAY_PID > "$PROJECT_ROOT/tmp/contextforge-test.pid"

echo "⏳ Waiting for Context Forge to be ready..."

# Wait for health check
MAX_RETRIES=30
RETRY_COUNT=0

until curl -f http://localhost:8000/health > /dev/null 2>&1; do
  RETRY_COUNT=$((RETRY_COUNT + 1))
  if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
    echo "❌ Context Forge failed to start within timeout"
    cat "$PROJECT_ROOT/tmp/contextforge-test.log"
    kill $GATEWAY_PID 2>/dev/null || true
    exit 1
  fi
  echo "Waiting for Context Forge... (attempt $RETRY_COUNT/$MAX_RETRIES)"
  sleep 2
done

echo "✅ Context Forge is ready!"
echo ""

# Generate JWT token for integration tests
echo "🔑 Generating JWT token for integration tests..."
uvx --from 'mcp-contextforge-gateway==1.0.0b1' python3 -m mcpgateway.utils.create_jwt_token \
  --username admin@test.local \
  --exp 10080 \
  --secret test-secret-key-for-integration-testing > "$PROJECT_ROOT/tmp/contextforge-test-token.txt" 2>&1

if [ $? -eq 0 ]; then
  # Extract just the token (remove any extra output)
  TOKEN=$(cat "$PROJECT_ROOT/tmp/contextforge-test-token.txt" | grep -o 'eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*' || cat "$PROJECT_ROOT/tmp/contextforge-test-token.txt")
  echo "$TOKEN" > "$PROJECT_ROOT/tmp/contextforge-test-token.txt"
  export CONTEXTFORGE_TEST_TOKEN="$TOKEN"
  echo "✅ JWT token generated successfully"
else
  echo "⚠️  Failed to generate JWT token, tests may fail"
  cat "$PROJECT_ROOT/tmp/contextforge-test-token.txt"
fi
echo ""

echo "📊 Test Environment Information:"
echo "   ContextForge Address: http://localhost:8000"
echo "   ContextForge PID: $GATEWAY_PID"
echo "   Admin Email: admin@test.local"
echo "   Admin Password: testpassword123"
echo "   JWT Token File: $PROJECT_ROOT/tmp/contextforge-test-token.txt"
echo "   MCP Time Server: http://localhost:8002"
echo "   Time Server PID File: $PROJECT_ROOT/tmp/time-server.pid"
echo "   Test Gateway ID File: $PROJECT_ROOT/tmp/contextforge-test-gateway-id.txt"
echo "   Test Server ID File: $PROJECT_ROOT/tmp/contextforge-test-server-id.txt"
echo "   Test Tool ID File: $PROJECT_ROOT/tmp/contextforge-test-tool-id.txt"
echo "   Test Resource ID File: $PROJECT_ROOT/tmp/contextforge-test-resource-id.txt"
echo "   Test Team ID File: $PROJECT_ROOT/tmp/contextforge-test-team-id.txt"
echo "   Test Agent ID File: $PROJECT_ROOT/tmp/contextforge-test-agent-id.txt"
echo "   Test Prompt ID File: $PROJECT_ROOT/tmp/contextforge-test-prompt-id.txt"
echo ""

# Get version info
echo "🔍 Gateway Version: $(curl -s -u admin@test.local:testpassword123 http://localhost:8000/version | jq -r .app.version || echo \"Version endpoint unavailable\")"
echo ""

# Start MCP time server for gateway testing
echo "⏰ Starting MCP time server..."
uvx --from 'mcp-contextforge-gateway==1.0.0b1' python3 -m mcpgateway.translate \
  --stdio "uvx mcp-server-time --local-timezone=UTC" \
  --port 8002 > "$PROJECT_ROOT/tmp/time-server.log" 2>&1 &
TIME_SERVER_PID=$!
echo $TIME_SERVER_PID > "$PROJECT_ROOT/tmp/time-server.pid"

echo "⏳ Waiting for MCP time server to be ready..."

# Wait for time server to start (no health endpoint, so use fixed delay)
sleep 5

# Verify process is still running
if ! ps -p $TIME_SERVER_PID > /dev/null 2>&1; then
  echo "❌ Time server process died"
  cat "$PROJECT_ROOT/tmp/time-server.log"
  kill $GATEWAY_PID 2>/dev/null || true
  exit 1
fi

echo "✅ MCP time server is ready!"
echo ""

# Start additional MCP time servers for gateway resource tests
echo "⏰ Starting additional MCP time servers for gateway resource tests..."

# Port 8003 for basic gateway resource test
uvx --from 'mcp-contextforge-gateway==1.0.0b1' python3 -m mcpgateway.translate \
  --stdio "uvx mcp-server-time --local-timezone=UTC" \
  --port 8003 > "$PROJECT_ROOT/tmp/time-server-8003.log" 2>&1 &
TIME_SERVER_8003_PID=$!
echo $TIME_SERVER_8003_PID > "$PROJECT_ROOT/tmp/time-server-8003.pid"

# Port 8004 for import gateway resource test
uvx --from 'mcp-contextforge-gateway==1.0.0b1' python3 -m mcpgateway.translate \
  --stdio "uvx mcp-server-time --local-timezone=UTC" \
  --port 8004 > "$PROJECT_ROOT/tmp/time-server-8004.log" 2>&1 &
TIME_SERVER_8004_PID=$!
echo $TIME_SERVER_8004_PID > "$PROJECT_ROOT/tmp/time-server-8004.pid"

echo "⏳ Waiting for additional MCP time servers to be ready..."
sleep 5

# Verify processes are still running
for port in 8003 8004; do
  pid=$(cat "$PROJECT_ROOT/tmp/time-server-$port.pid")
  if ! ps -p $pid > /dev/null 2>&1; then
    echo "❌ Time server on port $port died"
    cat "$PROJECT_ROOT/tmp/time-server-$port.log"
    kill $GATEWAY_PID 2>/dev/null || true
    kill $TIME_SERVER_PID 2>/dev/null || true
    exit 1
  fi
done

echo "✅ Additional MCP time servers are ready!"
echo "   Port 8003: PID $TIME_SERVER_8003_PID"
echo "   Port 8004: PID $TIME_SERVER_8004_PID"
echo ""

# Create test gateway pointing to time server
echo "🔧 Creating test gateway..."
GATEWAY_RESPONSE=$(curl -s -X POST http://localhost:8000/gateways \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-time-server",
    "url": "http://localhost:8002/sse",
    "description": "Test gateway for integration tests",
    "transport": "SSE"
  }')

if [ $? -eq 0 ]; then
  GATEWAY_ID=$(echo "$GATEWAY_RESPONSE" | jq -r '.id // empty')

  if [ -n "$GATEWAY_ID" ] && [ "$GATEWAY_ID" != "null" ]; then
    echo "$GATEWAY_ID" > "$PROJECT_ROOT/tmp/contextforge-test-gateway-id.txt"
    echo "✅ Test gateway created successfully"
    echo "   Gateway ID: $GATEWAY_ID"
  else
    echo "⚠️  Failed to extract gateway ID from response"
    echo "   Response: $GATEWAY_RESPONSE"
  fi
else
  echo "⚠️  Failed to create test gateway"
fi
echo ""

# Create test server
echo "🔧 Creating test server..."
SERVER_RESPONSE=$(curl -s -X POST http://localhost:8000/servers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "server": {
      "name": "test-server",
      "description": "Test server for integration tests",
      "tags": ["test", "integration"]
    }
  }')

if [ $? -eq 0 ]; then
  SERVER_ID=$(echo "$SERVER_RESPONSE" | jq -r '.id // empty')

  if [ -n "$SERVER_ID" ] && [ "$SERVER_ID" != "null" ]; then
    echo "$SERVER_ID" > "$PROJECT_ROOT/tmp/contextforge-test-server-id.txt"
    echo "✅ Test server created successfully"
    echo "   Server ID: $SERVER_ID"
  else
    echo "⚠️  Failed to extract server ID from response"
    echo "   Response: $SERVER_RESPONSE"
  fi
else
  echo "⚠️  Failed to create test server"
fi
echo ""

# Create test tool
echo "🔧 Creating test tool..."
TOOL_RESPONSE=$(curl -s -X POST http://localhost:8000/tools \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tool": {
      "name": "test-tool",
      "description": "Test tool for integration tests",
      "inputSchema": {
        "type": "object",
        "properties": {
          "message": {
            "type": "string",
            "description": "Message to process"
          }
        },
        "required": ["message"]
      },
      "tags": ["test", "integration"]
    }
  }')

if [ $? -eq 0 ]; then
  TOOL_ID=$(echo "$TOOL_RESPONSE" | jq -r '.id // empty')

  if [ -n "$TOOL_ID" ] && [ "$TOOL_ID" != "null" ]; then
    echo "$TOOL_ID" > "$PROJECT_ROOT/tmp/contextforge-test-tool-id.txt"
    echo "✅ Test tool created successfully"
    echo "   Tool ID: $TOOL_ID"
  else
    echo "⚠️  Failed to extract tool ID from response"
    echo "   Response: $TOOL_RESPONSE"
  fi
else
  echo "⚠️  Failed to create test tool"
fi
echo ""

# Create test resource
echo "🔧 Creating test resource..."
RESOURCE_RESPONSE=$(curl -s -X POST http://localhost:8000/resources \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "resource": {
      "uri": "test://integration/resource",
      "name": "test-resource",
      "description": "Test resource for integration tests",
      "content": "This is test content for integration testing",
      "mimeType": "text/plain",
      "tags": ["test", "integration"]
    }
  }')

if [ $? -eq 0 ]; then
  RESOURCE_ID=$(echo "$RESOURCE_RESPONSE" | jq -r '.id // empty')

  if [ -n "$RESOURCE_ID" ] && [ "$RESOURCE_ID" != "null" ]; then
    echo "$RESOURCE_ID" > "$PROJECT_ROOT/tmp/contextforge-test-resource-id.txt"
    echo "✅ Test resource created successfully"
    echo "   Resource ID: $RESOURCE_ID"
  else
    echo "⚠️  Failed to extract resource ID from response"
    echo "   Response: $RESOURCE_RESPONSE"
  fi
else
  echo "⚠️  Failed to create test resource"
fi
echo ""

# Create test team
echo "🔧 Creating test team..."
TEAM_RESPONSE=$(curl -L -s -X POST http://localhost:8000/teams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-team",
    "slug": "test-team",
    "description": "Test team for integration tests"
  }')

if [ $? -eq 0 ]; then
  TEAM_ID=$(echo "$TEAM_RESPONSE" | jq -r '.id // empty')

  if [ -n "$TEAM_ID" ] && [ "$TEAM_ID" != "null" ]; then
    echo "$TEAM_ID" > "$PROJECT_ROOT/tmp/contextforge-test-team-id.txt"
    echo "✅ Test team created successfully"
    echo "   Team ID: $TEAM_ID"
  else
    echo "⚠️  Failed to extract team ID from response"
    echo "   Response: $TEAM_RESPONSE"
  fi
else
  echo "⚠️  Failed to create test team"
fi
echo ""

# Create test agent
echo "🔧 Creating test agent..."
AGENT_RESPONSE=$(curl -s -X POST http://localhost:8000/a2a \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "agent": {
      "name": "test-agent",
      "description": "Test agent for integration tests",
      "endpoint_url": "http://localhost:9000/agent",
      "enabled": true,
      "tags": ["test", "integration"]
    }
  }')

if [ $? -eq 0 ]; then
  AGENT_ID=$(echo "$AGENT_RESPONSE" | jq -r '.id // empty')

  if [ -n "$AGENT_ID" ] && [ "$AGENT_ID" != "null" ]; then
    echo "$AGENT_ID" > "$PROJECT_ROOT/tmp/contextforge-test-agent-id.txt"
    echo "✅ Test agent created successfully"
    echo "   Agent ID: $AGENT_ID"
  else
    echo "⚠️  Failed to extract agent ID from response"
    echo "   Response: $AGENT_RESPONSE"
  fi
else
  echo "⚠️  Failed to create test agent"
fi
echo ""

# Create test prompt
echo "🔧 Creating test prompt..."
PROMPT_RESPONSE=$(curl -s -X POST http://localhost:8000/prompts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": {
      "name": "test-prompt",
      "description": "Test prompt for integration tests",
      "template": "Hello {{name}}, this is a test prompt.",
      "arguments": [
        {
          "name": "name",
          "description": "Name to greet",
          "required": true
        }
      ],
      "tags": ["test", "integration"]
    }
  }')

if [ $? -eq 0 ]; then
  PROMPT_ID=$(echo "$PROMPT_RESPONSE" | jq -r '.id // empty')

  if [ -n "$PROMPT_ID" ] && [ "$PROMPT_ID" != "null" ]; then
    echo "$PROMPT_ID" > "$PROJECT_ROOT/tmp/contextforge-test-prompt-id.txt"
    echo "✅ Test prompt created successfully"
    echo "   Prompt ID: $PROMPT_ID"
  else
    echo "⚠️  Failed to extract prompt ID from response"
    echo "   Response: $PROMPT_RESPONSE"
  fi
else
  echo "⚠️  Failed to create test prompt"
fi
echo ""

echo "✨ Test environment is ready for integration tests!"
