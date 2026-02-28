#!/usr/bin/env bash
set -uo pipefail

pids=()
suites=(test-bash test-packages test-scripts)
tags=(bash pkgs scripts)

for i in "${!suites[@]}"; do
    just "${suites[$i]}" 2>&1 | sed "s/^/[${tags[$i]}] /" &
    pids+=($!)
done

failed=0
for pid in "${pids[@]}"; do
    if ! wait "$pid"; then failed=1; fi
done
exit $failed
