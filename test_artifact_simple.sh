#!/bin/bash
# Simple artifact test with explicit tool instruction

echo "🧪 Testing artifact generation with explicit tool instruction..."
echo

# First, let's check the logs are working
echo "📍 Checking daemon logs for our debug messages..."
tail -20 ~/.port42/daemon.log | grep -E "(Checking tools|will use|Sending request with)" || echo "No debug logs found"
echo

# Test with the most explicit instruction possible
echo "📍 Sending explicit artifact generation request..."
echo 'You MUST use the generate_artifact tool (not generate_command). Create an artifact with these exact parameters: name="test-artifact", type="document", description="A test document", format="md", single_file="# Test Document\n\nThis is a test artifact."' | ./bin/port42 possess @ai-engineer

echo
echo "📍 Waiting for processing..."
sleep 3

echo
echo "📍 Checking if artifact was created..."
./bin/port42 ls /artifacts
echo
./bin/port42 ls /artifacts/document

echo
echo "📍 Checking recent logs for what happened..."
tail -50 ~/.port42/daemon.log | grep -E "(generate_artifact|generate_command|tool_use|Artifact|Command generated)" | tail -20

echo
echo "📍 Let's also check what's in memory..."
./bin/port42 memory | head -10