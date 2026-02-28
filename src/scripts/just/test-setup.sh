#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cd "$REPO_ROOT"

echo "Initializing shared test dependencies..."
git submodule update --init --recursive src/packages/tests/libs/

echo "Done. Run 'just test-packages' to verify."
