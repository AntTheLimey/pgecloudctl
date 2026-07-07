package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	pgecloudctl "github.com/AntTheLimey/pgecloudctl"
	"github.com/spf13/cobra"
)

var skillInstallDir string

func init() {
	rootCmd.AddCommand(skillCmd)
	skillCmd.AddCommand(skillInstallCmd)

	skillInstallCmd.Flags().StringVar(&skillInstallDir, "dir", "",
		"Target directory (default: ~/.claude/skills/pgecloudctl)")
}

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage the bundled Claude Code skill",
}

var skillInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the Claude Code skill to ~/.claude/skills",
	Long: `install copies the bundled pgEdge Cloud skill (SKILL.md and
knowledge-base.md) into ~/.claude/skills/pgecloudctl, where Claude Code
discovers skills automatically. Safe to re-run; upgrades in place.`,
	RunE: runSkillInstall,
}

// skillVersionFile records which CLI version installed the skill, so
// `doctor` can report it and upgrades can be detected.
const skillVersionFile = ".version"

func runSkillInstall(cmd *cobra.Command, _ []string) error {
	dir := skillInstallDir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolve home directory: %w", err)
		}
		dir = filepath.Join(home, ".claude", "skills", "pgecloudctl")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create skill directory: %w", err)
	}

	files := map[string][]byte{
		"SKILL.md":          pgecloudctl.SkillMD,
		"knowledge-base.md": pgecloudctl.SkillKnowledgeBase,
		skillVersionFile:    []byte(Version + "\n"),
	}
	for name, data := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(),
		"Skill installed to %s (version %s).\n"+
			"Claude Code discovers it automatically on the next session.\n",
		dir, Version)
	return nil
}
