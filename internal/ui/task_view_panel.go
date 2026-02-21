package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"

	"github.com/tiagokriok/kanji/internal/domain"
)

type taskViewerLayout struct {
	panelWidth    int
	panelHeight   int
	contentWidth  int
	contentHeight int
	leftWidth     int
	rightWidth    int
}

func (m Model) renderTaskViewerPanel(base string) string {
	_ = base

	task, ok := m.viewerTask()
	if !ok {
		return base
	}

	layout := m.taskViewerLayout()
	leftLines := m.renderTaskViewerLeftLines(task, layout.leftWidth, layout.contentHeight, m.viewDescScroll)
	rightLines := m.renderTaskViewerRightLines(layout.rightWidth, layout.contentHeight)
	leftBlock := lipgloss.NewStyle().
		Width(layout.leftWidth).
		Height(layout.contentHeight).
		MaxWidth(layout.leftWidth).
		MaxHeight(layout.contentHeight).
		Render(strings.Join(leftLines, "\n"))
	rightBlock := lipgloss.NewStyle().
		Width(layout.rightWidth).
		Height(layout.contentHeight).
		MaxWidth(layout.rightWidth).
		MaxHeight(layout.contentHeight).
		Render(strings.Join(rightLines, "\n"))
	sepLine := strings.Repeat("│\n", max(0, layout.contentHeight-1)) + "│"
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Height(layout.contentHeight).
		Render(sepLine)
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, separator, rightBlock)

	panel := lipgloss.NewStyle().
		Width(layout.contentWidth).
		Height(layout.contentHeight).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("250")).
		Render(body)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}

func (m Model) taskViewerLayout() taskViewerLayout {
	panelWidth := m.width - 6
	if panelWidth < 80 {
		panelWidth = 80
	}
	if panelWidth > m.width-2 {
		panelWidth = max(20, m.width-2)
	}

	panelHeight := m.height - 4
	if panelHeight < 18 {
		panelHeight = 18
	}
	if panelHeight > m.height-2 {
		panelHeight = max(10, m.height-2)
	}

	contentWidth := boxContentWidth(panelWidth, 1, true)
	contentHeight := boxContentHeight(panelHeight, true)

	rightWidth := max(24, contentWidth/4)
	if rightWidth > contentWidth-24 {
		rightWidth = contentWidth - 24
	}
	if rightWidth < 16 {
		rightWidth = 16
	}
	leftWidth := contentWidth - rightWidth - 1
	if leftWidth < 20 {
		leftWidth = 20
		rightWidth = max(16, contentWidth-leftWidth-1)
	}

	return taskViewerLayout{
		panelWidth:    panelWidth,
		panelHeight:   panelHeight,
		contentWidth:  contentWidth,
		contentHeight: contentHeight,
		leftWidth:     leftWidth,
		rightWidth:    rightWidth,
	}
}

func (m Model) taskViewerDescViewportLines() int {
	layout := m.taskViewerLayout()
	return max(1, layout.contentHeight-3) // title + meta + hint
}

func (m Model) taskViewerMaxDescScroll() int {
	task, ok := m.viewerTask()
	if !ok {
		return 0
	}
	layout := m.taskViewerLayout()
	descLines := renderViewerMarkdownLines(task.DescriptionMD, layout.leftWidth)
	viewport := max(1, layout.contentHeight-3)
	return max(0, len(descLines)-viewport)
}

func (m Model) renderTaskViewerLeftLines(task domain.Task, width, height, scroll int) []string {
	titleStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("231")).Bold(true)
	metaStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("246"))
	descStyle := lipgloss.NewStyle().Width(width)
	hintStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("244"))

	dueText := "-"
	dueColor := dueColorDefault
	if task.DueAt != nil {
		dueText, dueColor = m.dueDisplay(*task.DueAt)
	}
	dueValue := lipgloss.NewStyle().Foreground(dueColor).Bold(true).Render(dueText)
	priorityValue := lipgloss.NewStyle().
		Foreground(priorityColor(normalizePriority(task.Priority))).
		Bold(true).
		Render(fmt.Sprintf("p%d", task.Priority))
	statusValue := lipgloss.NewStyle().
		Foreground(m.statusColorForTask(task)).
		Bold(true).
		Render(m.statusLabelForTask(task))

	lines := []string{
		titleStyle.Render(truncate(task.Title, max(1, width))),
		metaStyle.Render(fmt.Sprintf("%s | %s | %s", dueValue, priorityValue, statusValue)),
		hintStyle.Render("j/k or ↑/↓ scroll description | Enter/Esc close"),
	}

	descLines := renderViewerMarkdownLines(task.DescriptionMD, width)
	viewport := max(1, height-len(lines))
	maxScroll := max(0, len(descLines)-viewport)
	if scroll < 0 {
		scroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}
	end := min(len(descLines), scroll+viewport)
	for i := scroll; i < end; i++ {
		lines = append(lines, descStyle.Width(width).Render(descLines[i]))
	}
	return lines
}

