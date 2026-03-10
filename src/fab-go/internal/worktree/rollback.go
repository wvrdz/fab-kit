package worktree

import (
	"os/exec"
	"strings"
)

// Rollback provides a LIFO stack of commands to execute on failure.
// Register commands as they succeed, Execute runs them in reverse order,
// and Disarm clears the stack on success.
type Rollback struct {
	commands []string
	armed    bool
}

// NewRollback creates a new armed rollback stack.
func NewRollback() *Rollback {
	return &Rollback{armed: true}
}

// Register pushes a rollback command onto the stack.
func (r *Rollback) Register(cmd string) {
	r.commands = append(r.commands, cmd)
}

// Execute runs all registered commands in LIFO (reverse) order.
// Individual command failures do not prevent subsequent commands from executing.
func (r *Rollback) Execute() {
	if !r.armed {
		return
	}
	for i := len(r.commands) - 1; i >= 0; i-- {
		parts := strings.Fields(r.commands[i])
		if len(parts) == 0 {
			continue
		}
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Run() // Intentionally ignore errors
	}
}

// Disarm clears the rollback stack, preventing Execute from doing anything.
func (r *Rollback) Disarm() {
	r.armed = false
	r.commands = nil
}

// Commands returns the current list of registered commands (for testing).
func (r *Rollback) Commands() []string {
	return r.commands
}
