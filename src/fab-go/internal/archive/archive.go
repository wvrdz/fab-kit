package archive

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wvrdz/fab-kit/src/fab-go/internal/change"
	"github.com/wvrdz/fab-kit/src/fab-go/internal/resolve"
)

// ArchiveResult holds the YAML output for archive operations.
type ArchiveResult struct {
	Action  string
	Name    string
	Clean   string
	Move    string
	Index   string
	Pointer string
}

// RestoreResult holds the YAML output for restore operations.
type RestoreResult struct {
	Action  string
	Name    string
	Move    string
	Index   string
	Pointer string
}

// Archive moves a change to the archive directory.
func Archive(fabRoot, changeArg, description string) (*ArchiveResult, error) {
	if changeArg == "" {
		return nil, fmt.Errorf("<change> argument is required for archive")
	}
	if description == "" {
		return nil, fmt.Errorf("--description is required for archive")
	}

	folder, err := resolve.ToFolder(fabRoot, changeArg)
	if err != nil {
		return nil, err
	}

	changesDir := filepath.Join(fabRoot, "changes")
	archiveDir := filepath.Join(changesDir, "archive")
	changeDir := filepath.Join(changesDir, folder)

	// 1. Clean: delete .pr-done if present
	cleanStatus := "not_present"
	prDonePath := filepath.Join(changeDir, ".pr-done")
	if _, err := os.Stat(prDonePath); err == nil {
		os.Remove(prDonePath)
		cleanStatus = "removed"
	}

	// 2. Move to archive
	os.MkdirAll(archiveDir, 0755)
	destPath := filepath.Join(archiveDir, folder)
	if _, err := os.Stat(destPath); err == nil {
		return nil, fmt.Errorf("Archive destination already exists: %s", destPath)
	}
	if err := os.Rename(changeDir, destPath); err != nil {
		return nil, fmt.Errorf("move to archive: %w", err)
	}

	// 3. Update index
	indexFile := filepath.Join(archiveDir, "index.md")
	indexStatus := updateIndex(indexFile, folder, description)

	// Backfill unindexed
	backfillIndex(archiveDir, indexFile)

	// 4. Clear pointer if active
	pointerStatus := "skipped"
	activeFolder, err := resolve.ToFolder(fabRoot, "")
	if err == nil && activeFolder == folder {
		change.SwitchBlank(fabRoot)
		pointerStatus = "cleared"
	}

	return &ArchiveResult{
		Action:  "archive",
		Name:    folder,
		Clean:   cleanStatus,
		Move:    "moved",
		Index:   indexStatus,
		Pointer: pointerStatus,
	}, nil
}

// Restore moves a change from the archive back to active.
func Restore(fabRoot, changeArg string, doSwitch bool) (*RestoreResult, error) {
	if changeArg == "" {
		return nil, fmt.Errorf("<change> argument is required for restore")
	}

	folder, err := resolveArchive(fabRoot, changeArg)
	if err != nil {
		return nil, err
	}

	changesDir := filepath.Join(fabRoot, "changes")
	archiveDir := filepath.Join(changesDir, "archive")

	// 1. Move from archive
	moveStatus := "restored"
	destPath := filepath.Join(changesDir, folder)
	if _, err := os.Stat(destPath); err == nil {
		moveStatus = "already_in_changes"
	} else {
		srcPath := filepath.Join(archiveDir, folder)
		if err := os.Rename(srcPath, destPath); err != nil {
			return nil, fmt.Errorf("restore: %w", err)
		}
	}

	// 2. Remove from index
	indexFile := filepath.Join(archiveDir, "index.md")
	indexStatus := removeFromIndex(indexFile, folder)

	// 3. Optionally switch
	pointerStatus := "skipped"
	if doSwitch {
		_, err := change.Switch(fabRoot, folder)
		if err == nil {
			pointerStatus = "switched"
		}
	}

	return &RestoreResult{
		Action:  "restore",
		Name:    folder,
		Move:    moveStatus,
		Index:   indexStatus,
		Pointer: pointerStatus,
	}, nil
}

