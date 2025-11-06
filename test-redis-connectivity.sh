#!/bin/bash
set -e
REDIS_VERSION="7"

echo "🧪 Testing Redis ${REDIS_VERSION} connectivity for Redish application"
echo "======================================================"

# Function to cleanup on exit
cleanup() {
    echo "🧹 Cleaning up..."
    if [[ -n "${ENGINE_CMD:-}" ]]; then
        $ENGINE_CMD stop test-redis 2>/dev/null || true
        $ENGINE_CMD rm -f test-redis 2>/dev/null || true
    else
        docker stop test-redis 2>/dev/null || true
        docker rm -f test-redis 2>/dev/null || true
    fi
}

trap cleanup EXIT

echo "Detecting container engine..."
if command -v docker >/dev/null 2>&1; then
    ENGINE_CMD="docker"
elif command -v podman >/dev/null 2>&1; then
    ENGINE_CMD="podman"
else
    echo "Neither docker nor podman found. Please install one to run tests." >&2
    exit 1
fi

echo "Starting Redis container..."
$ENGINE_CMD run -d --name test-redis -p 6379:6379 redis:${REDIS_VERSION}-alpine
sleep 5

echo "Testing basic connectivity..."
if $ENGINE_CMD exec test-redis redis-cli ping | grep -q "PONG"; then
    echo "✅ Redis is responding to PING"
else
    echo "❌ Redis is not responding to PING"
    exit 1
fi

echo "Running Go tests..."
if go test -v -run TestRedisConnection ./...; then
    echo "✅ Go tests passed"
else
    echo "❌ Go tests failed"
    exit 1
fi

echo "Testing application..."
if timeout 10s go run main.go -uri localhost:6379 -commands "PING;SET test:script success;GET test:script"; then
    echo "✅ Application test passed"
else
    echo "❌ Application test failed"
    exit 1
fi
echo "✅ All tests completed successfully"