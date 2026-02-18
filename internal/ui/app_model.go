package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tiagokriok/lazytask/internal/application"
	"github.com/tiagokriok/lazytask/internal/domain"
)

type viewMode int

const (
	viewList viewMode = iota
	viewKanban
)

type inputMode int

const (
	inputNone inputMode = iota
	inputSearch
	inputAddComment
	inputEditDescription
	inputTaskForm
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
	priority    textinput.Model
	status      textinput.Model

	descriptionFull string
}

func (f *taskForm) fields() []*textinput.Model {
	return []*textinput.Model{&f.title, &f.description, &f.dueDate, &f.priority, &f.status}
}

func (f *taskForm) currentField() *textinput.Model {
	fields := f.fields()
	if f.focus < 0 {
		f.focus = 0
	}
	if f.focus >= len(fields) {
		f.focus = len(fields) - 1
	}
	return fields[f.focus]
}

func (f *taskForm) setFocus(index int) {
	fields := f.fields()
	if len(fields) == 0 {
		f.focus = 0
		return
	}
	if index < 0 {
		index = len(fields) - 1
	}
	if index >= len(fields) {
		index = 0
	}
	f.focus = index
	for i, field := range fields {
		if i == f.focus {
			field.Focus()
		} else {
			field.Blur()
		}
	}
}

func (f *taskForm) moveFocus(delta int) {
	f.setFocus(f.focus + delta)
}

func (f *taskForm) modeLabel() string {
	if f.mode == taskFormEdit {
		return "Edit Task"
	}
	return "Create Task"
}

type tasksLoadedMsg struct {
	tasks []domain.Task
	err   error
}

type commentsLoadedMsg struct {
	comments []domain.Comment
	err      error
}

type opResultMsg struct {
	status string
	err    error
}

type executeActionMsg struct {
	action string
}

type descriptionEditedMsg struct {
	content string
	err     error
}

type keybindEntry struct {
	ID    string
	Key   string
	Label string
}

type Model struct {
	taskService    *application.TaskService
	commentService *application.CommentService

	dateFormat userDateFormat

	providerID    string
	workspaceID   string
	workspaceName string
	boardID       string
	boardName     string
	columns       []domain.Column

	tasks    []domain.Task
	comments []domain.Comment

	selected      int
	activeColumn  int
	kanbanRow     int
	columnFilter  string
	filterIndex   int
	titleFilter   string
	dueSoonFilter bool

	viewMode     viewMode
	showDetails  bool
	inputMode    inputMode
	taskForm     *taskForm
	showKeybinds bool

	textInput textinput.Model
	textArea  textarea.Model

	keyFilter   textinput.Model
	keySelected int

	statusLine string
	err        error

	width  int
	height int

	keys keyMap
}

