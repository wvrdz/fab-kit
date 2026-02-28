scripts := "src/scripts/just"

# Run all tests (bash + rust) with summary
test:
    just test-setup
    just _test-parallel
    # just test-rust  # uncomment when Rust libs exist

# Run test suites in parallel with prefixed live output
_test-parallel:
    {{scripts}}/test-parallel.sh

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
