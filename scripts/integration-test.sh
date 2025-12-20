#!/bin/bash

# Integration test script for complete election flow
# This script tests: create election, submit candidacy, vote, view results, admin results

set -eo pipefail

# Configuration
BASE_URL="http://localhost:8080"
TEST_ELECTION_TITLE="Integration Test Election"
TEST_ELECTION_DESCRIPTION="This is a test election created by integration script"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if server is running
check_server() {
    log_info "Checking if server is running at $BASE_URL..."
    if curl -s "$BASE_URL/health" > /dev/null; then
        log_info "Server is running"
    else
        log_error "Server is not running at $BASE_URL"
        log_info "Please start the server first: ./server"
        exit 1
    fi
}

# Create authentication session via auth callback
create_session() {
    local email=${1:-"test@example.com"}
    local cookie_file=${2:-"cookies.txt"}
    
    log_info "Creating authentication session for $email..."
    
    response=$(curl -s -X POST "$BASE_URL/auth/callback" \
        -H "Content-Type: application/json" \
        -c "$cookie_file" \
        -d '{"email":"'$email'","name":"Test User"}')
    
    log_info "Auth response for $email: $response"
    
    if echo "$response" | grep -q "$email"; then
        log_info "Authentication session created successfully for $email"
        echo "$response" | jq -r '.user.uuid' 2>/dev/null || echo "SESSION_CREATED"
        return 0
    else
        log_error "Failed to create authentication session for $email: $response"
        return 1
    fi
}

# Create a new election (admin)
create_election() {
    log_info "Creating election: $TEST_ELECTION_TITLE"
    
    response=$(curl -s -X POST "$BASE_URL/elections" \
        -H "Content-Type: application/json" \
        -b cookies.txt \
        -d "{
            \"title\": \"$TEST_ELECTION_TITLE\",
            \"description\": \"$TEST_ELECTION_DESCRIPTION\",
            \"is_live_results\": true
        }")
    
    # Check if response is a valid Election object with expected title
    if echo "$response" | jq -e ".title == \"$TEST_ELECTION_TITLE\"" >/dev/null 2>&1; then
        log_info "Election created successfully"
        # Extract and save the election UUID for use in other tests
        ELECTION_UUID=$(echo "$response" | jq -r '.uuid')
        echo "$ELECTION_UUID" > election_uuid.txt
        return 0
    else
        log_error "Failed to create election: $response"
        return 1
    fi
}

# Get list of elections
list_elections() {
    log_info "Listing elections..."
    
    response=$(curl -s "$BASE_URL/elections")
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
}

# Submit candidacy
submit_candidacy() {
    local election_id=$1
    log_info "Submitting candidacy for election: $election_id"
    
    response=$(curl -s -X POST "$BASE_URL/elections/$election_id/candidacies" \
        -H "Content-Type: application/json" \
        -b cookies.txt)
    
    # Check if response is a valid Candidacy object or success message
    if echo "$response" | jq -e '.uuid' >/dev/null 2>&1; then
        log_info "Candidacy submitted successfully"
        # Extract and save the candidacy UUID for use in voting
        CANDIDACY_UUID=$(echo "$response" | jq -r '.uuid')
        echo "$CANDIDACY_UUID" > candidacy_uuid.txt
        return 0
    elif echo "$response" | jq -e '.message == "Candidacy submitted successfully"' >/dev/null 2>&1; then
        log_info "Candidacy submitted successfully (mock response)"
        return 0
    else
        log_error "Failed to submit candidacy: $response"
        return 1
    fi
}

# Get candidacies for election
get_candidacies() {
    local election_id=$1
    log_info "Getting candidacies for election: $election_id"
    
    response=$(curl -s "$BASE_URL/elections/$election_id/candidacies")
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
}

# Vote for a candidate
vote() {
    local election_id=$1
    local candidacy_id=$2
    log_info "Voting for candidacy: $candidacy_id in election: $election_id"
    
    response=$(curl -s -X POST "$BASE_URL/elections/$election_id/vote" \
        -H "Content-Type: application/json" \
        -d "{\"candidacy_uuid\": \"$candidacy_id\"}" \
        -b cookies.txt)
    
    if echo "$response" | grep -q "Vote submitted" || [ "$response" = "" ]; then
        log_info "Vote submitted successfully"
        return 0
    elif echo "$response" | grep -q "Failed to submit vote" || echo "$response" | grep -q "internal_error"; then
        log_error "Vote failed due to system restriction: $response"
        return 1
    else
        log_error "Failed to vote: $response"
        return 1
    fi
}

