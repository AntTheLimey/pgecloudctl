package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pgecloudctl "github.com/AntTheLimey/pgecloudctl"
)

func TestSkillInstall(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "skills", "pgecloudctl")

	prevDir := skillInstallDir
	prevVersion := Version
	t.Cleanup(func() {
		skillInstallDir = prevDir
		Version = prevVersion
	})
	skillInstallDir = dir
	Version = "v0.5.0-test"

	if err := runSkillInstall(skillInstallCmd, nil); err != nil {
		t.Fatalf("skill install: %v", err)
	}

	files := map[string][]byte{
		"SKILL.md":          pgecloudctl.SkillMD,
		"knowledge-base.md": pgecloudctl.SkillKnowledgeBase,
		".version":          []byte("v0.5.0-test\n"),
	}
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read installed %s: %v", name, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("installed %s does not match embedded content", name)
		}
	}

	// Idempotent: re-running upgrades in place without error.
	Version = "v0.5.1-test"
	if err := runSkillInstall(skillInstallCmd, nil); err != nil {
		t.Fatalf("skill re-install: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, ".version"))
	if err != nil {
		t.Fatalf("read .version after re-install: %v", err)
	}
	if strings.TrimSpace(string(got)) != "v0.5.1-test" {
		t.Errorf(".version = %q, want v0.5.1-test", got)
	}
}

func TestSkillFrontmatterHasTriggers(t *testing.T) {
	// The skill only fires if its description mentions the terms agents
	// actually use. Guard the trigger vocabulary against regressions.
	head := string(pgecloudctl.SkillMD)
	if i := strings.Index(head[3:], "---"); i >= 0 {
		head = head[:i+3]
	}
	for _, term := range []string{
		"pgEdge Cloud", "pgecloudctl", "clusters", "databases",
		"MCP", "RAG", "ingresses",
	} {
		if !strings.Contains(head, term) {
			t.Errorf("SKILL.md frontmatter missing trigger term %q", term)
		}
	}
}

func TestLLMSCommandPrintsEmbeddedReference(t *testing.T) {
	out, err := executeCommand("llms")
	if err != nil {
		t.Fatalf("llms command: %v", err)
	}
	if out != string(pgecloudctl.LLMSFull) {
		t.Error("llms output does not match embedded llms-full.txt")
	}
	if !strings.Contains(out, "clusters create") {
		t.Error("llms output missing clusters create reference")
	}
}

func TestRootHelpMentionsLLMS(t *testing.T) {
	out, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("help: %v", err)
	}
	for _, want := range []string{"pgecloudctl llms", "skill install"} {
		if !strings.Contains(out, want) {
			t.Errorf("root help missing %q — AI agents land here first", want)
		}
	}
}
