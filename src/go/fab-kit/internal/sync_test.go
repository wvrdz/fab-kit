package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestScaffoldTreeWalk_CopyIfAbsent(t *testing.T) {
	scaffoldDir := t.TempDir()
	repoRoot := t.TempDir()

	// Create scaffold file
	os.MkdirAll(filepath.Join(scaffoldDir, "docs", "memory"), 0755)
	os.WriteFile(filepath.Join(scaffoldDir, "docs", "memory", "index.md"), []byte("# Index\n"), 0644)

	// Run tree-walk
	if err := scaffoldTreeWalk(scaffoldDir, repoRoot); err != nil {
		t.Fatalf("scaffoldTreeWalk failed: %v", err)
	}

	// Verify file was copied
	data, err := os.ReadFile(filepath.Join(repoRoot, "docs", "memory", "index.md"))
	if err != nil {
		t.Fatal("expected index.md to be created")
	}
	if string(data) != "# Index\n" {
		t.Errorf("unexpected content: %s", string(data))
	}
}

func TestScaffoldTreeWalk_CopyIfAbsentSkip(t *testing.T) {
	scaffoldDir := t.TempDir()
	repoRoot := t.TempDir()

	// Create scaffold file
	os.WriteFile(filepath.Join(scaffoldDir, "existing.md"), []byte("scaffold content\n"), 0644)

	// Create destination file with different content
	os.WriteFile(filepath.Join(repoRoot, "existing.md"), []byte("user content\n"), 0644)

	// Run tree-walk
	if err := scaffoldTreeWalk(scaffoldDir, repoRoot); err != nil {
		t.Fatalf("scaffoldTreeWalk failed: %v", err)
	}

	// Verify existing file was NOT overwritten
	data, _ := os.ReadFile(filepath.Join(repoRoot, "existing.md"))
	if string(data) != "user content\n" {
		t.Errorf("existing file should not be overwritten, got: %s", string(data))
	}
}

func TestJsonMergePermissions_CreateNew(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	src := filepath.Join(srcDir, "settings.json")
	dest := filepath.Join(destDir, "settings.json")

	srcJSON := map[string]interface{}{
		"permissions": map[string]interface{}{
			"allow": []interface{}{"Bash(git *)", "Read"},
		},
	}
	srcData, _ := json.MarshalIndent(srcJSON, "", "  ")
	os.WriteFile(src, srcData, 0644)

	if err := jsonMergePermissions(src, dest, "settings.json"); err != nil {
		t.Fatalf("jsonMergePermissions failed: %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal("expected dest file to be created")
	}

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	allow := extractPermissionsAllow(result)
	if len(allow) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(allow))
	}
}

func TestJsonMergePermissions_Merge(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	src := filepath.Join(srcDir, "settings.json")
	dest := filepath.Join(destDir, "settings.json")

	srcJSON := map[string]interface{}{
		"permissions": map[string]interface{}{
			"allow": []interface{}{"Bash(git *)", "Read", "Write"},
		},
	}
	destJSON := map[string]interface{}{
		"permissions": map[string]interface{}{
			"allow": []interface{}{"Bash(git *)", "Edit"},
		},
	}
	srcData, _ := json.MarshalIndent(srcJSON, "", "  ")
	destData, _ := json.MarshalIndent(destJSON, "", "  ")
	os.WriteFile(src, srcData, 0644)
	os.WriteFile(dest, destData, 0644)

	if err := jsonMergePermissions(src, dest, "settings.json"); err != nil {
		t.Fatalf("jsonMergePermissions failed: %v", err)
	}

	// Read merged result
	data, _ := os.ReadFile(dest)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	allow := extractPermissionsAllow(result)

	// Should have 4: Edit (existing), Bash(git *) (existing/deduped), Read (new), Write (new)
	if len(allow) != 4 {
		t.Errorf("expected 4 permissions after merge, got %d: %v", len(allow), allow)
	}
}

func TestJsonMergePermissions_NoDuplicates(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	src := filepath.Join(srcDir, "settings.json")
	dest := filepath.Join(destDir, "settings.json")

	// Same permissions in both — no change expected
	perms := map[string]interface{}{
		"permissions": map[string]interface{}{
			"allow": []interface{}{"Bash(git *)", "Read"},
		},
	}
	srcData, _ := json.MarshalIndent(perms, "", "  ")
	os.WriteFile(src, srcData, 0644)
	os.WriteFile(dest, srcData, 0644)

	if err := jsonMergePermissions(src, dest, "settings.json"); err != nil {
		t.Fatalf("jsonMergePermissions failed: %v", err)
	}

	data, _ := os.ReadFile(dest)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	allow := extractPermissionsAllow(result)
	if len(allow) != 2 {
		t.Errorf("expected 2 permissions (no duplicates), got %d", len(allow))
	}
}