func (m Model) renderTaskViewerRightLines(width, height int) []string {
	headerStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("231")).Bold(true).Align(lipgloss.Center)
	commentMetaStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("245"))
	commentBodyStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("252"))
	emptyStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("245"))

	lines := []string{
		headerStyle.Render("Comments"),
		emptyStyle.Render(""),
	}

	if m.comments == nil {
		lines = append(lines, emptyStyle.Render("(loading...)"))
		return trimViewerLines(lines, width, height)
	}
	if len(m.comments) == 0 {
		lines = append(lines, emptyStyle.Render("(none)"))
		return trimViewerLines(lines, width, height)
	}

	for _, comment := range m.comments {
		author := ""
		if comment.Author != nil {
			author = strings.TrimSpace(*comment.Author)
		}
		if author == "" {
			author = "unknown"
		}
		meta := fmt.Sprintf("%s  %s", author, m.formatCommentDateTime(comment.CreatedAt))
		lines = append(lines, commentMetaStyle.Render(truncate(meta, max(1, width))))

		body := strings.TrimSpace(comment.BodyMD)
		if body == "" {
			body = "(empty)"
		}
		body = normalizeViewerText(body)
		for _, bodyLine := range wrapViewerText(body, width-2) {
			lines = append(lines, commentBodyStyle.Render("  "+bodyLine))
			if len(lines) >= height {
				return lines[:height]
			}
		}
		lines = append(lines, emptyStyle.Render(""))
		if len(lines) >= height {
			return lines[:height]
		}
	}

	return trimViewerLines(lines, width, height)
}

func trimViewerLines(lines []string, width, height int) []string {
	if len(lines) > height {
		return lines[:height]
	}
	return append(lines, makeViewerBlankLines(width, height-len(lines))...)
}

func makeViewerBlankLines(width, count int) []string {
	style := lipgloss.NewStyle().Width(width)
	lines := make([]string, 0, count)
	for i := 0; i < count; i++ {
		lines = append(lines, style.Render(""))
	}
	return lines
}

func wrapViewerText(text string, width int) []string {
	text = normalizeViewerText(text)
	if width < 1 {
		return []string{text}
	}

	rawLines := strings.Split(text, "\n")
	wrapped := make([]string, 0, len(rawLines))
	for _, raw := range rawLines {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			wrapped = append(wrapped, "")
			continue
		}

		remaining := raw
		for len([]rune(remaining)) > width {
			runes := []rune(remaining)
			chunk := runes[:width]
			breakAt := -1
			for i := len(chunk) - 1; i >= 0; i-- {
				if chunk[i] == ' ' || chunk[i] == '\t' {
					breakAt = i
					break
				}
			}

			if breakAt <= 0 {
				wrapped = append(wrapped, string(chunk))
				remaining = strings.TrimLeft(string(runes[width:]), " \t")
				continue
			}

			wrapped = append(wrapped, strings.TrimSpace(string(chunk[:breakAt])))
			remaining = strings.TrimLeft(string(runes[breakAt:]), " \t")
		}
		if remaining != "" {
			wrapped = append(wrapped, remaining)
		}
	}
	return wrapped
}

func renderViewerMarkdownLines(md string, width int) []string {
	md = strings.TrimSpace(normalizeViewerText(md))
	if md == "" {
		return []string{"(empty)"}
	}

	rendered, err := renderMarkdownWithGlamour(md, width)
	if err != nil {
		return wrapViewerText(md, width)
	}
	rendered = strings.TrimRight(normalizeViewerText(rendered), "\n")
	if rendered == "" {
		return []string{"(empty)"}
	}
	return strings.Split(rendered, "\n")
}

func renderMarkdownWithGlamour(md string, width int) (string, error) {
	if width < 20 {
		width = 20
	}
	style := styles.DarkStyleConfig
	style.H1.Prefix = " "
	style.H2.Prefix = "  "
	style.H3.Prefix = "   "
	style.H4.Prefix = "    "
	style.H5.Prefix = "     "
	style.H6.Prefix = "      "
	style.H1.BackgroundColor = nil
	style.H1.Color = stringPtr("51")
	style.H1.Bold = boolPtr(true)
	style.H2.Bold = boolPtr(true)
	style.H2.Color = stringPtr("45")
	style.H2.Underline = boolPtr(false)
	style.H3.Bold = boolPtr(true)
	style.H3.Color = stringPtr("44")
	style.H4.Bold = boolPtr(false)
	style.H4.Color = stringPtr("43")
	style.H5.Bold = boolPtr(false)
	style.H5.Color = stringPtr("79")
	style.H6.Bold = boolPtr(false)
	style.H6.Color = stringPtr("116")

	renderer, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(width),
		glamour.WithStyles(style),
	)
	if err != nil {
		return "", err
	}
	return renderer.Render(md)
}

func boolPtr(v bool) *bool {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func normalizeViewerText(text string) string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return normalized
}
