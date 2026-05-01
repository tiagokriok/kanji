package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireAtLeastOneFlag_PassesWhenOneChanged(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("id", "", "")
	require.NoError(t, cmd.Flags().Set("name", "test"))

	err := RequireAtLeastOneFlag(cmd, "name", "id")
	assert.NoError(t, err)
}

func TestRequireAtLeastOneFlag_PassesWhenAllChanged(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("id", "", "")
	require.NoError(t, cmd.Flags().Set("name", "test"))
	require.NoError(t, cmd.Flags().Set("id", "123"))

	err := RequireAtLeastOneFlag(cmd, "name", "id")
	assert.NoError(t, err)
}

func TestRequireAtLeastOneFlag_FailsWhenNoneChanged(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("id", "", "")

	err := RequireAtLeastOneFlag(cmd, "name", "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one")
	assert.Contains(t, err.Error(), "--name")
	assert.Contains(t, err.Error(), "--id")
}

func TestMutuallyExclusiveFlags_PassesWhenSameSet(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("id", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("other", "", "")
	require.NoError(t, cmd.Flags().Set("id", "1"))
	require.NoError(t, cmd.Flags().Set("name", "test"))

	err := MutuallyExclusiveFlags(cmd, []string{"id", "name"}, []string{"other"})
	assert.NoError(t, err)
}

func TestMutuallyExclusiveFlags_PassesWhenOnlyOneSetChanged(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("id", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("other", "", "")
	require.NoError(t, cmd.Flags().Set("id", "1"))

	err := MutuallyExclusiveFlags(cmd, []string{"id", "name"}, []string{"other"})
	assert.NoError(t, err)
}

func TestMutuallyExclusiveFlags_PassesWhenNoneChanged(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("id", "", "")
	cmd.Flags().String("name", "", "")

	err := MutuallyExclusiveFlags(cmd, []string{"id", "name"}, []string{"other"})
	assert.NoError(t, err)
}

func TestMutuallyExclusiveFlags_FailsWhenDifferentSets(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("id", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("other", "", "")
	require.NoError(t, cmd.Flags().Set("id", "1"))
	require.NoError(t, cmd.Flags().Set("other", "x"))

	err := MutuallyExclusiveFlags(cmd, []string{"id", "name"}, []string{"other"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestMutuallyExclusiveFlags_FailsAcrossMultipleSets(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("a", "", "")
	cmd.Flags().String("b", "", "")
	cmd.Flags().String("c", "", "")
	require.NoError(t, cmd.Flags().Set("a", "1"))
	require.NoError(t, cmd.Flags().Set("c", "3"))

	err := MutuallyExclusiveFlags(cmd, []string{"a"}, []string{"b"}, []string{"c"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestNormalizeLabels(t *testing.T) {
	got := NormalizeLabels([]string{"  Bug  ", "BUG", "feature", "  Feature  ", "bug"})
	want := []string{"bug", "feature"}
	assert.Equal(t, want, got)
}

func TestNormalizeLabels_EmptyInput(t *testing.T) {
	got := NormalizeLabels([]string{})
	assert.Empty(t, got)
}

func TestNormalizeLabels_EmptyStrings(t *testing.T) {
	got := NormalizeLabels([]string{"", "  ", "bug", ""})
	assert.Equal(t, []string{"bug"}, got)
}

func TestNormalizeLabels_PreservesOrder(t *testing.T) {
	got := NormalizeLabels([]string{"z", "a", "Z", "A"})
	want := []string{"z", "a"}
	assert.Equal(t, want, got)
}

func TestRequireConfirmation_PassesWhenTrue(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("yes", false, "")
	require.NoError(t, cmd.Flags().Set("yes", "true"))

	err := RequireConfirmation(cmd, "yes")
	assert.NoError(t, err)
}

func TestRequireConfirmation_FailsWhenFalse(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("yes", false, "")
	require.NoError(t, cmd.Flags().Set("yes", "false"))

	err := RequireConfirmation(cmd, "yes")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "confirm")
}

func TestRequireConfirmation_FailsWhenNotSet(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("yes", false, "")

	err := RequireConfirmation(cmd, "yes")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "confirm")
}
