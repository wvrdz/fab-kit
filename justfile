# Run all tests from src/
test:
    #!/usr/bin/env bash
    set -euo pipefail
    failed=0
    for t in src/*/test.sh; do
        echo "── ${t} ──"
        if bash "$t"; then
            echo ""
        else
            failed=1
            echo ""
        fi
    done
    exit $failed
