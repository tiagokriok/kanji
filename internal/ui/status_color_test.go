package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestContrastingTextColorUsesLightTextOnDarkBackground(t *testing.T) {
	got := contrastingTextColorFromHexOrDefault("#1d4ed8", "252")
	want := lipgloss.Color("255")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestContrastingTextColorUsesDarkTextOnLightBackground(t *testing.T) {
	got := contrastingTextColorFromHexOrDefault("#f59e0b", "252")
	want := lipgloss.Color("232")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