// List returns archived change folder names.
func List(fabRoot string) ([]string, error) {
	archiveDir := filepath.Join(fabRoot, "changes", "archive")
	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, e := range entries {
		if e.IsDir() {
			results = append(results, e.Name())
		}
	}
	return results, nil
}

// FormatArchiveYAML formats an ArchiveResult.
func FormatArchiveYAML(r *ArchiveResult) string {
	return fmt.Sprintf("action: %s\nname: %s\nclean: %s\nmove: %s\nindex: %s\npointer: %s",
		r.Action, r.Name, r.Clean, r.Move, r.Index, r.Pointer)
}

// FormatRestoreYAML formats a RestoreResult.
func FormatRestoreYAML(r *RestoreResult) string {
	return fmt.Sprintf("action: %s\nname: %s\nmove: %s\nindex: %s\npointer: %s",
		r.Action, r.Name, r.Move, r.Index, r.Pointer)
}

func resolveArchive(fabRoot, override string) (string, error) {
	if override == "" {
		return "", fmt.Errorf("<change> argument is required for restore")
	}

	archiveDir := filepath.Join(fabRoot, "changes", "archive")
	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		return "", fmt.Errorf("No archive folder found.")
	}

	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		return "", err
	}

	var folders []string
	for _, e := range entries {
		if e.IsDir() {
			folders = append(folders, e.Name())
		}
	}

	if len(folders) == 0 {
		return "", fmt.Errorf("No archived changes found.")
	}

	overrideLower := strings.ToLower(override)

	// Exact match
	for _, f := range folders {
		if strings.ToLower(f) == overrideLower {
			return f, nil
		}
	}

	// Substring match
	var partials []string
	for _, f := range folders {
		if strings.Contains(strings.ToLower(f), overrideLower) {
			partials = append(partials, f)
		}
	}

	if len(partials) == 1 {
		return partials[0], nil
	}
	if len(partials) > 1 {
		return "", fmt.Errorf("Multiple archives match \"%s\": %s.", override, strings.Join(partials, ", "))
	}

	return "", fmt.Errorf("No archive matches \"%s\".", override)
}

func updateIndex(indexFile, folder, description string) string {
	indexStatus := "updated"
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		os.WriteFile(indexFile, []byte("# Archive Index\n\n"), 0644)
		indexStatus = "created"
	}

	// Normalize description
	description = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return ' '
		}
		return r
	}, description)
	description = strings.TrimSpace(description)

	newEntry := fmt.Sprintf("- **%s** — %s", folder, description)

	data, _ := os.ReadFile(indexFile)
	lines := strings.Split(string(data), "\n")

	var result []string
	if len(lines) >= 2 {
		result = append(result, lines[0], lines[1])
	} else if len(lines) == 1 {
		result = append(result, lines[0], "")
	} else {
		result = append(result, "# Archive Index", "")
	}
	result = append(result, newEntry)
	if len(lines) > 2 {
		result = append(result, lines[2:]...)
	}

	os.WriteFile(indexFile, []byte(strings.Join(result, "\n")), 0644)
	return indexStatus
}

func backfillIndex(archiveDir, indexFile string) {
	entries, _ := os.ReadDir(archiveDir)
	indexData, _ := os.ReadFile(indexFile)
	indexContent := string(indexData)

	f, err := os.OpenFile(indexFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		marker := fmt.Sprintf("**%s**", e.Name())
		if !strings.Contains(indexContent, marker) {
			fmt.Fprintf(f, "- **%s** — (no description — pre-index archive)\n", e.Name())
		}
	}
}

func removeFromIndex(indexFile, folder string) string {
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		return "not_found"
	}

	marker := fmt.Sprintf("**%s**", folder)

	f, err := os.Open(indexFile)
	if err != nil {
		return "not_found"
	}

	var found bool
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, marker) {
			found = true
			continue
		}
		lines = append(lines, line)
	}
	f.Close()

	if !found {
		return "not_found"
	}

	os.WriteFile(indexFile, []byte(strings.Join(lines, "\n")), 0644)
	return "removed"
}