func NewModel(taskService *application.TaskService, commentService *application.CommentService, setup application.BootstrapResult) Model {
	ti := textinput.New()
	ti.Placeholder = "Type..."
	ti.CharLimit = 512
	ti.Prompt = "> "

	ta := textarea.New()
	ta.Placeholder = "Markdown..."
	ta.SetHeight(8)
	ta.Prompt = ""

	kf := textinput.New()
	kf.Placeholder = "Filter keybindings..."
	kf.Prompt = "Filter: "
	kf.CharLimit = 128

	cols := make([]domain.Column, 0, len(setup.Columns))
	cols = append(cols, setup.Columns...)
	sort.Slice(cols, func(i, j int) bool {
		return cols[i].Position < cols[j].Position
	})

	return Model{
		taskService:    taskService,
		commentService: commentService,
		dateFormat:     detectUserDateFormat(),
		providerID:     setup.Provider.ID,
		workspaceID:    setup.Workspace.ID,
		workspaceName:  setup.Workspace.Name,
		boardID:        setup.Board.ID,
		boardName:      setup.Board.Name,
		columns:        cols,
		filterIndex:    -1,
		viewMode:       viewList,
		showDetails:    true,
		textInput:      ti,
		textArea:       ta,
		keyFilter:      kf,
		keys:           newKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadTasksCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showKeybinds {
		return m.updateKeybindPanel(msg)
	}
	if m.inputMode != inputNone {
		return m.updateInputMode(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textArea.SetWidth(max(20, msg.Width/2-6))
		m.keyFilter.Width = max(24, msg.Width/3)
		return m, nil
	case executeActionMsg:
		return m.executeAction(msg.action)
	case tasksLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.statusLine = msg.err.Error()
			return m, nil
		}
		m.sortTasksByPriority(msg.tasks)
		m.tasks = msg.tasks
		m.ensureSelection()
		if m.showDetails {
			if task, ok := m.currentTask(); ok {
				return m, m.loadCommentsCmd(task.ID)
			}
		}
		m.comments = nil
		return m, nil
	case commentsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.statusLine = msg.err.Error()
			return m, nil
		}
		m.comments = msg.comments
		return m, nil
	case opResultMsg:
		if msg.err != nil {
			m.err = msg.err
			m.statusLine = msg.err.Error()
			return m, nil
		}
		m.statusLine = ""
		return m, m.loadTasksCmd()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.ShowKeybinds):
			m.openKeybindPanel()
			return m, textinput.Blink
		case key.Matches(msg, m.keys.ToggleView):
			if m.viewMode == viewList {
				m.viewMode = viewKanban
			} else {
				m.viewMode = viewList
			}
			m.ensureSelection()
			if m.showDetails {
				if task, ok := m.currentTask(); ok {
					return m, m.loadCommentsCmd(task.ID)
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.ToggleDetails):
			m.showDetails = !m.showDetails
			if m.showDetails {
				if task, ok := m.currentTask(); ok {
					return m, m.loadCommentsCmd(task.ID)
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.Search):
			m.inputMode = inputSearch
			m.textInput.SetValue(m.titleFilter)
			m.textInput.Placeholder = "Search title"
			m.textInput.Focus()
			m.statusLine = "Search by title"
			return m, textinput.Blink
		case key.Matches(msg, m.keys.ClearSearch):
			if strings.TrimSpace(m.titleFilter) == "" {
				return m, nil
			}
			m.titleFilter = ""
			m.statusLine = ""
			return m, m.loadTasksCmd()
		case key.Matches(msg, m.keys.NewTask):
			m.startCreateTaskForm()
			return m, textinput.Blink
		case key.Matches(msg, m.keys.EditTitle):
			task, ok := m.currentTask()
			if !ok {
				return m, nil
			}
			m.startEditTaskForm(task)
			return m, textinput.Blink
		case key.Matches(msg, m.keys.AddComment):
			if _, ok := m.currentTask(); !ok {
				return m, nil
			}
			m.inputMode = inputAddComment
			m.textInput.SetValue("")
			m.textInput.Placeholder = "Comment body"
			m.textInput.Focus()
			m.statusLine = "Add comment"
			return m, textinput.Blink
		case key.Matches(msg, m.keys.EditDescription):
			task, ok := m.currentTask()
			if !ok {
				return m, nil
			}
			m.inputMode = inputEditDescription
			m.textArea.SetValue(task.DescriptionMD)
			m.textArea.Focus()
			m.statusLine = "Edit description (Ctrl+S save, Esc cancel)"
			return m, nil
		case key.Matches(msg, m.keys.CycleStatus):
			m.cycleColumnFilter()
			return m, m.loadTasksCmd()
		case key.Matches(msg, m.keys.ToggleDueSoon):
			m.dueSoonFilter = !m.dueSoonFilter
			return m, m.loadTasksCmd()
		case key.Matches(msg, m.keys.Up):
			m.moveUp()
			if m.showDetails {
				if task, ok := m.currentTask(); ok {
					return m, m.loadCommentsCmd(task.ID)
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.Down):
			m.moveDown()
			if m.showDetails {
				if task, ok := m.currentTask(); ok {
					return m, m.loadCommentsCmd(task.ID)
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.Left):
			if m.viewMode == viewKanban {
				if m.activeColumn > 0 {
					m.activeColumn--
				}
				m.ensureKanbanRow()
				if m.showDetails {
					if task, ok := m.currentTask(); ok {
						return m, m.loadCommentsCmd(task.ID)
					}
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.Right):
			if m.viewMode == viewKanban {
				if m.activeColumn < len(m.columns)-1 {
					m.activeColumn++
				}
				m.ensureKanbanRow()
				if m.showDetails {
					if task, ok := m.currentTask(); ok {
						return m, m.loadCommentsCmd(task.ID)
					}
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.OpenDetails):
			if m.viewMode == viewKanban {
				if task, ok := m.currentTask(); ok {
					return m, m.moveToNextColumnCmd(task)
				}
				return m, nil
			}
			m.showDetails = true
			if task, ok := m.currentTask(); ok {
				return m, m.loadCommentsCmd(task.ID)
			}
			return m, nil
		case key.Matches(msg, m.keys.MoveTask):
			if task, ok := m.currentTask(); ok {
				return m, m.moveToNextColumnCmd(task)
			}
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var base string
	if m.viewMode == viewList {
		base = m.renderListScreen()
	} else {
		header := m.renderHeader()
		footer := m.renderFooter()
		bodyHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
		if bodyHeight < 5 {
			bodyHeight = 5
		}

		mainPane := m.renderKanbanView(bodyHeight)
		if m.showDetails {
			detailWidth := m.width / 3
			if detailWidth < 34 {
				detailWidth = 34
			}
			mainWidth := m.width - detailWidth - 1
			if mainWidth < 20 {
				mainWidth = m.width
				detailWidth = 0
			}
			mainPane = lipgloss.NewStyle().Width(mainWidth).Render(mainPane)
			if detailWidth > 0 {
				detailPane := m.renderDetailView(detailWidth, bodyHeight)
				mainPane = lipgloss.JoinHorizontal(lipgloss.Top, mainPane, detailPane)
			}
		}

		content := lipgloss.JoinVertical(lipgloss.Left, header, mainPane, footer)
		base = lipgloss.NewStyle().Padding(0, 1).Render(content)
	}

	if m.showKeybinds {
		return m.renderKeybindPanel(base)
	}
	if m.inputMode == inputTaskForm && m.taskForm != nil {
		return m.renderTaskFormPanel(base)
	}
	return base
}

func (m Model) updateInputMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	mode := m.inputMode

	switch msg := msg.(type) {
	case descriptionEditedMsg:
		if mode == inputTaskForm && m.taskForm != nil {
			if msg.err != nil {
				m.statusLine = fmt.Sprintf("editor error: %v", msg.err)
				return m, nil
			}
			m.taskForm.descriptionFull = msg.content
			m.taskForm.description.SetValue(summarizeDescription(msg.content))
			m.statusLine = ""
			return m, nil
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel):
			if mode == inputTaskForm {
				m.closeTaskForm()
				return m, nil
			}
			m.inputMode = inputNone
			m.textInput.Blur()
			m.textArea.Blur()
			m.statusLine = ""
			return m, nil
		case mode == inputTaskForm:
			if handledModel, cmd, handled := m.handleTaskFormKey(msg); handled {
				return handledModel, cmd
			}
		case msg.String() == "ctrl+s" && mode == inputEditDescription:
			task, ok := m.currentTask()
			if !ok {
				m.inputMode = inputNone
				return m, nil
			}
			description := m.textArea.Value()
			m.inputMode = inputNone
			m.textArea.Blur()
			return m, m.updateTaskDescriptionCmd(task.ID, description)
		case key.Matches(msg, m.keys.Confirm) && mode != inputEditDescription:
			value := strings.TrimSpace(m.textInput.Value())
			m.inputMode = inputNone
			m.textInput.Blur()
			switch mode {
			case inputSearch:
				m.titleFilter = value
				m.statusLine = ""
				return m, m.loadTasksCmd()
			case inputAddComment:
				task, ok := m.currentTask()
				if !ok {
					return m, nil
				}
				if value == "" {
					m.statusLine = "comment is required"
					return m, nil
				}
				return m, m.addCommentCmd(task.ID, value)
			}
		}
	}

	if mode == inputEditDescription {
		var cmd tea.Cmd
		m.textArea, cmd = m.textArea.Update(msg)
		m.inputMode = mode
		return m, cmd
	}

	if mode == inputTaskForm && m.taskForm != nil {
		field := m.taskForm.currentField()
		var cmd tea.Cmd
		*field, cmd = field.Update(msg)
		if m.taskForm.focus == taskFieldDescription {
			m.taskForm.descriptionFull = m.taskForm.description.Value()
		}
		return m, cmd
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.inputMode = mode
	return m, cmd
}

func (m *Model) handleTaskFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
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
	case "ctrl+s":
		model, cmd := m.submitTaskForm()
		return model, cmd, true
	case "ctrl+g":
		if m.taskForm.focus == taskFieldDescription {
			return m, openDescriptionEditorCmd(m.taskForm.descriptionFull), true
		}
		return m, nil, true
	case "enter":
		if m.taskForm.focus == taskFieldCount-1 {
			model, cmd := m.submitTaskForm()
			return model, cmd, true
		}
		m.taskForm.moveFocus(1)
		return m, textinput.Blink, true
	}

	return m, nil, false
}

func (m *Model) submitTaskForm() (tea.Model, tea.Cmd) {
	cmd, err := m.submitTaskFormCmd()
	if err != nil {
		m.statusLine = err.Error()
		return m, nil
	}
	m.closeTaskForm()
	return m, cmd
}

func (m *Model) closeTaskForm() {
	m.inputMode = inputNone
	m.taskForm = nil
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
	status := ""
	if len(m.columns) > 0 {
		status = m.columns[0].Name
	}

	form := &taskForm{
		mode:            taskFormCreate,
		title:           newTaskFormInput("Title", "", 512),
		description:     newTaskFormInput("Description", "", 2048),
		dueDate:         newTaskFormInput(m.dueDatePlaceholder(), "", 32),
		priority:        newTaskFormInput("Priority 0-5", "0", 2),
		status:          newTaskFormInput("Status", status, 64),
		descriptionFull: "",
	}
	form.setFocus(taskFieldTitle)

	m.taskForm = form
	m.inputMode = inputTaskForm
	m.statusLine = "Create task"
}

func (m *Model) startEditTaskForm(task domain.Task) {
	status := ""
	if task.ColumnID != nil {
		status = m.columnName(*task.ColumnID)
	} else if task.Status != nil {
		status = *task.Status
	}

	due := ""
	if task.DueAt != nil {
		due = m.formatDueDate(*task.DueAt)
	}

	form := &taskForm{
		mode:            taskFormEdit,
		taskID:          task.ID,
		title:           newTaskFormInput("Title", task.Title, 512),
		description:     newTaskFormInput("Description", summarizeDescription(task.DescriptionMD), 2048),
		dueDate:         newTaskFormInput(m.dueDatePlaceholder(), due, 32),
		priority:        newTaskFormInput("Priority 0-5", strconv.Itoa(task.Priority), 2),
		status:          newTaskFormInput("Status", status, 64),
		descriptionFull: task.DescriptionMD,
	}
	form.setFocus(taskFieldTitle)

	m.taskForm = form
	m.inputMode = inputTaskForm
	m.statusLine = "Edit task"
}

func summarizeDescription(description string) string {
	trimmed := strings.TrimSpace(description)
	if trimmed == "" {
		return ""
	}
	firstLine := strings.Split(trimmed, "\n")[0]
	if len(firstLine) > 80 {
		return firstLine[:77] + "..."
	}
	return firstLine
}

func chooseEditor() string {
	if editor := strings.TrimSpace(os.Getenv("EDITOR")); editor != "" {
		return editor
	}
	for _, candidate := range []string{"nvim", "vim", "vi"} {
		if _, err := exec.LookPath(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func openDescriptionEditorCmd(initial string) tea.Cmd {
	editor := chooseEditor()
	if editor == "" {
		return func() tea.Msg {
			return descriptionEditedMsg{err: fmt.Errorf("no editor found (set $EDITOR or install nvim/vim/vi)")}
		}
	}

	tmpFile, err := os.CreateTemp("", "lazytask-description-*.md")
	if err != nil {
		return func() tea.Msg {
			return descriptionEditedMsg{err: err}
		}
	}

	path := tmpFile.Name()
	if _, err := tmpFile.WriteString(initial); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(path)
		return func() tea.Msg {
			return descriptionEditedMsg{err: err}
		}
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(path)
		return func() tea.Msg {
			return descriptionEditedMsg{err: err}
		}
	}

	parts := strings.Fields(editor)
	if len(parts) == 0 {
		_ = os.Remove(path)
		return func() tea.Msg {
			return descriptionEditedMsg{err: fmt.Errorf("invalid editor command")}
		}
	}

	cmd := exec.Command(parts[0], append(parts[1:], path)...)
	return tea.ExecProcess(cmd, func(execErr error) tea.Msg {
		defer os.Remove(path)
		if execErr != nil {
			return descriptionEditedMsg{err: execErr}
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return descriptionEditedMsg{err: readErr}
		}
		return descriptionEditedMsg{content: strings.TrimRight(string(content), "\n")}
	})
}

func (m Model) submitTaskFormCmd() (tea.Cmd, error) {
	if m.taskForm == nil {
		return nil, fmt.Errorf("task form is not active")
	}

	title := strings.TrimSpace(m.taskForm.title.Value())
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	priorityRaw := strings.TrimSpace(m.taskForm.priority.Value())
	if priorityRaw == "" {
		return nil, fmt.Errorf("priority is required (0-5)")
	}
	priority, err := strconv.Atoi(priorityRaw)
	if err != nil {
		return nil, fmt.Errorf("priority must be an integer between 0 and 5")
	}
	if priority < 0 || priority > 5 {
		return nil, fmt.Errorf("priority must be between 0 and 5")
	}

	dueAt, err := m.parseDueDateInput(m.taskForm.dueDate.Value())
	if err != nil {
		return nil, err
	}

	columnID, status := m.resolveStatusInput(strings.TrimSpace(m.taskForm.status.Value()), m.taskForm.mode)
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
	hints := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("Tab/Shift+Tab navigate | Ctrl+S save | Ctrl+G edit description in $EDITOR | Esc cancel")

	titleLabel := m.renderTaskFieldLabel(taskFieldTitle, "Title")
	descLabel := m.renderTaskFieldLabel(taskFieldDescription, "Description")
	dueLabel := m.renderTaskFieldLabel(taskFieldDueDate, "Due Date")
	priorityLabel := m.renderTaskFieldLabel(taskFieldPriority, "Priority")
	statusLabel := m.renderTaskFieldLabel(taskFieldStatus, "Status")

	descPreview := m.taskForm.descriptionFull
	if strings.TrimSpace(descPreview) == "" {
		descPreview = "(empty)"
	}
	descPreviewStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(contentWidth - 2)

	lines := []string{
		title,
		hints,
		"",
		fmt.Sprintf("%s %s", titleLabel, m.taskForm.title.View()),
		fmt.Sprintf("%s %s", descLabel, m.taskForm.description.View()),
		descPreviewStyle.Render(descPreview),
		fmt.Sprintf("%s %s", dueLabel, m.taskForm.dueDate.View()),
		fmt.Sprintf("%s %s", priorityLabel, m.taskForm.priority.View()),
		fmt.Sprintf("%s %s", statusLabel, m.taskForm.status.View()),
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

func (m Model) renderTaskFieldLabel(field int, label string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Width(12)
	if m.taskForm != nil && m.taskForm.focus == field {
		style = style.Foreground(lipgloss.Color("221")).Bold(true)
	}
	return style.Render(label + ":")
}

func (m *Model) openKeybindPanel() {
	m.showKeybinds = true
	m.keySelected = 0
	m.keyFilter.SetValue("")
	m.keyFilter.Width = max(24, m.width/3)
	m.keyFilter.Focus()
}

func (m *Model) closeKeybindPanel() {
	m.showKeybinds = false
	m.keyFilter.Blur()
}

func (m Model) keybindEntries() []keybindEntry {
	entries := []keybindEntry{
		{ID: "new_task", Key: "n", Label: "Create task"},
		{ID: "edit_task", Key: "e", Label: "Edit selected task"},
		{ID: "edit_description", Key: "E", Label: "Edit description"},
		{ID: "add_comment", Key: "c", Label: "Add comment"},
		{ID: "search", Key: "/", Label: "Search"},
		{ID: "toggle_details", Key: "d", Label: "Toggle details pane"},
		{ID: "open_move", Key: "Enter", Label: "Open details / move in kanban"},
		{ID: "move_task", Key: "m", Label: "Move task to next status"},
		{ID: "toggle_view", Key: "Tab", Label: "Switch list/kanban"},
		{ID: "cycle_status", Key: "s", Label: "Cycle status filter"},
		{ID: "toggle_due_soon", Key: "z", Label: "Toggle due soon filter"},
		{ID: "move_up", Key: "↑", Label: "Move selection up"},
		{ID: "move_down", Key: "↓", Label: "Move selection down"},
		{ID: "move_left", Key: "←", Label: "Move left (kanban)"},
		{ID: "move_right", Key: "→", Label: "Move right (kanban)"},
		{ID: "quit", Key: "q", Label: "Quit"},
	}
	if strings.TrimSpace(m.titleFilter) != "" {
		entries = append(entries, keybindEntry{ID: "clear_search", Key: "x", Label: "Clear search"})
	}
	return entries
}

func (m Model) filteredKeybindEntries() []keybindEntry {
	query := strings.ToLower(strings.TrimSpace(m.keyFilter.Value()))
	entries := m.keybindEntries()
	if query == "" {
		return entries
	}

	filtered := make([]keybindEntry, 0, len(entries))
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Key), query) || strings.Contains(strings.ToLower(entry.Label), query) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (m *Model) clampKeybindSelection() {
	entries := m.filteredKeybindEntries()
	if len(entries) == 0 {
		m.keySelected = 0
		return
	}
	if m.keySelected < 0 {
		m.keySelected = 0
	}
	if m.keySelected >= len(entries) {
		m.keySelected = len(entries) - 1
	}
}

func (m Model) updateKeybindPanel(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.keyFilter.Width = max(24, msg.Width/3)
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel), key.Matches(msg, m.keys.ShowKeybinds):
			m.closeKeybindPanel()
			return m, nil
		case key.Matches(msg, m.keys.Up):
			m.keySelected--
			m.clampKeybindSelection()
			return m, nil
		case key.Matches(msg, m.keys.Down):
			m.keySelected++
			m.clampKeybindSelection()
			return m, nil
		case key.Matches(msg, m.keys.Confirm):
			entries := m.filteredKeybindEntries()
			if len(entries) == 0 {
				return m, nil
			}
			action := entries[m.keySelected].ID
			m.closeKeybindPanel()
			return m, func() tea.Msg { return executeActionMsg{action: action} }
		}
	}

	var cmd tea.Cmd
	m.keyFilter, cmd = m.keyFilter.Update(msg)
	m.clampKeybindSelection()
	return m, cmd
}

func (m Model) renderKeybindPanel(base string) string {
	_ = base

	panelWidth := m.width * 3 / 4
	if panelWidth < 64 {
		panelWidth = 64
	}
	if panelWidth > m.width-2 {
		panelWidth = max(20, m.width-2)
	}

	panelHeight := m.height * 3 / 4
	if panelHeight < 12 {
		panelHeight = 12
	}
	if panelHeight > m.height-2 {
		panelHeight = max(8, m.height-2)
	}

	contentWidth := boxContentWidth(panelWidth, 1, true)
	contentHeight := boxContentHeight(panelHeight, true)
	listHeight := max(1, contentHeight-5)

	entries := m.filteredKeybindEntries()
	offset := 0
	if m.keySelected >= listHeight {
		offset = m.keySelected - listHeight + 1
	}
	maxOffset := max(0, len(entries)-listHeight)
	if offset > maxOffset {
		offset = maxOffset
	}

	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("151")).Render("Keybindings"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("Type to filter | Enter: execute | Esc: close"),
		m.keyFilter.View(),
		"",
	}

	if len(entries) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("(no keybindings match filter)"))
	} else {
		end := offset + listHeight
		if end > len(entries) {
			end = len(entries)
		}
		for i := offset; i < end; i++ {
			entry := entries[i]
			keyLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Width(12).Render(entry.Key)
			descLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(entry.Label)
			line := keyLabel + " " + descLabel
			if i == m.keySelected {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Render(line)
			}
			lines = append(lines, line)
		}
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

func (m Model) executeAction(action string) (tea.Model, tea.Cmd) {
	switch action {
	case "quit":
		return m, tea.Quit
	case "toggle_view":
		if m.viewMode == viewList {
			m.viewMode = viewKanban
		} else {
			m.viewMode = viewList
		}
		m.ensureSelection()
		if m.showDetails {
			if task, ok := m.currentTask(); ok {
				return m, m.loadCommentsCmd(task.ID)
			}
		}
		return m, nil
	case "toggle_details":
		m.showDetails = !m.showDetails
		if m.showDetails {
			if task, ok := m.currentTask(); ok {
				return m, m.loadCommentsCmd(task.ID)
			}
		}
		return m, nil
	case "search":
		m.inputMode = inputSearch
		m.textInput.SetValue(m.titleFilter)
		m.textInput.Placeholder = "Search title"
		m.textInput.Focus()
		m.statusLine = "Search by title"
		return m, textinput.Blink
	case "clear_search":
		if strings.TrimSpace(m.titleFilter) == "" {
			return m, nil
		}
		m.titleFilter = ""
		m.statusLine = ""
		return m, m.loadTasksCmd()
	case "new_task":
		m.startCreateTaskForm()
		return m, textinput.Blink
	case "edit_task":
		task, ok := m.currentTask()
		if !ok {
			return m, nil
		}
		m.startEditTaskForm(task)
		return m, textinput.Blink
	case "add_comment":
		if _, ok := m.currentTask(); !ok {
			return m, nil
		}
		m.inputMode = inputAddComment
		m.textInput.SetValue("")
		m.textInput.Placeholder = "Comment body"
		m.textInput.Focus()
		m.statusLine = "Add comment"
		return m, textinput.Blink
	case "edit_description":
		task, ok := m.currentTask()
		if !ok {
			return m, nil
		}
		m.inputMode = inputEditDescription
		m.textArea.SetValue(task.DescriptionMD)
		m.textArea.Focus()
		m.statusLine = "Edit description (Ctrl+S save, Esc cancel)"
		return m, nil
	case "cycle_status":
		m.cycleColumnFilter()
		return m, m.loadTasksCmd()
	case "toggle_due_soon":
		m.dueSoonFilter = !m.dueSoonFilter
		return m, m.loadTasksCmd()
	case "move_up":
		m.moveUp()
		if m.showDetails {
			if task, ok := m.currentTask(); ok {
				return m, m.loadCommentsCmd(task.ID)
			}
		}
		return m, nil
	case "move_down":
		m.moveDown()
		if m.showDetails {
			if task, ok := m.currentTask(); ok {
				return m, m.loadCommentsCmd(task.ID)
			}
		}
		return m, nil
	case "move_left":
		if m.viewMode == viewKanban {
			if m.activeColumn > 0 {
				m.activeColumn--
			}
			m.ensureKanbanRow()
			if m.showDetails {
				if task, ok := m.currentTask(); ok {
					return m, m.loadCommentsCmd(task.ID)
				}
			}
		}
		return m, nil
	case "move_right":
		if m.viewMode == viewKanban {
			if m.activeColumn < len(m.columns)-1 {
				m.activeColumn++
			}
			m.ensureKanbanRow()
			if m.showDetails {
				if task, ok := m.currentTask(); ok {
					return m, m.loadCommentsCmd(task.ID)
				}
			}
		}
		return m, nil
	case "open_move":
		if m.viewMode == viewKanban {
			if task, ok := m.currentTask(); ok {
				return m, m.moveToNextColumnCmd(task)
			}
			return m, nil
		}
		m.showDetails = true
		if task, ok := m.currentTask(); ok {
			return m, m.loadCommentsCmd(task.ID)
		}
		return m, nil
	case "move_task":
		if task, ok := m.currentTask(); ok {
			return m, m.moveToNextColumnCmd(task)
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) resolveStatusInput(raw string, mode taskFormMode) (*string, *string) {
	if raw == "" {
		if mode == taskFormCreate && len(m.columns) > 0 {
			columnID := m.columns[0].ID
			status := strings.ToLower(m.columns[0].Name)
			return &columnID, &status
		}
		return nil, nil
	}

	for _, col := range m.columns {
		if strings.EqualFold(col.Name, raw) || col.ID == raw {
			columnID := col.ID
			status := strings.ToLower(col.Name)
			return &columnID, &status
		}
	}

	status := strings.ToLower(raw)
	return nil, &status
}

func (m Model) createTaskWithDetailsCmd(title, description string, priority int, dueAt *time.Time, boardID, columnID, status *string) tea.Cmd {
	service := m.taskService
	providerID := m.providerID
	workspaceID := m.workspaceID

	return func() tea.Msg {
		_, err := service.CreateTask(context.Background(), application.CreateTaskInput{
			ProviderID:    providerID,
			WorkspaceID:   workspaceID,
			BoardID:       boardID,
			ColumnID:      columnID,
			Status:        status,
			Title:         title,
			DescriptionMD: description,
			Priority:      priority,
			DueAt:         dueAt,
			Labels:        []string{},
		})
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: "task created"}
	}
}

func (m Model) updateTaskWithDetailsCmd(taskID string, title, description *string, priority *int, dueAt *time.Time, columnID, status *string) tea.Cmd {
	service := m.taskService
	return func() tea.Msg {
		err := service.UpdateTask(context.Background(), taskID, application.UpdateTaskInput{
			Title:         title,
			DescriptionMD: description,
			Priority:      priority,
			DueAt:         dueAt,
			ColumnID:      columnID,
			Status:        status,
		})
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: "task updated"}
	}
}

func (m Model) updateTaskDescriptionCmd(taskID, description string) tea.Cmd {
	service := m.taskService
	return func() tea.Msg {
		err := service.UpdateTask(context.Background(), taskID, application.UpdateTaskInput{DescriptionMD: &description})
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: "description updated"}
	}
}

func (m Model) addCommentCmd(taskID, body string) tea.Cmd {
	service := m.commentService
	providerID := m.providerID
	return func() tea.Msg {
		_, err := service.AddComment(context.Background(), application.AddCommentInput{
			TaskID:     taskID,
			ProviderID: providerID,
			BodyMD:     body,
		})
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: "comment added"}
	}
}