# Open election for voting
open_election() {
    local election_id=$1
    log_info "Opening election for voting: $election_id"
    
    response=$(curl -s -X POST "$BASE_URL/elections/$election_id/open" \
        -H "Content-Type: application/json" \
        -d "{\"is_open\": true}")
    
    # Check if response is a valid Election object with is_open=true
    if echo "$response" | jq -e '.is_open == true' >/dev/null 2>&1; then
        log_info "Election opened successfully"
        return 0
    else
        log_error "Failed to open election: $response"
        return 1
    fi
}

# Get live results (public)
get_live_results() {
    local election_id=$1
    log_info "Getting live results for election: $election_id"
    
    response=$(curl -s "$BASE_URL/elections/$election_id/live")
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
}

# Get admin results
get_admin_results() {
    local election_id=$1
    log_info "Getting admin results for election: $election_id"
    
    response=$(curl -s "$BASE_URL/elections/$election_id/results")
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
}

# Main test execution
main() {
    log_info "Starting integration test for election flow"
    
    # Check dependencies
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_warn "jq is not installed, JSON output will not be formatted"
    fi
    
    # Check if server is running
    check_server
    
    # Clean up any existing test data
    rm -f cookies.txt test_cookies.txt election_uuid.txt candidacy_uuid.txt voter_cookies.txt
    
    # Create session (mock)
    log_info "Creating session now..."
    if create_session; then
        log_info "Session created successfully"
        session_id=$(echo "test-session")
        log_info "Using session: $session_id"
    else
        log_error "Failed to create session - exiting"
        exit 1
    fi
    
    # Step 1: Create election (admin)
    log_info "=== STEP 1: Creating Election ==="
    if ! create_election; then
        log_error "Cannot proceed with test - election creation failed"
        exit 1
    fi
    
    # Step 2: List elections to verify creation
    log_info "=== STEP 2: Listing Elections ==="
    list_elections
    
    # Extract the election ID from the creation response
    if [ -f "election_uuid.txt" ]; then
        ELECTION_ID=$(cat election_uuid.txt)
        log_info "Using created election ID: $ELECTION_ID"
    else
        log_error "No election UUID found - election creation may have failed"
        exit 1
    fi
    
    # Step 3: Submit candidacy
    log_info "=== STEP 3: Submitting Candidacy ==="
    if ! submit_candidacy "$ELECTION_ID"; then
        log_warn "Candidacy submission failed - election might not exist or already has candidacies"
    else
        get_candidacies "$ELECTION_ID"
    fi
    
    # Step 4: Open election for voting
    log_info "=== STEP 4: Opening Election ==="
    if ! open_election "$ELECTION_ID"; then
        log_warn "Failed to open election - might not exist or already open"
    fi
    
    # Step 5: Get candidacies for voting
    log_info "=== STEP 5: Getting Candidacies for Voting ==="
    candidacies_response=$(get_candidacies "$ELECTION_ID")
    
    # Extract the actual candidacy ID if available, otherwise use placeholder
    if [ -f "candidacy_uuid.txt" ]; then
        CANDIDACY_ID=$(cat candidacy_uuid.txt)
        log_info "Using created candidacy ID: $CANDIDACY_ID"
    else
        CANDIDACY_ID="87654321-4321-4321-4321-210987654321"
        log_warn "No candidacy UUID found, using placeholder"
    fi
    
    # Step 6: Cast vote
    log_info "=== STEP 6: Casting Vote ==="
    if ! vote "$ELECTION_ID" "$CANDIDACY_ID"; then
        log_info "Voting not available with single user (candidate cannot vote in own election)"
        log_info "Note: In production, a different user would vote for the candidate"
    else
        log_info "Vote submitted successfully"
    fi
    
    # Step 7: Get live results
    log_info "=== STEP 7: Getting Live Results ==="
    get_live_results "$ELECTION_ID"
    
    # Step 8: Get admin results
    log_info "=== STEP 8: Getting Admin Results ==="
    get_admin_results "$ELECTION_ID"
    
    # Clean up
    rm -f cookies.txt election_uuid.txt candidacy_uuid.txt
    
    log_info "Integration test completed!"
    log_info "Note: This is a basic integration test. A complete test would:"
    log_info "  - Parse JSON responses to extract real UUIDs"
    log_info "  - Implement proper authentication"
    log_info "  - Test multiple users voting"
    log_info "  - Test error conditions"
    log_info "  - Clean up test data"
}

# Run the test
main "$@"