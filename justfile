scripts := "src/scripts/just"

# Run all tests (bash + rust) with summary
test:
    just test-setup
    just _test-parallel
    # just test-rust  # uncomment when Rust libs exist

# Run test suites in parallel, fail if any fail
_test-parallel:
    #!/usr/bin/env bash
    set -uo pipefail
    pids=()
    logs=()
    suites=(test-bash test-packages test-scripts)
    for s in "${suites[@]}"; do
        log=$(mktemp)
        logs+=("$log")
        just "$s" >"$log" 2>&1 &
        pids+=($!)
    done
    failed=0
    for i in "${!pids[@]}"; do
        if ! wait "${pids[$i]}"; then
            failed=1
        fi
        cat "${logs[$i]}"
        rm -f "${logs[$i]}"
    done
    exit $failed

# Setup test dependencies
test-setup:
    {{scripts}}/test-setup.sh

# Run bash tests (bats)
test-bash:
    {{scripts}}/test-bash.sh

# Run package tests (bats)
test-packages:
    {{scripts}}/test-packages.sh

# Run script tests (bats)
test-scripts:
    {{scripts}}/test-scripts.sh

# Run Rust tests (placeholder)
test-rust:
    @echo "No Rust tests yet."

# Check prerequisites and environment health
doctor:
    fab/.kit/scripts/fab-doctor.sh
