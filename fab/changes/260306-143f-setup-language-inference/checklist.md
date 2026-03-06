# Quality Checklist: Replace Template-Driven Language Detection with Agent-Inferred Conventions

**Change**: 260306-143f-setup-language-inference
**Generated**: 2026-03-06
**Spec**: `spec.md`

## Functional Completeness
- [x] CHK-001 Agent inference flow: Step 1b-lang in `fab-setup.md` describes the three-phase Detection → Inference → Write flow
- [x] CHK-002 Detection table: Marker files listed for Rust, TypeScript, Node.js, Go, Python with framework signals
- [x] CHK-003 Inference phase: Skill describes reading detected files and using agent knowledge to derive conventions
- [x] CHK-004 Adaptive questions: Up to 3 free-form questions described, with adaptive skip logic
- [x] CHK-005 Write phase routing: Convention content routed to correct `fab/project/*` files by content type
- [x] CHK-006 Template deletion: `fab/.kit/templates/constitutions/` and `fab/.kit/templates/configs/` directories removed
- [x] CHK-007 Advisory removal: Section 2b removed from `fab/.kit/sync/2-sync-workspace.sh`

## Behavioral Correctness
- [x] CHK-008 No template references: No remaining references to `fab/.kit/templates/constitutions/` or `fab/.kit/templates/configs/` in `fab-setup.md`
- [x] CHK-009 No template references in sync: No remaining references to template constitutions/configs in `2-sync-workspace.sh`
- [x] CHK-010 Sync script integrity: Sections 1, 2 (scaffold), 3 (skills), 4 (cleanup), 5 (version stamp) in `2-sync-workspace.sh` remain unchanged

## Removal Verification
- [x] CHK-011 Constitution templates gone: `fab/.kit/templates/constitutions/` directory does not exist
- [x] CHK-012 Config templates gone: `fab/.kit/templates/configs/` directory does not exist
- [x] CHK-013 Advisory block gone: No "Language template advisory" comment block in `2-sync-workspace.sh`
- [x] CHK-014 Artifact templates preserved: `fab/.kit/templates/` still contains `intake.md`, `spec.md`, `tasks.md`, `checklist.md`, `status.yaml`

## Scenario Coverage
- [x] CHK-015 Rust detection scenario: Skill text covers Cargo.toml detection and Rust-specific config scanning
- [x] CHK-016 TypeScript+React detection scenario: Skill text covers combined TS+React detection with layered conventions
- [x] CHK-017 No language scenario: Skill text handles no marker files found (silent skip)
- [x] CHK-018 Monorepo scenario: Skill text addresses multiple languages detected
- [x] CHK-019 Re-run idempotency scenario: Skill text describes reading existing content, merging without duplication

## Edge Cases & Error Handling
- [x] CHK-020 Missing config files: Detection gracefully handles missing optional configs (e.g., no clippy.toml)
- [x] CHK-021 SRAD exclusion: Skill text explicitly states SRAD does not apply to setup question flow

## Code Quality
- [x] CHK-022 Pattern consistency: New step 1b-lang follows the same structural patterns as other setup steps (idempotency guard, conditional execution)
- [x] CHK-023 No unnecessary duplication: No duplicated detection logic between skill and sync script

## Documentation Accuracy
- [x] CHK-024 Constitution §V compliance: No language-specific content bundled in `fab/.kit/`
- [x] CHK-025 Skill describes process not content: Step 1b-lang describes what to detect/read/write, not hard-coded convention text

## Cross References
- [x] CHK-026 No dangling references: No other files in `fab/.kit/` reference the deleted template directories
