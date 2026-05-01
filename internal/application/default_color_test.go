package application

import (
	"testing"

	"github.com/tiagokriok/kanji/internal/domain"
)

func TestNextDefaultColor_Empty(t *testing.T) {
	got := NextDefaultColor(nil)
	if got != "#60A5FA" {
		t.Fatalf("expected first palette color, got %q", got)
	}
}

func TestNextDefaultColor_First(t *testing.T) {
	cols := []domain.Column{{Name: "Todo", Color: "#60A5FA"}}
	got := NextDefaultColor(cols)
	if got != "#F59E0B" {
		t.Fatalf("expected second palette color, got %q", got)
	}
}

func TestNextDefaultColor_Second(t *testing.T) {
	cols := []domain.Column{
		{Name: "Todo", Color: "#60A5FA"},
		{Name: "Doing", Color: "#F59E0B"},
	}
	got := NextDefaultColor(cols)
	if got != "#22C55E" {
		t.Fatalf("expected third palette color, got %q", got)
	}
}

func TestNextDefaultColor_Wraps(t *testing.T) {
	cols := []domain.Column{
		{Name: "Todo", Color: "#60A5FA"},
		{Name: "Doing", Color: "#F59E0B"},
		{Name: "Done", Color: "#22C55E"},
	}
	got := NextDefaultColor(cols)
	if got != "#60A5FA" {
		t.Fatalf("expected wrap to first palette color, got %q", got)
	}
}

func TestNextDefaultColor_ManyColumns(t *testing.T) {
	cols := make([]domain.Column, 10)
	for i := range cols {
		cols[i] = domain.Column{Name: "Col", Color: "#FFFFFF"}
	}
	got := NextDefaultColor(cols)
	if got != "#F59E0B" {
		t.Fatalf("expected second palette color for 10 columns (10 %% 3 = 1), got %q", got)
	}
}
