package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// RequireAtLeastOneFlag fails if none of the given flags were changed on the command.
func RequireAtLeastOneFlag(cmd *cobra.Command, flagNames ...string) error {
	for _, name := range flagNames {
		if cmd.Flags().Changed(name) {
			return nil
		}
	}
	quoted := make([]string, len(flagNames))
	for i, name := range flagNames {
		quoted[i] = "--" + name
	}
	return NewValidation(fmt.Sprintf("at least one of %s is required", strings.Join(quoted, ", ")))
}

// MutuallyExclusiveFlags fails if flags from different sets are both changed.
func MutuallyExclusiveFlags(cmd *cobra.Command, flagSets ...[]string) error {
	changedSets := 0
	for _, set := range flagSets {
		for _, name := range set {
			if cmd.Flags().Changed(name) {
				changedSets++
				break
			}
		}
	}
	if changedSets > 1 {
		return NewValidation("flags are mutually exclusive")
	}
	return nil
}

// NormalizeLabels trims, lowercases, deduplicates and filters out empty labels.
func NormalizeLabels(labels []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(labels))
	for _, label := range labels {
		normalized := strings.ToLower(strings.TrimSpace(label))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

// RequireConfirmation fails if the confirmation bool flag is not true.
func RequireConfirmation(cmd *cobra.Command, flagName string) error {
	confirmed, err := cmd.Flags().GetBool(flagName)
	if err != nil {
		return NewValidation(fmt.Sprintf("read --%s flag: %v", flagName, err))
	}
	if !confirmed {
		return NewValidation(fmt.Sprintf("operation requires --%s confirmation", flagName))
	}
	return nil
}
