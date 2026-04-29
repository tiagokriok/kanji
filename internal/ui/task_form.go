package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tiagokriok/kanji/internal/domain"
)

type taskFormMode int

const (
	taskFormCreate taskFormMode = iota
	taskFormEdit
)

const (
	taskFieldTitle = iota
	taskFieldDescription
	taskFieldDueDate
	taskFieldPriority
	taskFieldStatus
	taskFieldCount
)

type taskForm struct {
	mode taskFormMode

	taskID string
	focus  int

	title       textinput.Model
	description textinput.Model
	dueDate     textinput.Model

	descriptionFull string
	priorityIndex   int
	statusIndex     int
	statusOptions   []taskStatusOption
}

type taskPriorityOption struct {
	Value int
	Label string
}

type taskStatusOption struct {
	ColumnID string
	Status   string
	Label    string
	ColorHex string
}

func (m Model) handleTaskFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if m.taskForm == nil {
		return m, nil, false
	}

	switch msg.String() {
	case "tab", "down":
		m.taskForm.moveFocus(1)
		return m, textinput.Blink, true
	case "shift+tab", "up":
		m.taskForm.moveFocus(-1)
		return m, textinput.Blink, true
	case "left":
		if m.taskForm.focus == taskFieldPriority && m.taskForm.cyclePriority(-1) {
			return m, nil, true
		}
		if m.taskForm.focus == taskFieldStatus && m.taskForm.cycleStatus(-1) {
			return m, nil, true
		}
	case "right":
		if m.taskForm.focus == taskFieldPriority && m.taskForm.cyclePriority(1) {
			return m, nil, true
		}
		if m.taskForm.focus == taskFieldStatus && m.taskForm.cycleStatus(1) {
			return m, nil, true
		}
	case "ctrl+s":
		model, cmd := m.submitTaskForm()
		return model, cmd, true
	case "ctrl+g":
		return m, openDescriptionEditorCmd(m.taskForm.descriptionFull), true
	case "enter":
		if m.taskForm.focus == taskFieldCount-1 {
			model, cmd := m.submitTaskForm()
			return model, cmd, true
		}
		m.taskForm.moveFocus(1)
		return m, textinput.Blink, true
	case "0", "1", "2", "3", "4", "5":
		if m.taskForm.focus == taskFieldPriority {
			index, err := strconv.Atoi(msg.String())
			if err == nil && index >= 0 && index < len(taskPriorityOptions) {
				m.taskForm.priorityIndex = index
				m.taskForm.clampPriorityIndex()
			}
			return m, nil, true
		}
	}

	return m, nil, false
}

func (m Model) submitTaskForm() (tea.Model, tea.Cmd) {
	cmd, err := m.submitTaskFormCmd()
	if err != nil {
		m.statusLine = err.Error()
		return m, nil
	}
	m.closeTaskForm()
	return m, cmd
}

func (m *Model) closeTaskForm() {
	m.overlayState.closeTaskForm()
	m.statusLine = ""
}

func newTaskFormInput(placeholder, value string, limit int) textinput.Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = placeholder
	ti.CharLimit = limit
	ti.SetValue(value)
	return ti
}

func (m *Model) startCreateTaskForm() {
	statusOptions, statusIndex := m.buildTaskStatusOptions(nil)

	if m.viewMode == viewKanban && m.activeColumn < len(m.columns) {
		activeColID := m.columns[m.activeColumn].ID
		for i, opt := range statusOptions {
			if opt.ColumnID == activeColID {
				statusIndex = i
				break
			}
		}
	}

	form := &taskForm{
		mode:            taskFormCreate,
		title:           newTaskFormInput("Title", "", 512),
		description:     newTaskFormInput("Description", "", 2048),
		dueDate:         newTaskFormInput(m.dueDatePlaceholder(), "", 32),
		descriptionFull: "",
		priorityIndex:   0,
		statusOptions:   statusOptions,
		statusIndex:     statusIndex,
	}
	form.clampPriorityIndex()
	form.clampStatusIndex()
	form.setFocus(taskFieldTitle)

	m.overlayState.startTaskForm(form)
	m.statusLine = "Create task"
}

