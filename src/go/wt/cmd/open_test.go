package main

import (
	"strings"
	"testing"
)

func TestOpen_ErrorNonexistentWorktree(t *testing.T) {
	repo := createTestRepo(t)

	r := runWt(t, repo, nil, "open", "--app", "code", "nonexistent-wt")
	if r.ExitCode == 0 {
		t.Error("expected failure for nonexistent worktree")
	}
	assertContains(t, r.Stderr, "not found")
}

func TestOpen_ErrorFromMainRepoWithoutTarget(t *testing.T) {
	repo := createTestRepo(t)

	r := runWt(t, repo, nil, "open", "--app", "code")
	if r.ExitCode == 0 {
		t.Error("expected failure from main repo without target")
	}
	assertContains(t, r.Stderr, "No worktree specified")
}

func TestOpen_ErrorOutsideGitRepo(t *testing.T) {
	dir := t.TempDir()
	r := runWt(t, dir, nil, "open")
	if r.ExitCode == 0 {
		t.Error("expected failure outside git repo")
	}
	assertContains(t, r.Stderr, "Not a git repository")
}

func TestOpen_ErrorUnknownApp(t *testing.T) {
	repo := createTestRepo(t)
	wtPath := createWorktreeViaWt(t, repo, "app-err")

	r := runWt(t, repo, nil, "open", "--app", "nonexistent-app", wtPath)
	if r.ExitCode == 0 {
		t.Error("expected failure for unknown app")
	}
	assertContains(t, r.Stderr, "Unknown app")
}

func TestOpen_AppDefault_NoDefaultDetected(t *testing.T) {
	repo := createTestRepo(t)
	wtPath := createWorktreeViaWt(t, repo, "default-err")

	// Clear environment so no default can be detected, and only open_here is available
	env := []string{
		"TERM_PROGRAM=",
		"TMUX=",
		"BYOBU_BACKEND=",
		"BYOBU_TTY=",
		"BYOBU_SESSION=",
		"BYOBU_CONFIG_DIR=",
		"HOME=" + t.TempDir(),
	}

	r := runWt(t, repo, env, "open", "--app", "default", wtPath)
	// DetectDefaultApp falls through to the first non-open_here app.
	// Can't control installed apps, so accept either:
	// - exit 0 (some app resolved and opened)
	// - non-zero (no default detected — expected error message)
	if r.ExitCode != 0 {
		assertContains(t, r.Stderr, "No default app detected")
	}
	// Must never panic
	if strings.Contains(r.Stderr, "panic") {
		t.Errorf("command panicked: %s", r.Stderr)
	}
}

// NOTE: Testing actual app opening (code, cursor, etc.) requires mock binaries
// on PATH that log their invocations. We test the error paths here; the
// open-by-name success path is tested via the worktree resolution logic
// (which is shared with other commands).
