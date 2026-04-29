package ui

import (
	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

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
	if model, cmd, ok := m.dispatchOverlayUpdate(msg); ok {
		return model, cmd
	}
	if model, cmd, ok := m.dispatchGlobalMessage(msg); ok {
		return model, cmd
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		if model, cmd, ok := m.handleDeleteConfirmKey(msg); ok {
			return model, cmd
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
	return m.wrapOverlays(m.renderBaseView())
}

func (m Model) applyActiveFilters(tasks []domain.Task) []domain.Task {
	return m.toTaskFilterState().applyActiveFilters(tasks)
}

func (m *Model) sortTasks(tasks []domain.Task) {
	m.toTaskFilterState().sortTasks(tasks)
}

func (m *Model) sortTasksByPriority(tasks []domain.Task) {
	m.toTaskFilterState().sortTasksByPriority(tasks)
}

func (m Model) toTaskFilterState() taskFilterState {
	return taskFilterState{
		columnFilter:   m.columnFilter,
		priorityFilter: m.priorityFilter,
		dueFilter:      m.dueFilter,
		sortMode:       m.sortMode,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
