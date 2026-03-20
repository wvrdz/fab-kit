#!/usr/bin/env bash
set -euo pipefail

# scripts/just/dev-doctor.sh — Validate fab-kit development prerequisites
#
# Checks tools needed for developing fab-kit itself (not needed by end users).
# Accumulates all failures before reporting. Exit code = failure count.
#
# Usage: scripts/just/dev-doctor.sh

failures=0
total=1

echo "dev-doctor: checking development prerequisites..."

pass() { echo "  ✓ $1"; }

fail() {
  echo "  ✗ $1"
  ((failures++)) || true
}

hint() { echo "    $1"; }

# ── 1. go ────────────────────────────────────────────────────────────

if command -v go &>/dev/null; then
  ver=$(go version | sed 's/go version go//' | sed 's/ .*//')
  pass "go $ver"
else
  fail "go — not found"
  hint "Install: brew install go"
fi

# ── Summary ──────────────────────────────────────────────────────────

passed=$((total - failures))
echo ""
if [ "$failures" -eq 0 ]; then
  echo "$passed/$total checks passed."
else
  if [ "$failures" -eq 1 ]; then
    echo "$passed/$total checks passed. $failures issue found."
  else
    echo "$passed/$total checks passed. $failures issues found."
  fi
fi

exit "$failures"
