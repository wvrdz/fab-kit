# Quality Checklist: Brew Install System Shim

**Change**: 260325-lhhk-brew-install-system-shim
**Generated**: 2026-03-27
**Spec**: `spec.md`

## Functional Completeness

- [x] CHK-001 Config Discovery: Shim walks up from CWD to find `fab/project/config.yaml`
- [x] CHK-002 Version Resolution: Shim reads `fab_version` from config and resolves to cached version
- [x] CHK-003 Version Download: Shim downloads and caches missing versions from GitHub releases
- [x] CHK-004 Argument Passthrough: All args pass verbatim to per-repo runtime via exec
- [x] CHK-005 fab init (fresh): Creates config.yaml with fab_version in a new project
- [x] CHK-006 fab init (existing config): Adds fab_version without overwriting other fields
- [x] CHK-007 fab init (already initialized): Reports status without modifying config
- [x] CHK-008 Homebrew formula: Builds shim, wt, and idea from source
- [x] CHK-009 fab_version config field: Optional field parsed correctly by shim

## Behavioral Correctness

- [x] CHK-010 No fab_version error: Shim errors with actionable message when field absent
- [x] CHK-011 No config error: Shim errors for non-init commands outside a fab repo
- [x] CHK-012 --version and --help: Work without a config/repo context
- [x] CHK-013 Direct invocation unchanged: `fab/.kit/bin/fab` works independently of shim

## Scenario Coverage

- [x] CHK-014 Config found in ancestor directory scenario
- [x] CHK-015 Version cached scenario (cache hit → exec)
- [x] CHK-016 Version not cached scenario (download → cache → exec)
- [x] CHK-017 Download failure scenario (clear error with URL)
- [x] CHK-018 Concurrent download scenario (atomic extraction)
- [x] CHK-019 Platform-specific download URL construction
- [x] CHK-020 Init with network failure scenario (no files modified)

## Edge Cases & Error Handling

- [x] CHK-021 Filesystem root reached without finding config
- [x] CHK-022 Download to temp dir + rename atomicity
- [x] CHK-023 Invalid/corrupt cache version directory
- [x] CHK-024 GitHub API rate limiting during init (latest version query)

## Code Quality

- [x] CHK-025 Pattern consistency: New code follows naming and structural patterns of existing Go modules (src/go/fab, src/go/wt, src/go/idea)
- [x] CHK-026 No unnecessary duplication: Reuses patterns from existing Go modules where applicable
- [x] CHK-027 Readability: Functions focused, no god functions (>50 lines without reason)
- [x] CHK-028 No magic strings: URLs, paths, and error messages use named constants or are clearly structured

## Documentation Accuracy

- [x] CHK-029 distribution.md updated with Homebrew model, shim architecture, cache layout
- [x] CHK-030 Homebrew formula references correct source paths and build commands

## Cross References

- [x] CHK-031 Justfile recipes consistent with existing build patterns
- [x] CHK-032 Go module structure consistent with existing src/go/ layout

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
