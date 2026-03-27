scripts := "scripts/just"
fab_version := `cat fab/.kit/VERSION`
fab_ldflags := "-X main.version=" + fab_version

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

# Build all binaries for current platform (fab, wt, idea)
build:
    cd src/go/fab && CGO_ENABLED=0 go build -ldflags '{{fab_ldflags}}' -o ../../../fab/.kit/bin/fab-go ./cmd/fab
    cd src/go/idea && CGO_ENABLED=0 go build -ldflags '{{fab_ldflags}}' -o ../../../fab/.kit/bin/idea ./cmd
    cd src/go/wt && CGO_ENABLED=0 go build -ldflags '{{fab_ldflags}}' -o ../../../fab/.kit/bin/wt ./cmd

# Build the fab shim binary for current platform
build-shim:
    cd src/go/shim && CGO_ENABLED=0 go build -ldflags '{{fab_ldflags}}' -o ../../../.release-build/fab-shim ./cmd

# Cross-compile the fab shim for a specific target
build-shim-target os arch:
    just _build-binary src/go/shim ./cmd fab-shim {{os}} {{arch}} '{{fab_ldflags}}'

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
    just _build-binary src/go/idea ./cmd idea {{os}} {{arch}} '{{fab_ldflags}}'
    just _build-binary src/go/wt ./cmd wt {{os}} {{arch}} '{{fab_ldflags}}'

# Cross-compile all binaries for all release targets
build-all:
    just build-target darwin arm64
    just build-target darwin amd64
    just build-target linux arm64
    just build-target linux amd64

# Package kit archives for release (generic + per-platform with binaries)
package-kit:
    {{scripts}}/package-kit.sh

# Remove build artifacts and kit archives
clean:
    rm -rf .release-build
    rm -f kit.tar.gz kit-*.tar.gz