func (m *Model) startEditTaskForm(task domain.Task) {
	due := ""
	if task.DueAt != nil {
		due = m.formatDueDate(*task.DueAt)
	}

	statusOptions, statusIndex := m.buildTaskStatusOptions(&task)
	priorityIndex := normalizePriority(task.Priority)
	if priorityIndex > 5 {
		priorityIndex = 5
	}

	form := &taskForm{
		mode:            taskFormEdit,
		taskID:          task.ID,
		title:           newTaskFormInput("Title", task.Title, 512),
		description:     newTaskFormInput("Description", summarizeDescription(task.DescriptionMD), 2048),
		dueDate:         newTaskFormInput(m.dueDatePlaceholder(), due, 32),
		descriptionFull: task.DescriptionMD,
		priorityIndex:   priorityIndex,
		statusOptions:   statusOptions,
		statusIndex:     statusIndex,
	}
	form.clampPriorityIndex()
	form.clampStatusIndex()
	form.setFocus(taskFieldTitle)

	m.overlayState.startTaskForm(form)
	m.statusLine = "Edit task"
}

func (m Model) buildTaskStatusOptions(task *domain.Task) ([]taskStatusOption, int) {
	options := make([]taskStatusOption, 0, len(m.columns)+1)
	selected := 0

	if task != nil {
		options = append(options, taskStatusOption{Label: "(none)"})
	}

	for _, col := range m.columns {
		option := taskStatusOption{
			ColumnID: col.ID,
			Status:   strings.ToLower(strings.TrimSpace(col.Name)),
			Label:    col.Name,
			ColorHex: col.Color,
		}
		options = append(options, option)
		if task != nil && task.ColumnID != nil && *task.ColumnID == col.ID {
			selected = len(options) - 1
		}
	}

	if task != nil && task.ColumnID == nil && task.Status != nil {
		rawStatus := strings.TrimSpace(*task.Status)
		if rawStatus != "" {
			matched := false
			for i := range options {
				if strings.EqualFold(options[i].Label, rawStatus) || strings.EqualFold(options[i].Status, rawStatus) {
					selected = i
					matched = true
					break
				}
			}
			if !matched {
				options = append(options, taskStatusOption{
					Status: rawStatus,
					Label:  rawStatus,
				})
				selected = len(options) - 1
			}
		}
	}

	if len(options) == 0 {
		options = append(options, taskStatusOption{Label: "(none)"})
		selected = 0
	}

	return options, selected
}

func (f *taskForm) currentInputField() *textinput.Model {
	switch f.focus {
	case taskFieldTitle:
		return &f.title
	case taskFieldDescription:
		return &f.description
	case taskFieldDueDate:
		return &f.dueDate
	default:
		return nil
	}
}

func (f *taskForm) setFocus(index int) {
	if index < 0 {
		index = taskFieldCount - 1
	}
	if index >= taskFieldCount {
		index = 0
	}
	f.focus = index

	f.title.Blur()
	f.description.Blur()
	f.dueDate.Blur()
	if field := f.currentInputField(); field != nil {
		field.Focus()
	}
}

func (f *taskForm) moveFocus(delta int) {
	f.setFocus(f.focus + delta)
}

func (f *taskForm) clampPriorityIndex() {
	if len(taskPriorityOptions) == 0 {
		f.priorityIndex = 0
		return
	}
	if f.priorityIndex < 0 {
		f.priorityIndex = 0
	}
	if f.priorityIndex >= len(taskPriorityOptions) {
		f.priorityIndex = len(taskPriorityOptions) - 1
	}
}

func (f *taskForm) cyclePriority(delta int) bool {
	if len(taskPriorityOptions) == 0 {
		return false
	}
	f.clampPriorityIndex()
	f.priorityIndex = (f.priorityIndex + delta + len(taskPriorityOptions)) % len(taskPriorityOptions)
	return true
}

func (f *taskForm) selectedPriority() int {
	f.clampPriorityIndex()
	return taskPriorityOptions[f.priorityIndex].Value
}

func (f *taskForm) selectedPriorityLabel() string {
	f.clampPriorityIndex()
	option := taskPriorityOptions[f.priorityIndex]
	return fmt.Sprintf("%d - %s", option.Value, option.Label)
}

