// Package pgecloudctl embeds the AI-facing documentation shipped in
// this repository so the binary can serve it anywhere it is installed,
// without a checkout of the repo. llms-full.txt is the authoritative
// command reference (printed by `pgecloudctl llms`); the skill files
// are installed into ~/.claude/skills by `pgecloudctl skill install`.
package pgecloudctl

import _ "embed"

// LLMSFull is the complete AI-agent reference (llms-full.txt).
//
//go:embed llms-full.txt
var LLMSFull []byte

// SkillMD is the Claude Code skill definition.
//
//go:embed skills/pgecloudctl/SKILL.md
var SkillMD []byte

// SkillKnowledgeBase is the command reference consulted by the skill.
//
//go:embed skills/pgecloudctl/knowledge-base.md
var SkillKnowledgeBase []byte
