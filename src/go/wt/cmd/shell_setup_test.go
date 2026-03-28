package main

import (
	"testing"
)

func TestShellSetup_Bash(t *testing.T) {
	repo := createTestRepo(t)
	r := runWt(t, repo, []string{"SHELL=/bin/bash"}, "shell-setup")
	assertExitCode(t, r, 0)
	assertContains(t, r.Stdout, "wt()")
	assertContains(t, r.Stdout, "WT_WRAPPER=1")
	assertContains(t, r.Stdout, "eval")
	assertContains(t, r.Stdout, "command wt")
}

func TestShellSetup_Zsh(t *testing.T) {
	repo := createTestRepo(t)
	r := runWt(t, repo, []string{"SHELL=/usr/bin/zsh"}, "shell-setup")
	assertExitCode(t, r, 0)
	assertContains(t, r.Stdout, "wt()")
	assertContains(t, r.Stdout, "WT_WRAPPER=1")
	assertContains(t, r.Stdout, "eval")
}

func TestShellSetup_UnsupportedShell(t *testing.T) {
	repo := createTestRepo(t)
	r := runWt(t, repo, []string{"SHELL=/usr/bin/fish"}, "shell-setup")
	if r.ExitCode == 0 {
		t.Error("expected non-zero exit code for unsupported shell")
	}
	assertContains(t, r.Stderr, `unsupported shell "fish"`)
	assertContains(t, r.Stderr, "supported: bash, zsh")
}

func TestShellSetup_EmptyShell(t *testing.T) {
	repo := createTestRepo(t)
	r := runWt(t, repo, []string{"SHELL="}, "shell-setup")
	if r.ExitCode == 0 {
		t.Error("expected non-zero exit code for empty SHELL")
	}
	assertContains(t, r.Stderr, "unsupported shell")
}

func TestShellSetup_WrapperContainsCdDetection(t *testing.T) {
	repo := createTestRepo(t)
	r := runWt(t, repo, []string{"SHELL=/bin/bash"}, "shell-setup")
	assertExitCode(t, r, 0)
	assertContains(t, r.Stdout, `cd\ *`)
}
