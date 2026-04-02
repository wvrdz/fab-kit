#!/usr/bin/env bash
set -euo pipefail

# Package brew archives for Homebrew (fab router + fab-kit + wt + idea per platform)
# Called by: just package-brew
#
# Produces brew-{os}-{arch}.tar.gz for each platform.
# Each archive contains four binaries: fab, fab-kit, wt, idea.

platforms=("darwin/arm64" "darwin/amd64" "linux/arm64" "linux/amd64")
build_dir=".release-build"

echo "Packaging brew archives..."

for platform in "${platforms[@]}"; do
  os="${platform%%/*}"
  arch="${platform##*/}"

  fab="$build_dir/fab-${os}-${arch}"
  fab_kit="$build_dir/fab-kit-${os}-${arch}"
  wt="$build_dir/wt-${os}-${arch}"
  idea="$build_dir/idea-${os}-${arch}"

  for bin in "$fab" "$fab_kit" "$wt" "$idea"; do
    if [ ! -f "$bin" ]; then
      echo "ERROR: Missing $bin — run 'just build-all' first."
      exit 1
    fi
  done

  archive_name="brew-${os}-${arch}.tar.gz"
  staging="$build_dir/brew-staging-${os}-${arch}"
  rm -rf "$staging"
  mkdir -p "$staging"

  cp "$fab" "$staging/fab"
  cp "$fab_kit" "$staging/fab-kit"
  cp "$wt" "$staging/wt"
  cp "$idea" "$staging/idea"
  chmod +x "$staging/fab" "$staging/fab-kit" "$staging/wt" "$staging/idea"

  COPYFILE_DISABLE=1 tar czf "$archive_name" -C "$staging" fab fab-kit wt idea
  echo "  $archive_name ($(wc -c < "$archive_name") bytes)"
  rm -rf "$staging"
done

echo "Brew packaging complete: ${#platforms[@]} archives"
