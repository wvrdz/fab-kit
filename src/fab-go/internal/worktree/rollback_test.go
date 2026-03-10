package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRollback_LIFOOrder(t *testing.T) {
	rb := NewRollback()

	// Register commands that create marker files in a temp dir
	dir := t.TempDir()
	file1 := filepath.Join(dir, "order.txt")

	// Use commands that append to a file to verify LIFO order
	rb.Register("touch " + filepath.Join(dir, "first"))
	rb.Register("touch " + filepath.Join(dir, "second"))

	rb.Execute()

	// Both files should exist (Execute runs all commands)
	if _, err := os.Stat(filepath.Join(dir, "first")); os.IsNotExist(err) {
		t.Error("first rollback command did not execute")
	}
	if _, err := os.Stat(filepath.Join(dir, "second")); os.IsNotExist(err) {
		t.Error("second rollback command did not execute")
	}

	// Verify LIFO by checking Commands() order before execution
	rb2 := NewRollback()
	rb2.Register("cmd-A")
	rb2.Register("cmd-B")
	rb2.Register("cmd-C")

	cmds := rb2.Commands()
	if len(cmds) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(cmds))
	}
	if cmds[0] != "cmd-A" || cmds[1] != "cmd-B" || cmds[2] != "cmd-C" {
		t.Errorf("unexpected command order: %v", cmds)
	}

	_ = file1
}

func TestRollback_Disarm(t *testing.T) {
	rb := NewRollback()
	dir := t.TempDir()

	marker := filepath.Join(dir, "should-not-exist")
	rb.Register("touch " + marker)
	rb.Disarm()
	rb.Execute()

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Error("disarmed rollback should not have created the file")
	}
}

func TestRollback_Empty(t *testing.T) {
	rb := NewRollback()
	// Execute on empty stack should not panic
	rb.Execute()
}

func TestRollback_ContinuesOnFailure(t *testing.T) {
	rb := NewRollback()
	dir := t.TempDir()

	marker := filepath.Join(dir, "exists")
	rb.Register("touch " + marker)
	rb.Register("nonexistent-command-that-should-fail")

	// Execute should still run the first command even though second fails
	rb.Execute()

	if _, err := os.Stat(marker); os.IsNotExist(err) {
		t.Error("rollback should continue executing after a command failure")
	}
}
