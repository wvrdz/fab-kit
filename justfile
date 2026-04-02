scripts := "scripts/just"
fab_version := `cat fab/.kit/VERSION`
fab_ldflags := "-X main.version=" + fab_version
shim_ldflags := "-X main.version=" + fab_version

# ── Development ──────────────────────────────────────────────────────

# Run all tests
test:
    cd src/go/fab && go test ./... -count=1
    cd src/go/idea && go test ./... -count=1
    cd src/go/wt && go test ./... -count=1
    cd src/go/shim && go test ./... -count=1

# Run all tests (verbose)
test-v:
    cd src/go/fab && go test ./... -v -count=1
    cd src/go/idea && go test ./... -v -count=1
    cd src/go/wt && go test ./... -v -count=1
    cd src/go/shim && go test ./... -v -count=1

local_cache := env("HOME", "") / ".fab-kit/local-versions" / fab_version

# Build fab-go + copy .kit content to local cache (shim auto-discovers)
build:
    mkdir -p {{local_cache}}/kit
    cd src/go/fab && CGO_ENABLED=0 go build -ldflags '{{fab_ldflags}}' -o {{local_cache}}/fab-go ./cmd/fab
    rsync -a --delete --exclude='bin/' fab/.kit/ {{local_cache}}/kit/
    @echo "Built fab-go → {{local_cache}}/fab-go"

# Build shim to .release-build/ (install to PATH or test directly)
build-shim:
    mkdir -p .release-build
    cd src/go/shim && CGO_ENABLED=0 go build -ldflags '{{shim_ldflags}}' -o ../../../.release-build/fab ./cmd
    @echo "Built shim → .release-build/fab"

# Check prerequisites and environment health
doctor:
    fab/.kit/scripts/fab-doctor.sh
    @echo ""
    {{scripts}}/dev-doctor.sh

# ── Release ──────────────────────────────────────────────────────────

# Bump version, commit, tag, and push (CI handles the rest)
release bump="patch":
    scripts/release.sh {{bump}}

# Cross-compile a binary for a specific target
_build-binary src_dir cmd_path name os arch ldflags="":
    mkdir -p .release-build && cd {{src_dir}} && CGO_ENABLED=0 GOOS={{os}} GOARCH={{arch}} go build -ldflags '{{ldflags}}' -o ../../../.release-build/{{name}}-{{os}}-{{arch}} {{cmd_path}}

# Cross-compile all binaries for a specific target
build-target os arch:
    just _build-binary src/go/fab ./cmd/fab fab {{os}} {{arch}} '{{fab_ldflags}}'
    just _build-binary src/go/idea ./cmd idea {{os}} {{arch}}
    just _build-binary src/go/wt ./cmd wt {{os}} {{arch}}
    just _build-binary src/go/shim ./cmd shim {{os}} {{arch}} '{{shim_ldflags}}'

# Cross-compile all binaries for all release targets
build-all:
    just build-target darwin arm64
    just build-target darwin amd64
    just build-target linux arm64
    just build-target linux amd64

# Package kit archives for release (generic + per-platform with fab-go)
package-kit:
    {{scripts}}/package-kit.sh

# Package brew archives for Homebrew (per-platform with fab shim, wt, idea)
package-brew:
    {{scripts}}/package-brew.sh

# Remove build artifacts and kit archives
clean:
    rm -rf .release-build
    rm -f kit-*.tar.gz brew-*.tar.gz
