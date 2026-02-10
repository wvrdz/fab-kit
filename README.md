# Fab Kit

Fab Kit is a Specification-Driven Development (SDD) workflow kit that runs entirely as AI agent prompts — no CLI installation, no system dependencies. It provides named stages, markdown templates, and skill definitions that any AI agent (Claude Code, Cursor, Windsurf, etc.) can execute.

The core engine lives in `fab/.kit/` as markdown skill files, templates, and shell scripts. You copy it into your project and go.

## Quick Start

### Bootstrap a new project

```bash
mkdir -p fab
curl -sL https://github.com/wvrdz/fab-kit/releases/latest/download/kit.tar.gz | tar xz -C fab/
```

Then run setup and init:

```bash
fab/.kit/scripts/fab-setup.sh   # creates directories, symlinks, .gitignore
```

Once setup completes, use your AI agent to run:

```
/fab-init     # generates config.yaml and constitution.md
/fab-new      # starts your first change
```

### Alternative: manual copy

If you have a local clone of this repo:

```bash
cp -r fab/.kit /path/to/your-project/fab/.kit
```

Then run `fab-setup.sh` and `/fab-init` as above.

## Updating

To update `fab/.kit/` to the latest release:

```bash
bash fab/.kit/scripts/fab-update.sh
```

This will:
1. Download the latest `kit.tar.gz` from GitHub Releases
2. Atomically replace `fab/.kit/` (your `config.yaml`, `docs/`, `changes/`, etc. are never touched)
3. Display the version change (e.g., "0.1.0 → 0.2.0")
4. Re-run `fab-setup.sh` to repair symlinks

**Requires**: [gh CLI](https://cli.github.com/) installed and authenticated.

### Check your version

```bash
cat fab/.kit/VERSION
```

## Creating a Release

For maintainers of this repo — to publish a new release:

```bash
bash fab/.kit/scripts/fab-release.sh [patch|minor|major]
```

- `patch` (default): 0.1.0 → 0.1.1
- `minor`: 0.1.0 → 0.2.0
- `major`: 0.1.0 → 1.0.0

The script will:
1. Bump the version in `fab/.kit/VERSION`
2. Package `fab/.kit/` into `kit.tar.gz`
3. Commit the VERSION bump
4. Create a GitHub Release with `kit.tar.gz` as an asset

**Requires**: clean working tree, [gh CLI](https://cli.github.com/), and a configured `origin` remote.

## What's in the Box

```
fab/.kit/
├── VERSION          # Semver version string
├── skills/          # Markdown skill definitions for AI agents
├── templates/       # Artifact templates (proposal, spec, plan, tasks, checklist)
└── scripts/         # Shell utilities (setup, status, update, release)
```

The kit provides a 7-stage workflow: **proposal → specs → plan → tasks → apply → review → archive**. See [fab/specs/index.md](fab/specs/index.md) for the full specification.

## Documentation

- **Specs** (design intent): [fab/specs/index.md](fab/specs/index.md)
- **Docs** (what shipped): [fab/docs/](fab/docs/)

## References

The `references/` folder contains docs from other projects, included for reference:

- [references/speckit/](references/speckit/) — Analysis of GitHub's Spec-Kit
- [references/openspec/](references/openspec/) — Analysis of Fission AI's OpenSpec
