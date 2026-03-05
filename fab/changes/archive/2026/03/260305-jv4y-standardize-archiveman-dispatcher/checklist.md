# Quality Checklist: Standardize archiveman.sh Dispatcher Integration

**Change**: 260305-jv4y-standardize-archiveman-dispatcher
**Generated**: 2026-03-05
**Spec**: `spec.md`

## Functional Completeness
- [ ] CHK-001 Dispatcher pass-through: `fab archive` case in `fab/.kit/bin/fab` forwards args without injecting `"archive"`
- [ ] CHK-002 Default-to-archive fallback: archiveman.sh `*` case delegates to `cmd_archive "$@"`
- [ ] CHK-003 Explicit `archive` subcommand: `archiveman.sh archive <change> --description "..."` still works

## Behavioral Correctness
- [ ] CHK-004 `fab archive <change> --description "..."` works end-to-end via shell backend
- [ ] CHK-005 `fab archive restore <change>` works end-to-end via shell backend
- [ ] CHK-006 `fab archive list` works end-to-end via shell backend
- [ ] CHK-007 Empty args: `archiveman.sh` (no args) still errors

## Scenario Coverage
- [ ] CHK-008 Parity test "archive change" exercises default-to-archive path (no explicit `archive` subcommand)
- [ ] CHK-009 Parity test "list empty" still passes
- [ ] CHK-010 Bats tests pass without modification

## Edge Cases & Error Handling
- [ ] CHK-011 Unknown subcommand (e.g., `delete`) now triggers archive fallback instead of error — verify this is acceptable
- [ ] CHK-012 Help flag (`--help`, `-h`) still works

## Code Quality
- [ ] CHK-013 Pattern consistency: dispatcher `archive)` case matches other cases (plain pass-through)
- [ ] CHK-014 No unnecessary duplication: no duplicated logic in fallback path

## Documentation Accuracy
- [ ] CHK-015 Memory file `docs/memory/fab-workflow/kit-architecture.md` updated to reflect standardized pass-through

## Cross References
- [ ] CHK-016 `_scripts.md` archive documentation still accurate (already documents `fab archive`)

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
