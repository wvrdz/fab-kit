package worktree

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestBuildAvailableApps_OpenHereFirst(t *testing.T) {
	apps := BuildAvailableApps()
	if len(apps) == 0 {
		t.Fatal("BuildAvailableApps returned no apps")
	}
	first := apps[0]
	if first.Name != "Open here" || first.Cmd != "open_here" {
		t.Errorf("expected first app to be {\"Open here\", \"open_here\"}, got {%q, %q}", first.Name, first.Cmd)
	}
}

func TestDetectDefaultApp_SkipsOpenHere(t *testing.T) {
	apps := BuildAvailableApps()
	if len(apps) < 2 {
		t.Skip("need at least 2 apps to test fallback")
	}

	// Clear environment to force fallback path
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TMUX", "")
	t.Setenv("BYOBU_BACKEND", "")
	// Remove last-app cache to ensure fallback
	t.Setenv("HOME", t.TempDir())

	idx := DetectDefaultApp(apps)
	if idx > 0 && idx <= len(apps) {
		if apps[idx-1].Cmd == "open_here" {
			t.Errorf("DetectDefaultApp returned 'open_here' as default (index %d)", idx)
		}
	}
}

func TestOpenInApp_OpenHere(t *testing.T) {
	path := "/tmp/test-worktree"
	t.Setenv("WT_WRAPPER", "1") // Suppress stderr hint

	// Capture stdout
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	t.Cleanup(func() {
		os.Stdout = origStdout
		_ = r.Close()
		_ = w.Close()
	})
	os.Stdout = w

	openErr := OpenInApp("open_here", path, "repo", "wt-name")

	w.Close()

	if openErr != nil {
		t.Fatalf("OpenInApp returned error: %v", openErr)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy: %v", err)
	}

	expected := fmt.Sprintf("cd -- %q\n", path)
	if buf.String() != expected {
		t.Errorf("expected stdout %q, got %q", expected, buf.String())
	}
}

func TestOpenInApp_OpenHere_HintWhenNoWrapper(t *testing.T) {
	path := "/tmp/test-worktree"
	t.Setenv("WT_WRAPPER", "")

	// Capture stdout
	origStdout := os.Stdout
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe stdout: %v", err)
	}

	// Capture stderr
	origStderr := os.Stderr
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe stderr: %v", err)
	}

	t.Cleanup(func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
		_ = rOut.Close()
		_ = wOut.Close()
		_ = rErr.Close()
		_ = wErr.Close()
	})

	os.Stdout = wOut
	os.Stderr = wErr

	openErr := OpenInApp("open_here", path, "repo", "wt-name")

	wOut.Close()
	wErr.Close()

	if openErr != nil {
		t.Fatalf("OpenInApp returned error: %v", openErr)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	if _, err := io.Copy(&stdoutBuf, rOut); err != nil {
		t.Fatalf("io.Copy stdout: %v", err)
	}
	if _, err := io.Copy(&stderrBuf, rErr); err != nil {
		t.Fatalf("io.Copy stderr: %v", err)
	}

	// stdout should have the cd command
	expectedCd := fmt.Sprintf("cd -- %q\n", path)
	if stdoutBuf.String() != expectedCd {
		t.Errorf("expected stdout %q, got %q", expectedCd, stdoutBuf.String())
	}

	// stderr should have the hint
	if stderrStr := stderrBuf.String(); stderrStr == "" {
		t.Error("expected stderr hint when WT_WRAPPER is not set, got empty")
	} else {
		if !bytes.Contains(stderrBuf.Bytes(), []byte(`wt shell-setup`)) {
			t.Errorf("expected stderr to mention wt shell-setup, got %q", stderrStr)
		}
	}
}

func TestOpenInApp_OpenHere_NoHintWithWrapper(t *testing.T) {
	path := "/tmp/test-worktree"
	t.Setenv("WT_WRAPPER", "1")

	// Capture stdout
	origStdout := os.Stdout
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe stdout: %v", err)
	}

	// Capture stderr
	origStderr := os.Stderr
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe stderr: %v", err)
	}

	t.Cleanup(func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
		_ = rOut.Close()
		_ = wOut.Close()
		_ = rErr.Close()
		_ = wErr.Close()
	})

	os.Stdout = wOut
	os.Stderr = wErr

	openErr := OpenInApp("open_here", path, "repo", "wt-name")

	wOut.Close()
	wErr.Close()

	if openErr != nil {
		t.Fatalf("OpenInApp returned error: %v", openErr)
	}

	var stderrBuf bytes.Buffer
	if _, err := io.Copy(&stderrBuf, rErr); err != nil {
		t.Fatalf("io.Copy stderr: %v", err)
	}

	if stderrBuf.Len() != 0 {
		t.Errorf("expected no stderr when WT_WRAPPER=1, got %q", stderrBuf.String())
	}
}
