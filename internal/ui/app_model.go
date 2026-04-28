package ui

import (
	"context"
	"fmt"

	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tiagokriok/kanji/internal/application"
	"github.com/tiagokriok/kanji/internal/domain"
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

type dueFilterMode int

const (
	dueFilterAny dueFilterMode = iota
	dueFilterSoon
	dueFilterOverdue
	dueFilterNoDate
)

type taskSortMode int

const (
	sortByPriority taskSortMode = iota
	sortByDueDate
	sortByTitle
	sortByUpdated
	sortByCreated
)

var taskPriorityOptions = []taskPriorityOption{
	{Value: 0, Label: "Critical"},
	{Value: 1, Label: "Urgent"},
	{Value: 2, Label: "High"},
	{Value: 3, Label: "Medium"},
	{Value: 4, Label: "Low"},
	{Value: 5, Label: "None"},
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
	status   string
	err      error
	taskID   string
	columnID string
}

type descriptionEditedMsg struct {
	content string
	err     error
}

type Model struct {
	overlayState

	taskService    *application.TaskService
	taskFlow       *application.TaskFlow
	commentService *application.CommentService
	contextService *application.ContextService

	dateFormat userDateFormat

	providerID    string
	workspaceID   string
	workspaceName string
	boardID       string
	boardName     string
	columns       []domain.Column
	workspaces    []domain.Workspace
	boards        []domain.Board

	tasks    []domain.Task
	comments []domain.Comment

	selected       int
	activeColumn   int
	kanbanRow      int
	columnFilter   string
	filterIndex    int
	priorityFilter int
	titleFilter    string
	dueFilter      dueFilterMode
	sortMode       taskSortMode

	viewMode    viewMode
	showDetails bool

	textInput textinput.Model
	textArea  textarea.Model

	keyFilter             textinput.Model
	contextFilter         textinput.Model
	contextEditInput      textinput.Model
	state                 persistedUIState
	editingDescTask       string
	pendingKanbanTaskID   string
	pendingKanbanColumnID string

	confirmingDelete bool

	statusLine string
	err        error

	width  int
	height int

	keys keyMap
}

func NewModel(taskService *application.TaskService, taskFlow *application.TaskFlow, commentService *application.CommentService, contextService *application.ContextService, setup application.BootstrapResult) Model {
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

	cf := textinput.New()
	cf.Placeholder = "Filter..."
	cf.Prompt = "Filter: "
	cf.CharLimit = 128

	ce := textinput.New()
	ce.Placeholder = "Name..."
	ce.Prompt = "> "
	ce.CharLimit = 256

	cols := make([]domain.Column, 0, len(setup.Columns))
	cols = append(cols, setup.Columns...)
	sort.Slice(cols, func(i, j int) bool {
		return cols[i].Position < cols[j].Position
	})

	state := loadPersistedUIState()

	model := Model{
		taskService:      taskService,
		taskFlow:         taskFlow,
		commentService:   commentService,
		contextService:   contextService,
		dateFormat:       detectUserDateFormat(),
		providerID:       setup.Provider.ID,
		workspaceID:      setup.Workspace.ID,
		workspaceName:    setup.Workspace.Name,
		boardID:          setup.Board.ID,
		boardName:        setup.Board.Name,
		columns:          cols,
		filterIndex:      -1,
		priorityFilter:   -1,
		dueFilter:        dueFilterAny,
		sortMode:         sortByPriority,
		viewMode:         viewList,
		showDetails:      true,
		textInput:        ti,
		textArea:         ta,
		keyFilter:        kf,
		contextFilter:    cf,
		contextEditInput: ce,
		state:            state,
		keys:             newKeyMap(),
	}

	model.bootstrapContexts()
	return model
}

func (m Model) Init() tea.Cmd {
	return m.loadTasksCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.activeOverlay() {
	case overlayTaskView:
		return m.updateTaskViewer(msg)
	case overlayKeybinds:
		return m.updateKeybindPanel(msg)
	case overlayFilters:
		return m.updateFilterPanel(msg)
	case overlayContexts:
		return m.updateContextPanel(msg)
	case overlayInput:
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
		return m.handleTasksLoaded(msg, true, true)
	case commentsLoadedMsg:
		return m.handleCommentsLoaded(msg)
	case descriptionEditedMsg:
		return m.handleExternalDescriptionEdited(msg)
	case opResultMsg:
		return m.handleOpResult(msg)
	case tea.KeyMsg:
		if m.confirmingDelete {
			if msg.String() == "y" {
				if task, ok := m.currentTask(); ok {
					m.confirmingDelete = false
					m.statusLine = ""
					return m, m.deleteTaskCmd(task.ID)
				}
			}
			m.confirmingDelete = false
			m.statusLine = ""
			return m, nil
		}
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m.executeAction("quit")
		case key.Matches(msg, m.keys.ShowKeybinds):
			m.openKeybindPanel()
			return m, textinput.Blink
		case key.Matches(msg, m.keys.ShowFilters):
			return m.executeAction("open_filters")
		case key.Matches(msg, m.keys.OpenWorkspace):
			return m.executeAction("open_workspaces")
		case key.Matches(msg, m.keys.OpenBoardPanel):
			return m.executeAction("open_board_panel")
		case key.Matches(msg, m.keys.PrevBoard):
			return m.executeAction("prev_board")
		case key.Matches(msg, m.keys.NextBoard):
			return m.executeAction("next_board")
		case key.Matches(msg, m.keys.ToggleView):
			return m.executeAction("toggle_view")
		case key.Matches(msg, m.keys.ToggleDetails):
			return m.executeAction("toggle_details")
		case key.Matches(msg, m.keys.Search):
			return m.executeAction("search")
		case key.Matches(msg, m.keys.ClearSearch):
			return m.executeAction("clear_search")
		case key.Matches(msg, m.keys.NewTask):
			return m.executeAction("new_task")
		case key.Matches(msg, m.keys.EditTitle):
			return m.executeAction("edit_task")
		case key.Matches(msg, m.keys.AddComment):
			return m.executeAction("add_comment")
		case key.Matches(msg, m.keys.EditDescription):
			return m.executeAction("edit_description")
		case key.Matches(msg, m.keys.CycleStatus):
			return m.executeAction("cycle_status")
		case key.Matches(msg, m.keys.ToggleDueSoon):
			return m.executeAction("cycle_due_filter")
		case key.Matches(msg, m.keys.CycleSort):
			return m.executeAction("cycle_sort")
		case key.Matches(msg, m.keys.Up):
			return m.executeAction("move_up")
		case key.Matches(msg, m.keys.Down):
			return m.executeAction("move_down")
		case key.Matches(msg, m.keys.KanbanMoveTaskLeft):
			return m.executeAction("move_task_left")
		case key.Matches(msg, m.keys.Left):
			return m.executeAction("move_left")
		case key.Matches(msg, m.keys.KanbanMoveTaskRight):
			return m.executeAction("move_task_right")
		case key.Matches(msg, m.keys.Right):
			return m.executeAction("move_right")
		case key.Matches(msg, m.keys.OpenDetails):
			return m.executeAction("open_move")
		case key.Matches(msg, m.keys.MoveTask):
			return m.executeAction("move_task")
		case key.Matches(msg, m.keys.MoveTaskLeft):
			return m.executeAction("move_task_left")
		case key.Matches(msg, m.keys.DeleteTask):
			if m.viewMode == viewKanban {
				if _, ok := m.currentTask(); ok {
					m.confirmingDelete = true
					m.statusLine = "delete task? y/n"
				}
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
		containerWidth := max(40, m.width-2)
		header := m.renderHeader(containerWidth)
		footer := lipgloss.NewStyle().Width(containerWidth).Render(m.renderFooter())
		bodyHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
		if bodyHeight < 5 {
			bodyHeight = 5
		}

		detailWidth := 0
		mainWidth := containerWidth
		if m.showDetails {
			detailWidth = m.width / 3
			if detailWidth < 34 {
				detailWidth = 34
			}
			mainWidth = containerWidth - detailWidth - 1
			if mainWidth < 20 {
				mainWidth = containerWidth
				detailWidth = 0
			}
		}

		mainPane := m.renderKanbanView(mainWidth, bodyHeight)
		if detailWidth > 0 {
			detailPane := m.renderDetailView(detailWidth, bodyHeight)
			mainPane = lipgloss.JoinHorizontal(lipgloss.Top, mainPane, detailPane)
		}
		mainPane = lipgloss.NewStyle().Width(containerWidth).Height(bodyHeight).Render(mainPane)

		content := lipgloss.JoinVertical(lipgloss.Left, header, mainPane, footer)
		base = lipgloss.NewStyle().Padding(0, 1).Render(content)
	}

	if m.showKeybinds {
		return m.renderKeybindPanel(base)
	}
	if m.showFilters {
		return m.renderFilterPanel(base)
	}
	if m.showContexts {
		return m.renderContextPanel(base)
	}
	base = m.renderTaskFormOverlay(base)
	if m.showTaskView {
		return m.renderTaskViewerPanel(base)
	}
	return base
}

func (m Model) renderHeader(width int) string {
	viewLabel := "List"
	if m.viewMode == viewKanban {
		viewLabel = "Kanban"
	}
	filterParts := []string{fmt.Sprintf("status:%s", m.statusFilterLabel()), fmt.Sprintf("due:%s", strings.ToLower(m.dueFilterLabel()))}
	if m.priorityFilter >= 0 {
		filterParts = append(filterParts, fmt.Sprintf("priority:p%d", m.priorityFilter))
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))

	left := headerStyle.Render(fmt.Sprintf("%s / %s", m.workspaceName, m.boardName))
	right := metaStyle.Render(fmt.Sprintf("view:%s  sort:%s  filter:%s  search:%q", viewLabel, strings.ToLower(m.sortModeLabel()), strings.Join(filterParts, ","), m.titleFilter))
	if width > 20 {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(width/2).Render(left),
			lipgloss.NewStyle().Width(max(1, width-width/2-1)).Align(lipgloss.Right).Render(right),
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

	shortcuts := "?:help  n:new  /:search  enter:open  w:workspaces  b:boards  f:filters  q:quit"
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

// toSelectionState creates a selectionState from the Model's current selection-related fields.
func (m Model) toSelectionState() selectionState {
	return selectionState{
		viewMode:              m.viewMode,
		columns:               m.columns,
		tasks:                 m.tasks,
		selected:              m.selected,
		activeColumn:          m.activeColumn,
		kanbanRow:             m.kanbanRow,
		pendingKanbanTaskID:   m.pendingKanbanTaskID,
		pendingKanbanColumnID: m.pendingKanbanColumnID,
	}
}

// applySelectionState copies the selection-related fields from selectionState back to Model.
func (m *Model) applySelectionState(s selectionState) {
	m.selected = s.selected
	m.activeColumn = s.activeColumn
	m.kanbanRow = s.kanbanRow
	m.pendingKanbanTaskID = s.pendingKanbanTaskID
	m.pendingKanbanColumnID = s.pendingKanbanColumnID
}

func (m *Model) ensureSelection() {
	s := m.toSelectionState()
	s.ensureSelection()
	m.applySelectionState(s)
}

func (m *Model) setActiveColumnByID(columnID string) {
	s := m.toSelectionState()
	s.setActiveColumnByID(columnID)
	m.applySelectionState(s)
}

func (m *Model) restorePendingKanbanSelection() bool {
	s := m.toSelectionState()
	ok := s.restorePendingKanbanSelection()
	m.applySelectionState(s)
	return ok
}

func (m *Model) ensureKanbanRow() {
	s := m.toSelectionState()
	s.ensureKanbanRow()
	m.applySelectionState(s)
}

func (m *Model) moveUp() {
	s := m.toSelectionState()
	s.moveUp()
	m.applySelectionState(s)
}

func (m *Model) moveDown() {
	s := m.toSelectionState()
	s.moveDown()
	m.applySelectionState(s)
}

func (m Model) currentTask() (domain.Task, bool) {
	s := m.toSelectionState()
	return s.currentTask()
}

func (m Model) tasksForColumn(columnID string) []domain.Task {
	s := m.toSelectionState()
	return s.tasksForColumn(columnID)
}

func (m Model) loadTasksCmd() tea.Cmd {
	filters := application.ListTaskFilters{
		WorkspaceID: m.workspaceID,
		BoardID:     m.boardID,
		TitleQuery:  m.titleFilter,
		ColumnID:    m.columnFilter,
	}
	flow := m.taskFlow
	return func() tea.Msg {
		tasks, err := flow.ListTasks(context.Background(), filters)
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

func (m Model) applyActiveFilters(tasks []domain.Task) []domain.Task {
	if len(tasks) == 0 {
		return tasks
	}
	now := time.Now().UTC()
	soonLimit := now.AddDate(0, 0, 7)
	filtered := make([]domain.Task, 0, len(tasks))
	for _, task := range tasks {
		if m.columnFilter != "" {
			if task.ColumnID == nil || *task.ColumnID != m.columnFilter {
				continue
			}
		}
		if m.priorityFilter >= 0 && normalizePriority(task.Priority) != m.priorityFilter {
			continue
		}
		switch m.dueFilter {
		case dueFilterSoon:
			if task.DueAt == nil {
				continue
			}
			due := task.DueAt.UTC()
			if due.Before(now) || due.After(soonLimit) {
				continue
			}
		case dueFilterOverdue:
			if task.DueAt == nil || !task.DueAt.UTC().Before(now) {
				continue
			}
		case dueFilterNoDate:
			if task.DueAt != nil {
				continue
			}
		}
		filtered = append(filtered, task)
	}
	return filtered
}

func (m *Model) sortTasks(tasks []domain.Task) {
	switch m.sortMode {
	case sortByDueDate:
		sort.SliceStable(tasks, func(i, j int) bool {
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
	case sortByTitle:
		sort.SliceStable(tasks, func(i, j int) bool {
			ti := strings.ToLower(strings.TrimSpace(tasks[i].Title))
			tj := strings.ToLower(strings.TrimSpace(tasks[j].Title))
			if ti != tj {
				return ti < tj
			}
			return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
		})
	case sortByUpdated:
		sort.SliceStable(tasks, func(i, j int) bool {
			return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
		})
	case sortByCreated:
		sort.SliceStable(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		})
	default:
		m.sortTasksByPriority(tasks)
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
