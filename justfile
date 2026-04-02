scripts := "scripts/just"
fab_version := `cat fab/.kit/VERSION`
fab_ldflags := "-X main.version=" + fab_version
fab_kit_ldflags := "-X main.version=" + fab_version

# -- Development ------------------------------------------------------------------

# Run all tests
test:
    cd src/go/fab && go test ./... -count=1
    cd src/go/idea && go test ./... -count=1
    cd src/go/wt && go test ./... -count=1
    cd src/go/fab-kit && go test ./... -count=1

# Run all tests (verbose)
test-v:
    cd src/go/fab && go test ./... -v -count=1
    cd src/go/idea && go test ./... -v -count=1
    cd src/go/wt && go test ./... -v -count=1
    cd src/go/fab-kit && go test ./... -v -count=1

local_cache := env("HOME", "") / ".fab-kit/local-versions" / fab_version

# Build fab-go + copy .kit content to local cache (shim auto-discovers)
build:
    mkdir -p {{local_cache}}/kit
    cd src/go/fab && CGO_ENABLED=0 go build -ldflags '{{fab_ldflags}}' -o {{local_cache}}/fab-go ./cmd/fab
    rsync -a --delete --exclude='bin/' fab/.kit/ {{local_cache}}/kit/
    @echo "Built fab-go → {{local_cache}}/fab-go"

# Build fab-kit and fab router to .release-build/ (install to PATH or test directly)
build-fab-kit:
    mkdir -p .release-build
    cd src/go/fab-kit && CGO_ENABLED=0 go build -ldflags '{{fab_kit_ldflags}}' -o ../../../.release-build/fab-kit ./cmd/fab-kit
    cd src/go/fab-kit && CGO_ENABLED=0 go build -ldflags '{{fab_kit_ldflags}}' -o ../../../.release-build/fab ./cmd/fab
    @echo "Built fab-kit → .release-build/fab-kit"
    @echo "Built fab (router) → .release-build/fab"

# Check prerequisites and environment health
doctor:
    fab/.kit/scripts/fab-doctor.sh
    @echo ""
    {{scripts}}/dev-doctor.sh

# -- Release -----------------------------------------------------------------------

# Bump version, commit, tag, and push (CI handles the rest)
release bump="patch":
    scripts/release.sh {{bump}}

# Cross-compile a binary for a specific target
_build-binary src_dir cmd_path name os arch ldflags="":
    mkdir -p .release-build && cd {{src_dir}} && CGO_ENABLED=0 GOOS={{os}} GOARCH={{arch}} go build -ldflags '{{ldflags}}' -o ../../../.release-build/{{name}}-{{os}}-{{arch}} {{cmd_path}}

# Cross-compile all binaries for a specific target (5 binaries)
build-target os arch:
    just _build-binary src/go/fab ./cmd/fab fab-go {{os}} {{arch}} '{{fab_ldflags}}'
    just _build-binary src/go/idea ./cmd idea {{os}} {{arch}}
    just _build-binary src/go/wt ./cmd wt {{os}} {{arch}}
    just _build-binary src/go/fab-kit ./cmd/fab-kit fab-kit {{os}} {{arch}} '{{fab_kit_ldflags}}'
    just _build-binary src/go/fab-kit ./cmd/fab fab {{os}} {{arch}} '{{fab_kit_ldflags}}'

# Cross-compile all binaries for all release targets (5 binaries x 4 platforms = 20)
build-all:
    just build-target darwin arm64
    just build-target darwin amd64
    just build-target linux arm64
    just build-target linux amd64

# Package kit archives for release (generic + per-platform with fab-go)
package-kit:
    {{scripts}}/package-kit.sh

# Package brew archives for Homebrew (per-platform with fab, fab-kit, wt, idea)
package-brew:
    {{scripts}}/package-brew.sh

# Remove build artifacts and kit archives
clean:
    rm -rf .release-build
    rm -f kit-*.tar.gz brew-*.tar.gz
