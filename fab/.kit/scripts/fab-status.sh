#!/usr/bin/env bash
set -euo pipefail

current_file="$(dirname "$0")/../../current"

if [ ! -f "$current_file" ]; then
  echo "No active change"
  exit 0
fi

name=$(cat "$current_file")
status_file="$(dirname "$0")/../../changes/$name/.status.yaml"

if [ ! -f "$status_file" ]; then
  echo "Active: $name (missing — run /fab:switch or /fab:new)"
  exit 1
fi

stage=$(grep '^stage:' "$status_file" | cut -d' ' -f2)
branch=$(grep '^branch:' "$status_file" | cut -d' ' -f2 || true)
if [ -n "$branch" ]; then
  echo "Active: $name (stage: $stage, branch: $branch)"
else
  echo "Active: $name (stage: $stage)"
fi
