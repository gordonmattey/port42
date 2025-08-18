#!/bin/bash

# Port 42 Manual Test Runner
# Provides easy access to run specific test sections

set -e

TEST_SUITE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/manual-test-suite.sh"

show_usage() {
    echo "Port 42 Manual Test Runner"
    echo ""
    echo "Usage: $0 [test-section]"
    echo ""
    echo "Available test sections:"
    echo "  all              - Run complete test suite (default)"
    echo "  basic            - Basic setup and simple tools"
    echo "  references       - All reference type tests" 
    echo "  prompts          - Custom prompt functionality"
    echo "  advanced         - Combined prompts + references"
    echo "  debug            - Debug mode verification"
    echo "  vfs              - Virtual filesystem navigation"
    echo "  errors           - Error handling tests"
    echo ""
    echo "Examples:"
    echo "  $0                # Run all tests"
    echo "  $0 basic          # Just basic functionality"
    echo "  $0 references     # Test all reference types"
    echo "  $0 advanced       # Ultimate combined tests"
    echo ""
}

run_basic_tests() {
    echo "🧪 Running Basic Tests..."
    source "$TEST_SUITE"
    create_test_files
    test_basic_setup
    test_simple_tool_declaration
}

run_reference_tests() {
    echo "🧪 Running Reference Tests..."
    source "$TEST_SUITE"
    create_test_files
    test_basic_setup
    test_file_reference
    test_multiple_file_references
    test_url_reference
    test_p42_vfs_reference
    test_search_reference
}

run_prompt_tests() {
    echo "🧪 Running Prompt Tests..."
    source "$TEST_SUITE"
    create_test_files
    test_basic_setup
    test_custom_prompt
    test_artifact_creation
}

run_advanced_tests() {
    echo "🧪 Running Advanced Tests..."
    source "$TEST_SUITE"
    create_test_files
    test_basic_setup
    test_simple_tool_declaration  # Need some tools for P42 refs
    test_combined_prompt_references
}

run_debug_tests() {
    echo "🧪 Running Debug Tests..."
    source "$TEST_SUITE"
    test_debug_mode
}

run_vfs_tests() {
    echo "🧪 Running VFS Tests..."
    source "$TEST_SUITE"
    create_test_files
    test_basic_setup
    test_simple_tool_declaration
    test_vfs_navigation
}

run_error_tests() {
    echo "🧪 Running Error Handling Tests..."
    source "$TEST_SUITE"
    test_error_handling
}

# Parse command line arguments
case "${1:-all}" in
    "all")
        exec "$TEST_SUITE"
        ;;
    "basic")
        run_basic_tests
        ;;
    "references"|"refs")
        run_reference_tests
        ;;
    "prompts")
        run_prompt_tests
        ;;
    "advanced")
        run_advanced_tests
        ;;
    "debug")
        run_debug_tests
        ;;
    "vfs")
        run_vfs_tests
        ;;
    "errors")
        run_error_tests
        ;;
    "help"|"-h"|"--help")
        show_usage
        ;;
    *)
        echo "Unknown test section: $1"
        echo ""
        show_usage
        exit 1
        ;;
esac