func (f *taskForm) clampStatusIndex() {
	if len(f.statusOptions) == 0 {
		f.statusIndex = 0
		return
	}
	if f.statusIndex < 0 {
		f.statusIndex = 0
	}
	if f.statusIndex >= len(f.statusOptions) {
		f.statusIndex = len(f.statusOptions) - 1
	}
}

func (f *taskForm) cycleStatus(delta int) bool {
	if len(f.statusOptions) == 0 {
		return false
	}
	f.clampStatusIndex()
	f.statusIndex = (f.statusIndex + delta + len(f.statusOptions)) % len(f.statusOptions)
	return true
}

func (f *taskForm) selectedStatus() (*string, *string) {
	if len(f.statusOptions) == 0 {
		return nil, nil
	}
	f.clampStatusIndex()
	option := f.statusOptions[f.statusIndex]

	var columnID *string
	if option.ColumnID != "" {
		value := option.ColumnID
		columnID = &value
	}

	var status *string
	if option.Status != "" {
		value := option.Status
		status = &value
	}

	return columnID, status
}

func (f *taskForm) selectedStatusLabel() string {
	if len(f.statusOptions) == 0 {
		return "(none)"
	}
	f.clampStatusIndex()
	return f.statusOptions[f.statusIndex].Label
}

func (f *taskForm) modeLabel() string {
	if f.mode == taskFormEdit {
		return "Edit Task"
	}
	return "Create Task"
}

func (m Model) submitTaskFormCmd() (tea.Cmd, error) {
	if m.taskForm == nil {
		return nil, fmt.Errorf("task form is not active")
	}

	title := strings.TrimSpace(m.taskForm.title.Value())
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	priority := m.taskForm.selectedPriority()

	dueAt, err := m.parseDueDateInput(m.taskForm.dueDate.Value())
	if err != nil {
		return nil, err
	}

	columnID, status := m.taskForm.selectedStatus()
	description := strings.TrimSpace(m.taskForm.descriptionFull)
	if description == "" {
		description = strings.TrimSpace(m.taskForm.description.Value())
	}

	if m.taskForm.mode == taskFormCreate {
		boardID := m.boardID
		return m.createTaskWithDetailsCmd(title, description, priority, dueAt, &boardID, columnID, status), nil
	}

	return m.updateTaskWithDetailsCmd(
		m.taskForm.taskID,
		&title,
		&description,
		&priority,
		dueAt,
		columnID,
		status,
	), nil
}

func (m Model) enterCreateTaskForm() (tea.Model, tea.Cmd) {
	m.startCreateTaskForm()
	return m, textinput.Blink
}

func (m Model) enterEditTaskForm(task domain.Task) (tea.Model, tea.Cmd) {
	m.startEditTaskForm(task)
	return m, textinput.Blink
}

func (m Model) renderTaskFormOverlay(base string) string {
	if m.inputMode == inputTaskForm && m.taskForm != nil {
		return m.renderTaskFormPanel(base)
	}
	return base
}