func (m Model) moveToNextColumnCmd(task domain.Task) tea.Cmd {
	if len(m.columns) == 0 {
		return nil
	}
	current := 0
	if task.ColumnID != nil {
		for i, col := range m.columns {
			if col.ID == *task.ColumnID {
				current = i
				break
			}
		}
	}
	next := (current + 1) % len(m.columns)
	col := m.columns[next]
	columnID := col.ID
	status := strings.ToLower(col.Name)
	service := m.taskService
	return func() tea.Msg {
		err := service.MoveTask(context.Background(), task.ID, &columnID, &status, float64(time.Now().UTC().UnixNano()))
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: fmt.Sprintf("moved to %s", col.Name)}
	}
}

func (m Model) renderHeader() string {
	viewLabel := "List"
	if m.viewMode == viewKanban {
		viewLabel = "Kanban"
	}
	filterLabel := "all"
	if m.filterIndex >= 0 && m.filterIndex < len(m.columns) {
		filterLabel = m.columns[m.filterIndex].Name
	}
	dueLabel := "off"
	if m.dueSoonFilter {
		dueLabel = "7d"
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))

	left := headerStyle.Render(fmt.Sprintf("%s / %s", m.workspaceName, m.boardName))
	right := metaStyle.Render(fmt.Sprintf("view:%s  filter:%s  due:%s  search:%q", viewLabel, filterLabel, dueLabel, m.titleFilter))
	if m.width > 20 {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(m.width/2).Render(left),
			lipgloss.NewStyle().Width(max(1, m.width-m.width/2-2)).Align(lipgloss.Right).Render(right),
		)
	}
	return left + " " + right
}

