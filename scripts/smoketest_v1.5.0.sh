#!/usr/bin/env bash

# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

# smoketest_v1.5.0.sh - KafScale v1.5.0 Release Smoketest Runner
#
# Exit codes:
#   0 - All tests passed
#   1 - One or more tests failed
#   2 - Setup/configuration error

set -uo pipefail

# Configuration
K6_BIN="${K6_BIN:-./k6}"
K6_PROFILE="${K6_PROFILE:-local-service}"
K6_TARGET="${K6_TARGET:-kafscale}"
REPORT_FILE="${REPORT_FILE:-smoketest-v1.5.0-results.json}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results tracking
declare -a TEST_NAMES=()
declare -a TEST_RESULTS=()
declare -a TEST_DURATIONS=()
TOTAL_PASSED=0
TOTAL_FAILED=0

# Print colored output
print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_phase() {
    echo ""
    echo -e "${YELLOW}--- $1 ---${NC}"
}

print_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Run a single test and track results
run_test() {
    local test_name="$1"
    local test_file="$2"
    local start_time
    local end_time
    local duration
    local exit_code

    print_info "Running: $test_name"

    start_time=$(date +%s)

    if K6_PROFILE="$K6_PROFILE" K6_TARGET="$K6_TARGET" "$K6_BIN" run "$PROJECT_ROOT/$test_file" 2>&1; then
        exit_code=0
    else
        exit_code=$?
    fi

    end_time=$(date +%s)
    duration=$((end_time - start_time))

    TEST_NAMES+=("$test_name")
    TEST_DURATIONS+=("${duration}s")

    if [ $exit_code -eq 0 ]; then
        TEST_RESULTS+=("pass")
        print_pass "$test_name (${duration}s)"
        ((TOTAL_PASSED++))
        return 0
    else
        TEST_RESULTS+=("fail")
        print_fail "$test_name (${duration}s)"
        ((TOTAL_FAILED++))
        return 1
    fi
}

# Write JSON results file
write_results() {
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    {
        echo "{"
        echo "  \"version\": \"v1.5.0\","
        echo "  \"timestamp\": \"$timestamp\","
        echo "  \"profile\": \"$K6_PROFILE\","
        echo "  \"target\": \"$K6_TARGET\","
        echo "  \"summary\": {"
        echo "    \"total\": $((TOTAL_PASSED + TOTAL_FAILED)),"
        echo "    \"passed\": $TOTAL_PASSED,"
        echo "    \"failed\": $TOTAL_FAILED"
        echo "  },"
        echo "  \"tests\": ["

        local first=true
        for i in "${!TEST_NAMES[@]}"; do
            if [ "$first" = true ]; then
                first=false
            else
                echo ","
            fi
            printf "    {\"name\": \"%s\", \"result\": \"%s\", \"duration\": \"%s\"}" \
                "${TEST_NAMES[$i]}" "${TEST_RESULTS[$i]}" "${TEST_DURATIONS[$i]}"
        done

        echo ""
        echo "  ]"
        echo "}"
    } > "$REPORT_FILE"

    print_info "Results written to: $REPORT_FILE"
}

# Print summary table
print_summary() {
    echo ""
    print_header "SMOKETEST SUMMARY"
    echo ""
    printf "%-45s %-10s %s\n" "TEST" "RESULT" "DURATION"
    printf "%-45s %-10s %s\n" "----" "------" "--------"

    for i in "${!TEST_NAMES[@]}"; do
        local result_color
        if [ "${TEST_RESULTS[$i]}" = "pass" ]; then
            result_color="${GREEN}PASS${NC}"
        else
            result_color="${RED}FAIL${NC}"
        fi
        printf "%-45s %-10b %s\n" "${TEST_NAMES[$i]}" "$result_color" "${TEST_DURATIONS[$i]}"
    done

    echo ""
    echo -e "Total: $((TOTAL_PASSED + TOTAL_FAILED)) | ${GREEN}Passed: $TOTAL_PASSED${NC} | ${RED}Failed: $TOTAL_FAILED${NC}"
}

# Verify prerequisites
check_prerequisites() {
    print_phase "Prerequisites Check"

    if [ ! -x "$K6_BIN" ]; then
        print_fail "k6 binary not found or not executable: $K6_BIN"
        echo "Build with: xk6 build --with github.com/mostafa/xk6-kafka"
        exit 2
    fi
    print_pass "k6 binary found: $K6_BIN"

    if [ ! -d "$PROJECT_ROOT/tests/k6" ]; then
        print_fail "Test directory not found: $PROJECT_ROOT/tests/k6"
        exit 2
    fi
    print_pass "Test directory found"

    print_info "Profile: $K6_PROFILE"
    print_info "Target: $K6_TARGET"
}

# Main execution
main() {
    local overall_exit=0

    print_header "KafScale v1.5.0 Smoketest Suite"
    echo "Started: $(date)"
    echo ""

    check_prerequisites

    # Phase 1: Pre-flight (S1, S2)
    print_phase "Phase 1: Pre-flight (Connectivity & Metrics)"

    if ! run_test "S1: Connectivity (diagnose)" "tests/k6/diagnose.js"; then
        print_fail "Pre-flight failed - connectivity issue. Aborting."
        write_results
        print_summary
        exit 1
    fi

    # S2 metrics test - continue even if it fails (non-critical)
    run_test "S2: Metrics endpoint" "tests/k6/smoke_metrics.js" || true

    # Phase 2: Core Functionality (S3 basic)
    print_phase "Phase 2: Core Functionality (Basic Produce/Consume)"

    if ! run_test "S3: Single message" "tests/k6/smoke_single.js"; then
        print_fail "Core functionality failed - basic produce/consume broken."
        overall_exit=1
    fi

    run_test "S3: Topic auto-create" "tests/k6/smoke_topic_autocreate.js" || overall_exit=1

    # Phase 3: Concurrency (S3 advanced)
    print_phase "Phase 3: Concurrency Tests"

    run_test "S3: Concurrent VUs" "tests/k6/smoke_concurrent.js" || overall_exit=1
    run_test "S3: Shared connection" "tests/k6/smoke_shared.js" || overall_exit=1
    run_test "S3: Multi-producer single-consumer" "tests/k6/smoke_multi_producer_single_consumer.js" || overall_exit=1
    run_test "S3: Consumer group" "tests/k6/smoke_consumer_group.js" || overall_exit=1

    # Phase 4: v1.5.0 Features (ACL)
    print_phase "Phase 4: v1.5.0 Features (ACL)"

    if [ -f "$PROJECT_ROOT/tests/k6/smoke_acl_basic.js" ]; then
        run_test "S8: ACL basic (v1.5.0)" "tests/k6/smoke_acl_basic.js" || overall_exit=1
    else
        print_info "ACL test not found - skipping (tests/k6/smoke_acl_basic.js)"
    fi

    # Write results and summary
    write_results
    print_summary

    echo ""
    if [ $overall_exit -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
    else
        echo -e "${RED}Some tests failed. Review results above.${NC}"
    fi

    echo "Completed: $(date)"
    exit $overall_exit
}

main "$@"