func (m Model) renderTaskFormPanel(base string) string {
	_ = base
	if m.taskForm == nil {
		return base
	}

	panelWidth := m.width * 4 / 5
	if panelWidth < 72 {
		panelWidth = 72
	}
	if panelWidth > m.width-2 {
		panelWidth = max(20, m.width-2)
	}

	panelHeight := m.height * 3 / 4
	if panelHeight < 18 {
		panelHeight = 18
	}
	if panelHeight > m.height-2 {
		panelHeight = max(10, m.height-2)
	}

	contentWidth := boxContentWidth(panelWidth, 1, true)
	contentHeight := boxContentHeight(panelHeight, true)

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("151")).Render(m.taskForm.modeLabel())
	hints := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("Tab/Shift+Tab navigate | \u2190/\u2192 select priority/status | 0-5 priority | Ctrl+S save | Ctrl+G edit description in $EDITOR | Esc cancel")

	titleLabel := m.renderTaskFieldLabel(taskFieldTitle, "Title")
	descLabel := m.renderTaskFieldLabel(taskFieldDescription, "Description")
	dueLabel := m.renderTaskFieldLabel(taskFieldDueDate, "Due Date")
	priorityLabel := m.renderTaskFieldLabel(taskFieldPriority, "Priority")
	statusLabel := m.renderTaskFieldLabel(taskFieldStatus, "Status")

	descPreview := strings.TrimSpace(m.taskForm.descriptionFull)
	descPreviewLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("Description preview:")
	descPreviewStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(max(12, contentWidth-2))
	previewLines := []string{}
	if descPreview == "" {
		previewLines = []string{"(empty)"}
	} else {
		previewLines = wrapViewerText(descPreview, max(12, contentWidth-4))
		maxPreviewLines := max(3, min(8, contentHeight/3))
		if len(previewLines) > maxPreviewLines {
			previewLines = append(previewLines[:maxPreviewLines-1], "...")
		}
	}
	priorityValue := lipgloss.NewStyle().
		Foreground(priorityColor(m.taskForm.selectedPriority())).
		Bold(true).
		Render(m.taskForm.selectedPriorityLabel())

	statusValueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true)
	if len(m.taskForm.statusOptions) > 0 {
		m.taskForm.clampStatusIndex()
		statusColor := colorFromHexOrDefault(m.taskForm.statusOptions[m.taskForm.statusIndex].ColorHex, "252")
		statusValueStyle = statusValueStyle.Foreground(statusColor)
	}
	statusValue := statusValueStyle.Render(m.taskForm.selectedStatusLabel())

	lines := []string{
		title,
		hints,
		"",
		fmt.Sprintf("%s %s", titleLabel, m.taskForm.title.View()),
		fmt.Sprintf("%s %s", descLabel, m.taskForm.description.View()),
		descPreviewLabel,
		fmt.Sprintf("%s %s", dueLabel, m.taskForm.dueDate.View()),
		fmt.Sprintf("%s %s", priorityLabel, priorityValue),
		fmt.Sprintf("%s %s", statusLabel, statusValue),
	}
	for _, line := range previewLines {
		lines = append(lines, descPreviewStyle.Render(line))
	}

	if m.taskForm.focus == taskFieldPriority {
		lines = append(lines, m.renderTaskPrioritySelectorLines(max(12, contentWidth-4))...)
	}
	if m.taskForm.focus == taskFieldStatus {
		lines = append(lines, m.renderTaskStatusSelectorLines(max(12, contentWidth-4), max(4, contentHeight/3))...)
	}

	panel := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("151")).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}

func (m Model) renderTaskPrioritySelectorLines(width int) []string {
	lines := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render("Priority options:"),
	}

	for i, option := range taskPriorityOptions {
		base := fmt.Sprintf("  %d) %s", option.Value, option.Label)
		style := lipgloss.NewStyle().Width(width).Foreground(priorityColor(option.Value))
		if m.taskForm != nil && i == m.taskForm.priorityIndex {
			style = style.Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Bold(true)
		}
		lines = append(lines, style.Render(base))
	}

	return lines
}

func (m Model) renderTaskStatusSelectorLines(width, maxVisible int) []string {
	lines := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render("Status options:"),
	}
	if m.taskForm == nil || len(m.taskForm.statusOptions) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("  (none)"))
		return lines
	}

	if maxVisible < 1 {
		maxVisible = 1
	}

	selected := m.taskForm.statusIndex
	offset := 0
	if selected >= maxVisible {
		offset = selected - maxVisible + 1
	}
	maxOffset := max(0, len(m.taskForm.statusOptions)-maxVisible)
	if offset > maxOffset {
		offset = maxOffset
	}

	end := min(len(m.taskForm.statusOptions), offset+maxVisible)
	if offset > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  ↑ more"))
	}
	for i := offset; i < end; i++ {
		option := m.taskForm.statusOptions[i]
		base := "  " + option.Label
		color := colorFromHexOrDefault(option.ColorHex, "252")
		style := lipgloss.NewStyle().Width(width).Foreground(color)
		if i == selected {
			style = style.Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Bold(true)
		}
		lines = append(lines, style.Render(base))
	}
	if end < len(m.taskForm.statusOptions) {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  ↓ more"))
	}

	return lines
}

func (m Model) renderTaskFieldLabel(field int, label string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Width(12)
	if m.taskForm != nil && m.taskForm.focus == field {
		style = style.Foreground(lipgloss.Color("221")).Bold(true)
	}
	return style.Render(label + ":")
}