func (m Model) renderFooter() string {
	inputLine := ""
	switch m.inputMode {
	case inputSearch, inputAddComment:
		inputLine = lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Render(m.textInput.View())
	case inputEditDescription:
		inputLine = lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Render(m.textArea.View())
	}

	shortcuts := "?:keybinds n:new /:search enter:open/move q:quit"
	if strings.TrimSpace(m.titleFilter) != "" {
		shortcuts += " x:clear-search"
	}
	lines := make([]string, 0, 3)
	if strings.TrimSpace(m.statusLine) != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("222")).Render(m.statusLine))
	}
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(shortcuts))
	if inputLine != "" {
		lines = append(lines, inputLine)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *Model) ensureSelection() {
	if m.viewMode == viewKanban {
		if len(m.columns) == 0 {
			m.activeColumn = 0
			m.kanbanRow = 0
			return
		}
		if m.activeColumn < 0 {
			m.activeColumn = 0
		}
		if m.activeColumn >= len(m.columns) {
			m.activeColumn = len(m.columns) - 1
		}
		m.ensureKanbanRow()
		return
	}
	if len(m.tasks) == 0 {
		m.selected = 0
		return
	}
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= len(m.tasks) {
		m.selected = len(m.tasks) - 1
	}
}

func (m *Model) ensureKanbanRow() {
	if len(m.columns) == 0 {
		m.kanbanRow = 0
		return
	}
	colID := m.columns[m.activeColumn].ID
	tasks := m.tasksForColumn(colID)
	if len(tasks) == 0 {
		m.kanbanRow = 0
		return
	}
	if m.kanbanRow < 0 {
		m.kanbanRow = 0
	}
	if m.kanbanRow >= len(tasks) {
		m.kanbanRow = len(tasks) - 1
	}
}

