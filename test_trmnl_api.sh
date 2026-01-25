#!/bin/bash

echo "🦉 Testing ExpenseOwl TRMNL API..."
echo "=================================="
echo ""

# Check if server is running
if ! curl -s http://localhost:8080/version > /dev/null 2>&1; then
    echo "❌ Server is not running on localhost:8080"
    echo "Please start the server first: ./expenseowl"
    exit 1
fi

echo "✅ Server is running"
echo ""

# Test TRMNL API
echo "📊 Fetching TRMNL data..."
echo ""

response=$(curl -s http://localhost:8080/api/trmnl)

# Check if response is valid JSON
if echo "$response" | python3 -m json.tool > /dev/null 2>&1; then
    echo "✅ Valid JSON response"
    echo ""
    echo "Response:"
    echo "$response" | python3 -m json.tool
else
    echo "❌ Invalid JSON response"
    echo "$response"
    exit 1
fi

echo ""
echo "=================================="
echo "✅ Test completed successfully!"
