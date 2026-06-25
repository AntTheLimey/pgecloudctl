package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// stdinIsTerminal reports whether stdin is an interactive terminal. It uses a
// real terminal check (ioctl) rather than the file mode, so pipes, regular
// files, and character devices such as /dev/null are all treated as
// non-interactive. It is a package var so tests can override it.
var stdinIsTerminal = func() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// confirmDestructive gates a destructive operation behind user confirmation.
//
//   - skip (the --yes flag) bypasses the prompt and proceeds.
//   - In a non-interactive context (stdin is not a terminal) without --yes, it
//     returns an ExitError so scripts and AI agents fail loudly rather than
//     hang on a prompt or silently no-op.
//   - On a terminal it prompts; a "y"/"yes" answer proceeds, anything else
//     prints "Aborted." and returns (false, nil).
//
// It returns (true, nil) when the caller should proceed.
func confirmDestructive(
	cmd *cobra.Command, skip bool, prompt string,
) (bool, error) {
	if skip {
		return true, nil
	}

	if !stdinIsTerminal() {
		return false, &ExitError{
			msg: "refusing to proceed without --yes in a " +
				"non-interactive context",
			code: ExitGeneral,
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s [y/N]: ", prompt)
	reader := bufio.NewReader(cmd.InOrStdin())
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "y" || answer == "yes" {
		return true, nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
	return false, nil
}