func (m *Model) moveUp() {
	if m.viewMode == viewKanban {
		m.kanbanRow--
		m.ensureKanbanRow()
		return
	}
	m.selected--
	m.ensureSelection()
}

func (m *Model) moveDown() {
	if m.viewMode == viewKanban {
		m.kanbanRow++
		m.ensureKanbanRow()
		return
	}
	m.selected++
	m.ensureSelection()
}

func (m *Model) cycleColumnFilter() {
	if len(m.columns) == 0 {
		m.filterIndex = -1
		m.columnFilter = ""
		return
	}
	m.filterIndex++
	if m.filterIndex >= len(m.columns) {
		m.filterIndex = -1
		m.columnFilter = ""
		return
	}
	m.columnFilter = m.columns[m.filterIndex].ID
}

func (m Model) currentTask() (domain.Task, bool) {
	if len(m.tasks) == 0 {
		return domain.Task{}, false
	}
	if m.viewMode == viewKanban {
		if len(m.columns) == 0 {
			return domain.Task{}, false
		}
		col := m.columns[m.activeColumn]
		tasks := m.tasksForColumn(col.ID)
		if len(tasks) == 0 || m.kanbanRow < 0 || m.kanbanRow >= len(tasks) {
			return domain.Task{}, false
		}
		return tasks[m.kanbanRow], true
	}
	if m.selected < 0 || m.selected >= len(m.tasks) {
		return domain.Task{}, false
	}
	return m.tasks[m.selected], true
}

