package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/sahil87/fab-kit/src/go/fab-kit/internal"
)

var version = "dev"

// fabKitArgs is the static allowlist of arguments routed to fab-kit.
var fabKitArgs = map[string]bool{
	"init":         true,
	"upgrade-repo": true,
	"sync":         true,
	"update":       true,
	"doctor":       true,
}

// fabGoNoConfigArgs is the static allowlist of fab-go subcommands that may
// execute without fab/project/config.yaml present. These commands resolve
// state from their arguments (e.g., tmux pane IDs), not from the invoker's CWD.
var fabGoNoConfigArgs = map[string]bool{
	"pane": true,
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	arg := os.Args[1]

	switch {
	case arg == "--version" || arg == "-v":
		cfg, _ := internal.ResolveConfig()
		printVersion(os.Stdout, version, cfg)
	case arg == "--help" || arg == "-h" || arg == "help":
		printHelp()
	case fabKitArgs[arg]:
		execFabKit(os.Args[1:])
	default:
		execFabGo(os.Args[1:])
	}
}

// printVersion writes the system version and, when cfg is non-nil, the project-pinned version.
func printVersion(w io.Writer, sysVersion string, cfg *internal.ConfigResult) {
	fmt.Fprintf(w, "fab %s\n", sysVersion)
	if cfg != nil {
		fmt.Fprintf(w, "project: %s\n", cfg.FabVersion)
	}
}

// execFabKit dispatches to the fab-kit binary via syscall.Exec.
func execFabKit(args []string) {
	bin, err := exec.LookPath("fab-kit")
	if err != nil {
		// Fall back to fab-kit next to the current binary
		self, _ := os.Executable()
		bin = filepath.Join(filepath.Dir(self), "fab-kit")
	}
	argv := append([]string{bin}, args...)
	if err := syscall.Exec(bin, argv, os.Environ()); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot exec fab-kit: %s\n", err)
		os.Exit(1)
	}
}

// resolveFabVersion picks which fab-go version to use based on whether we're
// inside a fab repo and whether the invoked subcommand is exempt from the
// config requirement. Pure: no I/O, no os.Exit, no syscall.Exec.
//
// Rules:
//   - cfg != nil                                    → (cfg.FabVersion, false)
//   - cfg == nil && fabGoNoConfigArgs[arg0] == true → (routerVersion, false)
//   - cfg == nil && not exempt (incl. empty arg0)   → ("", true)
func resolveFabVersion(cfg *internal.ConfigResult, arg0 string, routerVersion string) (fabVersion string, shouldExit bool) {
	if cfg != nil {
		return cfg.FabVersion, false
	}
	if fabGoNoConfigArgs[arg0] {
		return routerVersion, false
	}
	return "", true
}

// execFabGo resolves the fab version from config.yaml (or falls back to the
// router's bundled version for exempt subcommands), ensures the binary is
// cached, and replaces the current process with fab-go.
func execFabGo(args []string) {
	cfg, err := internal.ResolveConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	var arg0 string
	if len(args) > 0 {
		arg0 = args[0]
	}
	fabVersion, shouldExit := resolveFabVersion(cfg, arg0, version)
	if shouldExit {
		fmt.Fprintln(os.Stderr, "Not in a fab-managed repo. Run 'fab init' to set one up.")
		os.Exit(1)
	}

	bin, err := internal.EnsureCached(fabVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	argv := append([]string{bin}, args...)
	if err := syscall.Exec(bin, argv, os.Environ()); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot exec fab-go: %s\n", err)
		os.Exit(1)
	}
}

// printHelp composes help output from both sub-binaries.
func printHelp() {
	fmt.Printf("fab %s — workspace & workflow toolkit\n\n", version)
	fmt.Println("Workspace commands:")
	fmt.Println("  init          Initialize fab in the current repo")
	fmt.Println("  upgrade-repo  Upgrade to a specific or latest version")
	fmt.Println("  sync          Sync workspace (skills, directories, scaffold)")
	fmt.Println("  update        Update fab-kit itself via Homebrew")
	fmt.Println("  doctor        Validate fab-kit prerequisites")
	fmt.Println()

	// Show workflow commands. Inside a fab repo, use the project-pinned version.
	// Outside a fab repo, fall back to the router's bundled version so exempt
	// commands (e.g., pane) remain discoverable from scratch tabs. Errors are
	// silently swallowed — the help section is best-effort.
	cfg, _ := internal.ResolveConfig()
	var fabVersion string
	if cfg != nil {
		fabVersion = cfg.FabVersion
	} else {
		fabVersion = version
	}
	if bin, err := internal.EnsureCached(fabVersion); err == nil {
		if out, err := exec.Command(bin, "--help").Output(); err == nil {
			fmt.Println("Workflow commands (fab-go):")
			fmt.Print(string(out))
			fmt.Println()
		}
	}

	fmt.Println("Flags:")
	fmt.Println("  --version, -v   Show version")
	fmt.Println("  --help, -h      Show this help")
}
