# Tasks: Brew Install System Shim

**Change**: 260325-lhhk-brew-install-system-shim
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Setup

- [x] T001 Create Go module scaffold at `src/go/shim/` — `go.mod` (module `github.com/wvrdz/fab-kit/src/go/shim`), `cmd/main.go` entry point, `internal/` package directory. Dependency: `gopkg.in/yaml.v3` only. No Cobra.
- [x] T002 [P] Add justfile recipe `build-shim` — compile `src/go/shim/cmd` to a local output path (e.g., `.release-build/fab-shim` or a developer-chosen location). Add `build-shim-target` for cross-compilation matching the `_build-binary` pattern.

## Phase 2: Core Implementation

- [x] T003 Implement config discovery in `src/go/shim/internal/config.go` — walk up from CWD checking for `fab/project/config.yaml` at each level. Return the parsed config (including `fab_version`) or a "not found" result. Use `gopkg.in/yaml.v3` for parsing.
- [x] T004 Implement version resolution and cache lookup in `src/go/shim/internal/cache.go` — given a version string, check if `~/.fab-kit/versions/{version}/` exists and contains `fab/.kit/bin/fab`. Return the resolved runtime path or "not cached" result. Cache root: `~/.fab-kit/versions/`.
- [x] T005 Implement version download in `src/go/shim/internal/download.go` — download `kit-{os}-{arch}.tar.gz` from GitHub releases (`https://github.com/wvrdz/fab-kit/releases/download/v{version}/kit-{os}-{arch}.tar.gz`). Use `runtime.GOOS`/`runtime.GOARCH` for platform detection. Extract to temp dir, then `os.Rename` to cache path for atomicity. Handle download errors with clear messages including the URL attempted.
- [x] T006 Implement `fab init` in `src/go/shim/internal/init.go` — query GitHub API for latest release (`https://api.github.com/repos/wvrdz/fab-kit/releases/latest`), download and cache that version, run `fab-sync.sh` from the cached version, create or update `fab/project/config.yaml` with `fab_version`. Handle: fresh project (create config), existing config without `fab_version` (add field), already initialized (report and exit). Handle network failure gracefully.
- [x] T007 Wire shim main in `src/go/shim/cmd/main.go` — parse args to intercept `--version`, `--help`, `init`. For all other commands: discover config → resolve version → ensure cached → exec runtime with passthrough args. Handle error cases: no config (non-repo error for non-init commands), no `fab_version` (error with guidance), download failure.

## Phase 3: Integration & Edge Cases

- [x] T008 [P] Add tests for config discovery in `src/go/shim/internal/config_test.go` — test walking up directories, finding config at various depths, missing config, config without `fab_version` field.
<!-- clarified: Added [P] marker — T008 is parallelizable with T009-T011, all are independent test tasks in the same phase -->
- [x] T009 [P] Add tests for cache operations in `src/go/shim/internal/cache_test.go` — test cache hit, cache miss, cache directory creation, version validation.
- [x] T010 [P] Add tests for download logic in `src/go/shim/internal/download_test.go` — test URL construction for different platforms, atomic extraction, error handling for network failures and missing releases.
- [x] T011 [P] Add tests for init logic in `src/go/shim/internal/init_test.go` — test fresh init, existing config update, already-initialized detection, network failure handling.
- [x] T012 Add integration test for the full shim flow in `src/go/shim/cmd/main_test.go` — test end-to-end: config discovery → version resolution → cache check → exec (mocked). Test --version and --help passthrough.

## Phase 4: Polish

- [x] T013 Create Homebrew formula file at `Formula/fab-kit.rb` — Ruby formula that builds `src/go/shim/cmd` → `fab`, `src/go/wt/cmd` → `wt`, `src/go/idea/cmd` → `idea` from source. References `wvrdz/homebrew-tap` tap. Include `test do` block.
- [x] T014 Update `docs/memory/fab-workflow/distribution.md` — add new section documenting the Homebrew distribution model, shim architecture, cache layout, version resolution, and `fab init` behavior.
- [x] T015 Update `docs/memory/fab-workflow/kit-architecture.md` — modify to reflect `wt` and `idea` moving to system-only Homebrew distribution, the shim dispatcher model, and the `src/go/shim/` module addition.
<!-- clarified: Added T015 — spec metadata lists kit-architecture.md as a modify target but no task covered it -->

---

## Execution Order

- T001 blocks T003-T007 (module scaffold needed first)
- T003 blocks T007 (config discovery used by main)
- T004 blocks T007 (cache lookup used by main)
- T005 blocks T007 (download used by main)
- T006 blocks T007 (init used by main)
- T003-T006 are independent of each other, can run in parallel
- T008-T012 depend on their respective implementation tasks
- T013 depends on T007 (needs working shim to reference in formula)
- T014 is independent of implementation
- T015 is independent of implementation
