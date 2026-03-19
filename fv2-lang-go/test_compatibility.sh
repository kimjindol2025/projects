#!/bin/bash

# FV 2.0 Compatibility Test Suite
# Tests V language code samples for parsing compatibility

set -e

COMPILER="./bin/fv2"
TEST_DIR="./test_cases"
RESULTS_FILE="compatibility_report.txt"

if [ ! -f "$COMPILER" ]; then
    echo "Error: Compiler not found at $COMPILER"
    echo "Run: go build -o bin/fv2 ./cmd/fv2"
    exit 1
fi

if [ ! -d "$TEST_DIR" ]; then
    echo "Error: Test cases directory not found at $TEST_DIR"
    exit 1
fi

# Initialize results
> "$RESULTS_FILE"
TOTAL=0
PASSED=0
FAILED=0

echo "=== FV 2.0 Compatibility Test Report ===" | tee -a "$RESULTS_FILE"
echo "Date: $(date)" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Test each file
for testfile in "$TEST_DIR"/*.fv; do
    if [ ! -f "$testfile" ]; then
        continue
    fi

    filename=$(basename "$testfile")
    TOTAL=$((TOTAL + 1))

    # Run compiler
    if output=$("$COMPILER" "$testfile" 2>&1); then
        PASSED=$((PASSED + 1))
        status="✅ PASS"
    else
        FAILED=$((FAILED + 1))
        status="❌ FAIL"
    fi

    echo "$status - $filename" | tee -a "$RESULTS_FILE"
done

echo "" | tee -a "$RESULTS_FILE"
echo "=== Summary ===" | tee -a "$RESULTS_FILE"
echo "Total: $TOTAL" | tee -a "$RESULTS_FILE"
echo "Passed: $PASSED" | tee -a "$RESULTS_FILE"
echo "Failed: $FAILED" | tee -a "$RESULTS_FILE"

if [ $TOTAL -gt 0 ]; then
    PERCENTAGE=$((PASSED * 100 / TOTAL))
    echo "Compatibility Rate: $PERCENTAGE%" | tee -a "$RESULTS_FILE"
fi

echo ""
echo "Results saved to: $RESULTS_FILE"

if [ $FAILED -eq 0 ]; then
    echo "✅ All tests passed!"
    exit 0
else
    echo "❌ $FAILED test(s) failed"
    exit 1
fi
