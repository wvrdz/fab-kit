scripts := "src/scripts/just"
rust_src := "src/fab-rust"
go_src := "src/fab-go"

# Run all tests with summary (excludes Rust)
test:
    just test-setup
    just test-hooks
    just test-packages
    just test-scripts
    just test-go

# Setup test dependencies
test-setup:
    {{scripts}}/test-setup.sh

# Run hook tests (bats)
test-hooks:
    {{scripts}}/test-hooks.sh

# Run package tests (bats)
test-packages:
    {{scripts}}/test-packages.sh

# Run script tests (bats)
test-scripts:
    {{scripts}}/test-scripts.sh

# Run Go unit tests (all binaries)
test-go:
    cd {{go_src}} && go test ./... -count=1

# Run Go unit tests (verbose)
test-go-v:
    cd {{go_src}} && go test ./... -v -count=1

# Build all Go binaries for current platform (fab, wt, idea)
build-go:
    cd {{go_src}} && CGO_ENABLED=0 go build -o ../../fab/.kit/bin/fab-go ./cmd/fab
    cd {{go_src}} && CGO_ENABLED=0 go build -o ../../fab/.kit/bin/wt ./cmd/wt
    cd {{go_src}} && CGO_ENABLED=0 go build -o ../../fab/.kit/bin/idea ./cmd/idea

# Cross-compile a Go binary for a specific target
_build-go-binary name cmd os arch:
    mkdir -p .release-build && cd {{go_src}} && CGO_ENABLED=0 GOOS={{os}} GOARCH={{arch}} go build -o ../../.release-build/{{name}}-{{os}}-{{arch}} ./cmd/{{cmd}}

# Cross-compile all Go binaries for a specific target
build-go-target os arch:
    just _build-go-binary fab fab {{os}} {{arch}}
    just _build-go-binary wt wt {{os}} {{arch}}
    just _build-go-binary idea idea {{os}} {{arch}}

# Cross-compile all Go binaries for all release targets
build-go-all:
    just build-go-target darwin arm64
    just build-go-target darwin amd64
    just build-go-target linux arm64
    just build-go-target linux amd64

# Build everything for release (Go binaries only — use build-rust-all for Rust)
build-all:
    mkdir -p .release-build
    just build-go-all

# Run Rust integration tests
test-rust:
    cargo test --manifest-path src/fab-rust/Cargo.toml

# Build Rust binary for the current platform (local dev)
build-rust:
    cargo build --manifest-path {{rust_src}}/Cargo.toml --release
    cp {{rust_src}}/target/release/fab fab/.kit/bin/fab-rust

# Map os/arch pair to Rust target triple
_rust-target os arch:
    #!/usr/bin/env sh
    case "{{os}}-{{arch}}" in
      darwin-arm64) echo "aarch64-apple-darwin" ;;
      darwin-amd64) echo "x86_64-apple-darwin" ;;
      linux-arm64)  echo "aarch64-unknown-linux-musl" ;;
      linux-amd64)  echo "x86_64-unknown-linux-musl" ;;
      *) echo "ERROR: unknown platform {{os}}-{{arch}}" >&2; exit 1 ;;
    esac

# Cross-compile Rust binary for a specific target triple
build-rust-target target:
    cargo zigbuild --manifest-path {{rust_src}}/Cargo.toml --release --target {{target}}
    mkdir -p .release-build
    cp {{rust_src}}/target/{{target}}/release/fab .release-build/fab-rust-{{target}}

# Cross-compile Rust binary for all release targets
build-rust-all:
    just build-rust-target aarch64-apple-darwin
    just build-rust-target x86_64-apple-darwin
    just build-rust-target aarch64-unknown-linux-musl
    just build-rust-target x86_64-unknown-linux-musl

# Package kit archives for release (generic + per-platform with Go binaries)
package-kit:
    #!/usr/bin/env bash
    set -euo pipefail
    platforms=("darwin/arm64" "darwin/amd64" "linux/arm64" "linux/amd64")
    build_dir=".release-build"
    # Verify all cross-compiled Go binaries exist
    for platform in "${platforms[@]}"; do
      os="${platform%%/*}"
      arch="${platform##*/}"
      for bin in fab wt idea; do
        binary="$build_dir/${bin}-${os}-${arch}"
        if [ ! -f "$binary" ]; then
          echo "ERROR: Missing $bin binary $binary — run 'just build-go-all' first."
          exit 1
        fi
      done
    done
    # Generic archive (no binaries)
    echo "Packaging kit.tar.gz (generic, no binary)..."
    COPYFILE_DISABLE=1 tar czf kit.tar.gz -C fab --exclude='.kit/bin/fab-go' --exclude='.kit/bin/fab-rust' --exclude='.kit/bin/wt' --exclude='.kit/bin/idea' .kit
    echo "  kit.tar.gz ($(wc -c < kit.tar.gz) bytes)"
    # Per-platform archives (kit + Go binaries)
    for platform in "${platforms[@]}"; do
      os="${platform%%/*}"
      arch="${platform##*/}"
      archive_name="kit-${os}-${arch}.tar.gz"
      staging="$build_dir/staging-${os}-${arch}"
      rm -rf "$staging"
      mkdir -p "$staging/.kit/bin"
      cp -a fab/.kit "$staging/.kit"
      for bin_pair in "fab:fab-go" "wt:wt" "idea:idea"; do
        src_name="${bin_pair%%:*}"
        dest_name="${bin_pair##*:}"
        cp "$build_dir/${src_name}-${os}-${arch}" "$staging/.kit/bin/${dest_name}"
        chmod +x "$staging/.kit/bin/${dest_name}"
      done
      COPYFILE_DISABLE=1 tar czf "$archive_name" -C "$staging" .kit
      echo "  $archive_name ($(wc -c < "$archive_name") bytes)"
      rm -rf "$staging"
    done
    echo "Packaging complete: kit.tar.gz + ${#platforms[@]} platform archives"

# Remove build artifacts and kit archives
clean:
    rm -rf .release-build
    rm -f kit.tar.gz kit-*.tar.gz

# Check prerequisites and environment health
doctor:
    fab/.kit/scripts/fab-doctor.sh
    @echo ""
    {{scripts}}/dev-doctor.sh
