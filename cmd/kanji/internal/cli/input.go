package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// ResolveTextInput resolves rich text from an inline flag, a file flag, or stdin (--file -).
// It enforces mutual exclusivity between inline and file sources.
// If neither flag is provided and allowEmpty is false, it returns an error.
func ResolveTextInput(cmd *cobra.Command, inlineFlag, fileFlag string, allowEmpty bool, stdin io.Reader) (string, error) {
	inlineChanged := cmd.Flags().Changed(inlineFlag)
	fileChanged := cmd.Flags().Changed(fileFlag)

	if inlineChanged && fileChanged {
		return "", NewValidation(fmt.Sprintf("--%s and --%s are mutually exclusive", inlineFlag, fileFlag))
	}

	if !inlineChanged && !fileChanged {
		if allowEmpty {
			return "", nil
		}
		return "", NewValidation(fmt.Sprintf("--%s or --%s is required", inlineFlag, fileFlag))
	}

	if inlineChanged {
		val, err := cmd.Flags().GetString(inlineFlag)
		if err != nil {
			return "", NewValidation(fmt.Sprintf("read --%s: %v", inlineFlag, err))
		}
		return val, nil
	}

	filePath, err := cmd.Flags().GetString(fileFlag)
	if err != nil {
		return "", NewValidation(fmt.Sprintf("read --%s: %v", fileFlag, err))
	}

	if filePath == "-" {
		r := stdin
		if r == nil {
			r = os.Stdin
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return string(data), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read file %q: %w", filePath, err)
	}
	return string(data), nil
}
