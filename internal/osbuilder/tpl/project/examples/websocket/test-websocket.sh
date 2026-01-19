
#!/bin/bash

# WebSocket Client Test Script
set -e

SERVER_URL=${SERVER_URL:-"ws://localhost:5555/v1/ws"}
CLIENT_BIN=${CLIENT_BIN:-"/tmp/ws-client"}

echo "=== WebSocket Client Test Suite ==="
echo "Server URL: $SERVER_URL"

# Function to run test and capture output
run_test() {
    local test_name="$1"
    local test_args="$2"
    local expected_duration="$3"
    
    echo ""
    echo "Running: $test_name"
    echo "Command: $CLIENT_BIN $test_args"
    echo "Expected duration: ${expected_duration}s"
    echo "----------------------------------------"
    
    start_time=$(date +%s)
    $CLIENT_BIN $test_args
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    
    echo "Actual duration: ${duration}s"
    echo "✅ $test_name completed"
}

# Build client first
echo "Building WebSocket test client..."
go build -o /tmp/ws-client ./ws-client.go

# Test 1: Basic connectivity
run_test "Basic Connectivity Test" \
    "-url=$SERVER_URL -ping=2s -duration=10s" \
    10

# Test 2: Multiple clients
run_test "Multiple Clients Test" \
    "-url=$SERVER_URL -clients=3 -ping=3s -duration=15s" \
    15

# Test 3: Fast ping test
run_test "Fast Ping Test" \
    "-url=$SERVER_URL -user-prefix=fast_ping -ping=500ms -duration=10s" \
    10

# Test 4: Error handling test
run_test "Error Handling Test" \
    "-url=$SERVER_URL -user-prefix=error_test -ping=5s -duration=15s -invalid=true" \
    15

# Test 5: Load test (many clients)
run_test "Load Test" \
    "-url=$SERVER_URL -clients=20 -ping=2s -duration=20s" \
    20

# Test 6: Long duration test
run_test "Long Duration Test" \
    "-url=$SERVER_URL -user-prefix=long_test -ping=1s -duration=30s" \
    30

echo ""
echo "=== All tests completed successfully! ==="

# Optional: Generate test report
echo ""
echo "=== Test Summary ==="
echo "Total test scenarios: 6"
echo "All tests passed ✅"
echo "Server URL tested: $SERVER_URL"
echo "Test completed at: $(date)"