func TestLineEnsureMerge_CreateNew(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	src := filepath.Join(srcDir, "gitignore")
	dest := filepath.Join(destDir, ".gitignore")

	os.WriteFile(src, []byte("# comment\nnode_modules/\n.env\n"), 0644)

	if err := lineEnsureMerge(src, dest, ".gitignore"); err != nil {
		t.Fatalf("lineEnsureMerge failed: %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal("expected dest file to be created")
	}

	content := string(data)
	if content == "" {
		t.Fatal("file should not be empty")
	}
}

func TestLineEnsureMerge_AppendNew(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	src := filepath.Join(srcDir, "gitignore")
	dest := filepath.Join(destDir, ".gitignore")

	os.WriteFile(src, []byte("node_modules/\n.env\n"), 0644)
	os.WriteFile(dest, []byte("node_modules/\n"), 0644)

	if err := lineEnsureMerge(src, dest, ".gitignore"); err != nil {
		t.Fatalf("lineEnsureMerge failed: %v", err)
	}

	data, _ := os.ReadFile(dest)
	content := string(data)
	// Should contain .env but not duplicate node_modules/
	if content == "" {
		t.Fatal("file should not be empty")
	}
}

func TestLineEnsureMerge_SkipComments(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	src := filepath.Join(srcDir, "entries")
	dest := filepath.Join(destDir, "entries")

	os.WriteFile(src, []byte("# this is a comment\nactual-entry\n"), 0644)

	if err := lineEnsureMerge(src, dest, "entries"); err != nil {
		t.Fatalf("lineEnsureMerge failed: %v", err)
	}

	data, _ := os.ReadFile(dest)
	content := string(data)
	// Should only contain "actual-entry", not the comment
	if content == "" {
		t.Fatal("file should not be empty")
	}
}

func TestListSkills(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "fab-new.md"), []byte("# New\n"), 0644)
	os.WriteFile(filepath.Join(dir, "_preamble.md"), []byte("# Preamble\n"), 0644)
	os.WriteFile(filepath.Join(dir, "fab-setup.md"), []byte("# Setup\n"), 0644)
	os.WriteFile(filepath.Join(dir, "README.txt"), []byte("Not a skill\n"), 0644)

	skills := listSkills(dir)
	if len(skills) != 3 {
		t.Errorf("expected 3 skills (.md files), got %d: %v", len(skills), skills)
	}
}

func TestAgentAvailable_FABAgentsOverride(t *testing.T) {
	t.Setenv("FAB_AGENTS", "claude codex")

	if !agentAvailable("claude") {
		t.Error("expected claude to be available via FAB_AGENTS")
	}
	if !agentAvailable("codex") {
		t.Error("expected codex to be available via FAB_AGENTS")
	}
	if agentAvailable("opencode") {
		t.Error("expected opencode to NOT be available when FAB_AGENTS is set without it")
	}
}

func TestCleanStaleSkills_Directory(t *testing.T) {
	baseDir := t.TempDir()
	repoRoot := filepath.Dir(baseDir)

	// Create directory-format skill entries
	os.MkdirAll(filepath.Join(baseDir, "fab-new"), 0755)
	os.WriteFile(filepath.Join(baseDir, "fab-new", "SKILL.md"), []byte("# New\n"), 0644)
	os.MkdirAll(filepath.Join(baseDir, "old-skill"), 0755)
	os.WriteFile(filepath.Join(baseDir, "old-skill", "SKILL.md"), []byte("# Old\n"), 0644)

	// Canonical skills: only fab-new
	skills := []string{"fab-new"}
	cleanStaleSkills(baseDir, "directory", skills, repoRoot)

	// old-skill should be removed
	if _, err := os.Stat(filepath.Join(baseDir, "old-skill")); !os.IsNotExist(err) {
		t.Error("expected old-skill directory to be removed")
	}
	// fab-new should still exist
	if _, err := os.Stat(filepath.Join(baseDir, "fab-new", "SKILL.md")); err != nil {
		t.Error("expected fab-new skill to still exist")
	}
}

func TestCleanStaleSkills_Flat(t *testing.T) {
	baseDir := t.TempDir()
	repoRoot := filepath.Dir(baseDir)

	// Create flat-format skill entries
	os.WriteFile(filepath.Join(baseDir, "fab-new.md"), []byte("# New\n"), 0644)
	os.WriteFile(filepath.Join(baseDir, "old-skill.md"), []byte("# Old\n"), 0644)

	skills := []string{"fab-new"}
	cleanStaleSkills(baseDir, "flat", skills, repoRoot)

	// old-skill.md should be removed
	if _, err := os.Stat(filepath.Join(baseDir, "old-skill.md")); !os.IsNotExist(err) {
		t.Error("expected old-skill.md to be removed")
	}
	// fab-new.md should still exist
	if _, err := os.Stat(filepath.Join(baseDir, "fab-new.md")); err != nil {
		t.Error("expected fab-new.md to still exist")
	}
}
