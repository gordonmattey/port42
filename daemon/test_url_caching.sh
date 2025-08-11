#!/bin/bash

# Test script for URL artifact caching (Phase 2)
set -e

echo "🧪 Testing URL Artifact Caching - Phase 2: Artifact Management"
echo "============================================================="

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test URLs
TEST_URL_1="https://httpbin.org/json"
TEST_URL_2="https://httpbin.org/uuid"
TEST_URL_INVALID="not-a-url"

echo -e "\n${BLUE}Test 1: Fresh URL fetch (cache miss)${NC}"
echo "Testing first fetch of $TEST_URL_1"
../bin/port42 declare tool url-test-1 --ref url:$TEST_URL_1
echo "✅ Check daemon logs for: '🌐 URL cache miss, fetching' and '💾 Cached artifact'"

echo -e "\n${BLUE}Test 2: Cached URL fetch (cache hit)${NC}"
echo "Testing second fetch of same URL (should hit cache)"
../bin/port42 declare tool url-test-2 --ref url:$TEST_URL_1
echo "✅ Check daemon logs for: '🎯 Cache hit for artifact' and '[Cached from YYYY-MM-DD]'"

echo -e "\n${BLUE}Test 3: Different URL (cache miss)${NC}"
echo "Testing different URL $TEST_URL_2"
../bin/port42 declare tool url-test-3 --ref url:$TEST_URL_2
echo "✅ Should see fresh fetch for new URL"

echo -e "\n${BLUE}Test 4: Invalid URL handling${NC}"
echo "Testing invalid URL: $TEST_URL_INVALID"
../bin/port42 declare tool url-test-4 --ref url:$TEST_URL_INVALID || true
echo "✅ Should gracefully handle invalid URL without crashing"

echo -e "\n${BLUE}Test 5: Check stored Relations in daemon logs${NC}"
echo "URL artifacts are stored as Relations - check daemon logs for:"
echo "  - '🌟 Declaring relation: url-artifact-XXXXX (type: URLArtifact)'"
echo "  - '✅ Relation stored: url-artifact-XXXXX'"
echo "  - '💾 Cached artifact url-artifact-XXXXX' (if working correctly)"

echo -e "\n${YELLOW}Manual Verification:${NC}"
echo "1. Check daemon logs for cache hit/miss indicators"
echo "2. Verify [Freshly fetched] vs [Cached from DATE] in tool output"
echo "3. Confirm Relations storage contains URLArtifact entries"
echo "4. Test that same URL gives cache hit consistently"

echo -e "\n${GREEN}Test completion! Check the above indicators to verify caching works.${NC}"
echo -e "\n${YELLOW}Expected Behavior (FIXED):${NC}"
echo "✅ First URL fetch:"
echo "   - '🌐 URL cache MISS: ... -> fetching fresh'"
echo "   - '💾 Cached URL artifact: url-artifact-XXXX (URL, SIZE bytes)'"
echo "   - '📊 Data-only relation stored: ... (type: URLArtifact)'"
echo "   - Tool output shows: '[Freshly fetched]'"
echo ""
echo "✅ Second URL fetch (same URL):"
echo "   - '✅ Cache VALID: ... (age: Xs, TTL: 24h0m0s)'"
echo "   - '🎯 URL cache HIT: ... -> url-artifact-XXXX'"
echo "   - Tool output shows: '[Cached from YYYY-MM-DD HH:MM:SS]'"
echo ""
echo "✅ No more duplicate resolutions or confusing warnings!"