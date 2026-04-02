#!/usr/bin/env bash
set -euo pipefail

# Package brew archives for Homebrew (fab shim + wt + idea per platform)
# Called by: just package-brew
#
# Produces brew-{os}-{arch}.tar.gz for each platform.
# Each archive contains three binaries: fab, wt, idea.

platforms=("darwin/arm64" "darwin/amd64" "linux/arm64" "linux/amd64")
build_dir=".release-build"

echo "Packaging brew archives..."

for platform in "${platforms[@]}"; do
  os="${platform%%/*}"
  arch="${platform##*/}"

  shim="$build_dir/shim-${os}-${arch}"
  wt="$build_dir/wt-${os}-${arch}"
  idea="$build_dir/idea-${os}-${arch}"

  for bin in "$shim" "$wt" "$idea"; do
    if [ ! -f "$bin" ]; then
      echo "ERROR: Missing $bin — run 'just build-all' first."
      exit 1
    fi
  done

  archive_name="brew-${os}-${arch}.tar.gz"
  staging="$build_dir/brew-staging-${os}-${arch}"
  rm -rf "$staging"
  mkdir -p "$staging"

  cp "$shim" "$staging/fab"
  cp "$wt" "$staging/wt"
  cp "$idea" "$staging/idea"
  chmod +x "$staging/fab" "$staging/wt" "$staging/idea"

  COPYFILE_DISABLE=1 tar czf "$archive_name" -C "$staging" fab wt idea
  echo "  $archive_name ($(wc -c < "$archive_name") bytes)"
  rm -rf "$staging"
done

echo "Brew packaging complete: ${#platforms[@]} archives"
