package cmd

import (
	"bytes"
	"testing"

	pgecloudctl "github.com/AntTheLimey/pgecloudctl"
)

// TestLLMSCmdPrintsEmbeddedReference verifies that running `llms` writes
// the embedded llms-full.txt verbatim to stdout, not just that the
// reference content mentions every command (which docs_sync_test covers).
func TestLLMSCmdPrintsEmbeddedReference(t *testing.T) {
	var out bytes.Buffer
	llmsCmd.SetOut(&out)
	t.Cleanup(func() { llmsCmd.SetOut(nil) })

	llmsCmd.Run(llmsCmd, nil)

	if got := out.String(); got != string(pgecloudctl.LLMSFull) {
		t.Errorf("llms output (%d bytes) does not match embedded "+
			"llms-full.txt (%d bytes)", len(got), len(pgecloudctl.LLMSFull))
	}
}