func (m Model) tasksForColumn(columnID string) []domain.Task {
	result := make([]domain.Task, 0)
	for _, t := range m.tasks {
		if t.ColumnID != nil && *t.ColumnID == columnID {
			result = append(result, t)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Position == result[j].Position {
			return result[i].UpdatedAt.After(result[j].UpdatedAt)
		}
		return result[i].Position < result[j].Position
	})
	return result
}

func (m Model) loadTasksCmd() tea.Cmd {
	filters := application.ListTaskFilters{
		WorkspaceID: m.workspaceID,
		TitleQuery:  m.titleFilter,
		ColumnID:    m.columnFilter,
	}
	if m.dueSoonFilter {
		filters.DueSoonDays = 7
	}
	service := m.taskService
	return func() tea.Msg {
		tasks, err := service.ListTasks(context.Background(), filters)
		return tasksLoadedMsg{tasks: tasks, err: err}
	}
}

func (m Model) loadCommentsCmd(taskID string) tea.Cmd {
	service := m.commentService
	return func() tea.Msg {
		comments, err := service.ListComments(context.Background(), taskID)
		return commentsLoadedMsg{comments: comments, err: err}
	}
}

func (m *Model) sortTasksByPriority(tasks []domain.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		pi := normalizePriority(tasks[i].Priority)
		pj := normalizePriority(tasks[j].Priority)
		if pi != pj {
			return pi < pj
		}
		if tasks[i].DueAt != nil && tasks[j].DueAt == nil {
			return true
		}
		if tasks[i].DueAt == nil && tasks[j].DueAt != nil {
			return false
		}
		if tasks[i].DueAt != nil && tasks[j].DueAt != nil && !tasks[i].DueAt.Equal(*tasks[j].DueAt) {
			return tasks[i].DueAt.Before(*tasks[j].DueAt)
		}
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})
}

func normalizePriority(priority int) int {
	if priority < 0 || priority > 5 {
		return 6
	}
	return priority
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
