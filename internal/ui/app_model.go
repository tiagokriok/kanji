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

type contextMode int

const (
	contextWorkspace contextMode = iota
	contextBoard
)

type contextEditMode int

const (
	contextEditNone contextEditMode = iota
	contextEditCreate
	contextEditRename
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

var boardColorPickerPalette = []string{
	"#60A5FA",
	"#F59E0B",
	"#22C55E",
	"#A78BFA",
	"#F472B6",
	"#38BDF8",
	"#F87171",
	"#34D399",
	"#FACC15",
	"#9CA3AF",
}

var taskPriorityOptions = []taskPriorityOption{
	{Value: 0, Label: "Critical"},
	{Value: 1, Label: "Urgent"},
	{Value: 2, Label: "High"},
	{Value: 3, Label: "Medium"},
	{Value: 4, Label: "Low"},
	{Value: 5, Label: "None"},
}

type boardColumnFormRow struct {
	name  textinput.Model
	color textinput.Model
}

type boardCreateForm struct {
	focus     int
	boardName textinput.Model
	columns   []boardColumnFormRow
}

type boardColumnsOrderForm struct {
	boardID   string
	boardName string
	columns   []domain.Column
	selected  int
}

func (f *boardColumnsOrderForm) clampSelection() {
	if len(f.columns) == 0 {
		f.selected = 0
		return
	}
	if f.selected < 0 {
		f.selected = 0
	}
	if f.selected >= len(f.columns) {
		f.selected = len(f.columns) - 1
	}
}

func (f *boardColumnsOrderForm) moveSelection(delta int) {
	f.selected += delta
	f.clampSelection()
}

func (f *boardColumnsOrderForm) moveSelected(delta int) bool {
	if len(f.columns) < 2 {
		return false
	}
	f.clampSelection()
	target := f.selected + delta
	if target < 0 || target >= len(f.columns) {
		return false
	}
	f.columns[f.selected], f.columns[target] = f.columns[target], f.columns[f.selected]
	f.selected = target
	for i := range f.columns {
		f.columns[i].Position = i + 1
	}
	return true
}

func (f *boardColumnsOrderForm) orderedColumnIDs() []string {
	ids := make([]string, 0, len(f.columns))
	for _, col := range f.columns {
		ids = append(ids, col.ID)
	}
	return ids
}

func newBoardColumnRow(index int, name, color string) boardColumnFormRow {
	if index < 1 {
		index = 1
	}
	if strings.TrimSpace(color) == "" {
		if len(boardColorPickerPalette) > 0 {
			color = boardColorPickerPalette[(index-1)%len(boardColorPickerPalette)]
		} else {
			color = "#6B7280"
		}
	}
	return boardColumnFormRow{
		name:  newTaskFormInput(fmt.Sprintf("Column %d name", index), name, 64),
		color: newTaskFormInput(fmt.Sprintf("Column %d color (#RRGGBB)", index), strings.ToUpper(strings.TrimSpace(color)), 7),
	}
}

func (f *boardCreateForm) fields() []*textinput.Model {
	fields := make([]*textinput.Model, 0, 1+(len(f.columns)*2))
	fields = append(fields, &f.boardName)
	for i := range f.columns {
		fields = append(fields, &f.columns[i].name, &f.columns[i].color)
	}
	return fields
}

func (f *boardCreateForm) currentField() *textinput.Model {
	fields := f.fields()
	if f.focus < 0 {
		f.focus = 0
	}
	if f.focus >= len(fields) {
		f.focus = len(fields) - 1
	}
	return fields[f.focus]
}

func (f *boardCreateForm) setFocus(index int) {
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

func (f *boardCreateForm) moveFocus(delta int) {
	f.setFocus(f.focus + delta)
}

func (f *boardCreateForm) focusedColumnIndex() (int, bool) {
	if f.focus <= 0 {
		return 0, false
	}
	index := (f.focus - 1) / 2
	if index < 0 || index >= len(f.columns) {
		return 0, false
	}
	return index, true
}

func (f *boardCreateForm) focusedColorColumnIndex() (int, bool) {
	columnIndex, ok := f.focusedColumnIndex()
	if !ok {
		return 0, false
	}
	index := f.focus - 1
	if index%2 == 1 && columnIndex >= 0 && columnIndex < len(f.columns) {
		return columnIndex, true
	}
	return 0, false
}

func (f *boardCreateForm) cycleFocusedColor(delta int) bool {
	columnIndex, ok := f.focusedColorColumnIndex()
	if !ok || columnIndex < 0 || columnIndex >= len(f.columns) || len(boardColorPickerPalette) == 0 {
		return false
	}

	current := strings.ToUpper(strings.TrimSpace(f.columns[columnIndex].color.Value()))
	paletteIndex := -1
	for i, c := range boardColorPickerPalette {
		if strings.EqualFold(c, current) {
			paletteIndex = i
			break
		}
	}
	if paletteIndex < 0 {
		if delta < 0 {
			paletteIndex = len(boardColorPickerPalette) - 1
		} else {
			paletteIndex = 0
		}
	} else {
		paletteIndex = (paletteIndex + delta + len(boardColorPickerPalette)) % len(boardColorPickerPalette)
	}

	f.columns[columnIndex].color.SetValue(boardColorPickerPalette[paletteIndex])
	return true
}

func (f *boardCreateForm) addColumn() {
	index := len(f.columns) + 1
	f.columns = append(f.columns, newBoardColumnRow(index, "", ""))
	f.setFocus(1 + (len(f.columns)-1)*2)
}

func (f *boardCreateForm) removeFocusedColumn() bool {
	columnIndex, ok := f.focusedColumnIndex()
	if !ok || len(f.columns) <= 1 {
		return false
	}

	f.columns = append(f.columns[:columnIndex], f.columns[columnIndex+1:]...)
	nextFocus := 1 + columnIndex*2
	if nextFocus >= len(f.fields()) {
		nextFocus = len(f.fields()) - 1
	}
	f.setFocus(nextFocus)
	return true
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

	viewMode     viewMode
	showDetails  bool
	showTaskView bool
	inputMode    inputMode
	taskForm     *taskForm
	showKeybinds bool
	showFilters  bool
	showContexts bool
	contextMode  contextMode
	boardForm    *boardCreateForm
	boardOrder   *boardColumnsOrderForm

	textInput textinput.Model
	textArea  textarea.Model

	keyFilter        textinput.Model
	keySelected      int
	filterFocus      int
	contextFilter    textinput.Model
	contextSelected  int
	contextEditMode  contextEditMode
	contextEditInput textinput.Model
	state            persistedUIState
	editingDescTask  string
	viewTaskID       string
	viewDescScroll   int
	returnTaskView   bool
	returnTaskID     string

	statusLine string
	err        error

	width  int
	height int

	keys keyMap
}

func NewModel(taskService *application.TaskService, commentService *application.CommentService, contextService *application.ContextService, setup application.BootstrapResult) Model {
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
	if m.showTaskView {
		return m.updateTaskViewer(msg)
	}
	if m.showKeybinds {
		return m.updateKeybindPanel(msg)
	}
	if m.showFilters {
		return m.updateFilterPanel(msg)
	}
	if m.showContexts {
		return m.updateContextPanel(msg)
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
		m.tasks = m.applyActiveFilters(msg.tasks)
		m.sortTasks(m.tasks)
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
	case descriptionEditedMsg:
		if m.editingDescTask == "" {
			return m, nil
		}
		taskID := m.editingDescTask
		m.editingDescTask = ""
		if msg.err != nil {
			m.err = msg.err
			m.statusLine = fmt.Sprintf("editor error: %v", msg.err)
			return m, nil
		}
		return m, m.updateTaskDescriptionCmd(taskID, msg.content)
	case opResultMsg:
		if msg.err != nil {
			m.err = msg.err
			m.statusLine = msg.err.Error()
			m.clearTaskViewerReturn()
			return m, nil
		}
		m.statusLine = ""
		if m.returnTaskView && strings.TrimSpace(m.returnTaskID) != "" {
			taskID := m.returnTaskID
			m.clearTaskViewerReturn()
			commentsCmd := m.openTaskViewerByID(taskID)
			return m, tea.Batch(m.loadTasksCmd(), commentsCmd)
		}
		return m, m.loadTasksCmd()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.ShowKeybinds):
			m.openKeybindPanel()
			return m, textinput.Blink
		case key.Matches(msg, m.keys.ShowFilters):
			m.openFilterPanel()
			return m, nil
		case key.Matches(msg, m.keys.OpenWorkspace):
			m.openContextPanel(contextWorkspace)
			return m, textinput.Blink
		case key.Matches(msg, m.keys.OpenBoardPanel):
			m.openContextPanel(contextBoard)
			return m, textinput.Blink
		case key.Matches(msg, m.keys.PrevBoard):
			changed, err := m.switchBoardByOffset(-1)
			if err != nil {
				m.statusLine = err.Error()
				return m, nil
			}
			if changed {
				return m, m.loadTasksCmd()
			}
			return m, nil
		case key.Matches(msg, m.keys.NextBoard):
			changed, err := m.switchBoardByOffset(1)
			if err != nil {
				m.statusLine = err.Error()
				return m, nil
			}
			if changed {
				return m, m.loadTasksCmd()
			}
			return m, nil
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
			return m, m.startExternalDescriptionEdit(task)
		case key.Matches(msg, m.keys.CycleStatus):
			m.cycleColumnFilter()
			return m, m.loadTasksCmd()
		case key.Matches(msg, m.keys.ToggleDueSoon):
			m.cycleDueFilter()
			return m, m.loadTasksCmd()
		case key.Matches(msg, m.keys.CycleSort):
			m.cycleSortMode()
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
			return m, m.openTaskViewer()
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
	if m.showFilters {
		return m.renderFilterPanel(base)
	}
	if m.showContexts {
		return m.renderContextPanel(base)
	}
	if m.inputMode == inputTaskForm && m.taskForm != nil {
		return m.renderTaskFormPanel(base)
	}
	if m.showTaskView {
		return m.renderTaskViewerPanel(base)
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
				if m.returnTaskView && strings.TrimSpace(m.returnTaskID) != "" {
					taskID := m.returnTaskID
					m.clearTaskViewerReturn()
					return m, m.openTaskViewerByID(taskID)
				}
				return m, nil
			}
			m.inputMode = inputNone
			m.textInput.Blur()
			m.textArea.Blur()
			m.statusLine = ""
			if mode == inputAddComment && m.returnTaskView && strings.TrimSpace(m.returnTaskID) != "" {
				taskID := m.returnTaskID
				m.clearTaskViewerReturn()
				return m, m.openTaskViewerByID(taskID)
			}
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
					m.inputMode = inputAddComment
					m.textInput.Focus()
					m.statusLine = "comment is required"
					return m, textinput.Blink
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
		if field := m.taskForm.currentInputField(); field != nil {
			prevDescription := m.taskForm.description.Value()
			var cmd tea.Cmd
			*field, cmd = field.Update(msg)
			if m.taskForm.focus == taskFieldDescription && m.taskForm.description.Value() != prevDescription {
				m.taskForm.descriptionFull = m.taskForm.description.Value()
			}
			return m, cmd
		}
		return m, nil
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
	statusOptions, statusIndex := m.buildTaskStatusOptions(nil)

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

	m.taskForm = form
	m.inputMode = inputTaskForm
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

	m.taskForm = form
	m.inputMode = inputTaskForm
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

func (m *Model) startExternalDescriptionEdit(task domain.Task) tea.Cmd {
	m.editingDescTask = task.ID
	m.statusLine = ""
	return openDescriptionEditorCmd(task.DescriptionMD)
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

	tmpFile, err := os.CreateTemp("", "kanji-description-*.md")
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

func (m *Model) openTaskViewer() tea.Cmd {
	task, ok := m.currentTask()
	if !ok {
		return nil
	}
	return m.openTaskViewerByID(task.ID)
}

func (m *Model) openTaskViewerByID(taskID string) tea.Cmd {
	if strings.TrimSpace(taskID) == "" {
		return nil
	}
	m.showTaskView = true
	m.viewTaskID = taskID
	m.viewDescScroll = 0
	m.comments = nil
	return m.loadCommentsCmd(taskID)
}

func (m *Model) closeTaskViewer() {
	m.showTaskView = false
	m.viewTaskID = ""
	m.viewDescScroll = 0
}

func (m *Model) setTaskViewerReturn(taskID string) {
	m.returnTaskView = true
	m.returnTaskID = strings.TrimSpace(taskID)
}

func (m *Model) clearTaskViewerReturn() {
	m.returnTaskView = false
	m.returnTaskID = ""
}

func (m Model) viewerTask() (domain.Task, bool) {
	if strings.TrimSpace(m.viewTaskID) != "" {
		for _, task := range m.tasks {
			if task.ID == m.viewTaskID {
				return task, true
			}
		}
	}
	return m.currentTask()
}

func (m Model) updateTaskViewer(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tasksLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.statusLine = msg.err.Error()
			return m, nil
		}
		m.tasks = m.applyActiveFilters(msg.tasks)
		m.sortTasks(m.tasks)
		m.ensureSelection()
		return m, nil
	case commentsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.statusLine = msg.err.Error()
			return m, nil
		}
		m.comments = msg.comments
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel), key.Matches(msg, m.keys.Confirm), key.Matches(msg, m.keys.OpenDetails):
			m.closeTaskViewer()
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			m.closeTaskViewer()
			return m, nil
		case key.Matches(msg, m.keys.EditTitle):
			task, ok := m.viewerTask()
			if !ok {
				return m, nil
			}
			m.setTaskViewerReturn(task.ID)
			m.closeTaskViewer()
			m.startEditTaskForm(task)
			return m, textinput.Blink
		case key.Matches(msg, m.keys.AddComment):
			task, ok := m.viewerTask()
			if !ok {
				return m, nil
			}
			m.setTaskViewerReturn(task.ID)
			m.closeTaskViewer()
			m.inputMode = inputAddComment
			m.textInput.SetValue("")
			m.textInput.Placeholder = "Comment body"
			m.textInput.Focus()
			m.statusLine = "Add comment"
			return m, textinput.Blink
		case key.Matches(msg, m.keys.Up):
			if m.viewDescScroll > 0 {
				m.viewDescScroll--
			}
			return m, nil
		case key.Matches(msg, m.keys.Down):
			if maxScroll := m.taskViewerMaxDescScroll(); m.viewDescScroll < maxScroll {
				m.viewDescScroll++
			}
			return m, nil
		}

		switch msg.String() {
		case "pgup":
			if m.viewDescScroll > 0 {
				step := max(1, m.taskViewerDescViewportLines()/2)
				m.viewDescScroll -= step
				if m.viewDescScroll < 0 {
					m.viewDescScroll = 0
				}
			}
			return m, nil
		case "pgdown":
			if maxScroll := m.taskViewerMaxDescScroll(); m.viewDescScroll < maxScroll {
				step := max(1, m.taskViewerDescViewportLines()/2)
				m.viewDescScroll += step
				if m.viewDescScroll > maxScroll {
					m.viewDescScroll = maxScroll
				}
			}
			return m, nil
		case "home":
			m.viewDescScroll = 0
			return m, nil
		case "end":
			m.viewDescScroll = m.taskViewerMaxDescScroll()
			return m, nil
		}
	}
	return m, nil
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

func (m *Model) openFilterPanel() {
	m.showFilters = true
	m.filterFocus = 0
}

func (m *Model) closeFilterPanel() {
	m.showFilters = false
}

func (m *Model) bootstrapContexts() {
	if m.contextService == nil {
		return
	}
	_ = m.reloadContextsFromStorage()
}

func (m *Model) persistContextSelection() {
	m.state.LastWorkspaceID = m.workspaceID
	if m.state.LastBoardByWorkspace == nil {
		m.state.LastBoardByWorkspace = map[string]string{}
	}
	if m.workspaceID != "" && m.boardID != "" {
		m.state.LastBoardByWorkspace[m.workspaceID] = m.boardID
	}
	_ = savePersistedUIState(m.state)
}

func (m *Model) reloadContextsFromStorage() error {
	if m.contextService == nil {
		return nil
	}
	ctx := context.Background()
	workspaces, err := m.contextService.ListWorkspaces(ctx)
	if err != nil {
		return err
	}
	m.workspaces = workspaces
	if len(workspaces) == 0 {
		return nil
	}

	targetWorkspaceID := m.workspaceID
	if m.state.LastWorkspaceID != "" && containsWorkspace(workspaces, m.state.LastWorkspaceID) {
		targetWorkspaceID = m.state.LastWorkspaceID
	}
	if targetWorkspaceID == "" || !containsWorkspace(workspaces, targetWorkspaceID) {
		targetWorkspaceID = workspaces[0].ID
	}
	if err := m.switchWorkspace(targetWorkspaceID); err != nil {
		return err
	}
	return nil
}

func containsWorkspace(items []domain.Workspace, workspaceID string) bool {
	for _, item := range items {
		if item.ID == workspaceID {
			return true
		}
	}
	return false
}

func containsBoard(items []domain.Board, boardID string) bool {
	for _, item := range items {
		if item.ID == boardID {
			return true
		}
	}
	return false
}

func workspaceName(items []domain.Workspace, workspaceID string) string {
	for _, item := range items {
		if item.ID == workspaceID {
			return item.Name
		}
	}
	return ""
}

func boardName(items []domain.Board, boardID string) string {
	for _, item := range items {
		if item.ID == boardID {
			return item.Name
		}
	}
	return ""
}

func (m *Model) switchWorkspace(workspaceID string) error {
	if m.contextService == nil {
		return nil
	}
	ctx := context.Background()
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return fmt.Errorf("workspace id is required")
	}

	if len(m.workspaces) == 0 {
		workspaces, err := m.contextService.ListWorkspaces(ctx)
		if err != nil {
			return err
		}
		m.workspaces = workspaces
	}
	if !containsWorkspace(m.workspaces, workspaceID) {
		return fmt.Errorf("workspace not found")
	}

	m.workspaceID = workspaceID
	m.workspaceName = workspaceName(m.workspaces, workspaceID)
	m.columnFilter = ""
	m.filterIndex = -1

	boards, err := m.contextService.ListBoards(ctx, workspaceID)
	if err != nil {
		return err
	}
	m.boards = boards
	if len(boards) == 0 {
		board, createErr := m.contextService.CreateBoard(ctx, workspaceID, "Main")
		if createErr != nil {
			return createErr
		}
		m.boards = []domain.Board{board}
	}

	targetBoardID := ""
	if saved, ok := m.state.LastBoardByWorkspace[workspaceID]; ok && containsBoard(m.boards, saved) {
		targetBoardID = saved
	}
	if targetBoardID == "" && containsBoard(m.boards, m.boardID) {
		targetBoardID = m.boardID
	}
	if targetBoardID == "" {
		targetBoardID = m.boards[0].ID
	}
	if err := m.switchBoard(targetBoardID); err != nil {
		return err
	}
	m.persistContextSelection()
	return nil
}

func (m *Model) switchBoard(boardID string) error {
	if m.contextService == nil {
		return nil
	}
	ctx := context.Background()
	boardID = strings.TrimSpace(boardID)
	if boardID == "" {
		return fmt.Errorf("board id is required")
	}

	if len(m.boards) == 0 {
		boards, err := m.contextService.ListBoards(ctx, m.workspaceID)
		if err != nil {
			return err
		}
		m.boards = boards
	}
	if !containsBoard(m.boards, boardID) {
		return fmt.Errorf("board not found")
	}

	m.boardID = boardID
	m.boardName = boardName(m.boards, boardID)
	m.columnFilter = ""
	m.filterIndex = -1

	columns, err := m.contextService.ListColumns(ctx, boardID)
	if err != nil {
		return err
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Position < columns[j].Position
	})
	m.columns = columns
	m.persistContextSelection()
	return nil
}

func (m *Model) openContextPanel(mode contextMode) {
	m.showContexts = true
	m.contextMode = mode
	m.contextSelected = 0
	m.contextEditMode = contextEditNone
	m.boardForm = nil
	m.boardOrder = nil
	m.contextEditInput.SetValue("")
	m.contextFilter.SetValue("")
	m.contextFilter.Focus()
}

func (m *Model) closeContextPanel() {
	m.showContexts = false
	m.contextEditMode = contextEditNone
	m.boardForm = nil
	m.boardOrder = nil
	m.contextEditInput.Blur()
	m.contextFilter.Blur()
}

func (m Model) contextItems() []string {
	query := strings.ToLower(strings.TrimSpace(m.contextFilter.Value()))
	items := make([]string, 0)
	if m.contextMode == contextWorkspace {
		for _, ws := range m.workspaces {
			if query == "" || strings.Contains(strings.ToLower(ws.Name), query) {
				items = append(items, ws.ID)
			}
		}
	} else {
		for _, b := range m.boards {
			if query == "" || strings.Contains(strings.ToLower(b.Name), query) {
				items = append(items, b.ID)
			}
		}
	}
	return items
}

func (m *Model) clampContextSelection() {
	items := m.contextItems()
	if len(items) == 0 {
		m.contextSelected = 0
		return
	}
	if m.contextSelected < 0 {
		m.contextSelected = 0
	}
	if m.contextSelected >= len(items) {
		m.contextSelected = len(items) - 1
	}
}

func (m *Model) selectedContextID() string {
	items := m.contextItems()
	if len(items) == 0 {
		return ""
	}
	m.clampContextSelection()
	return items[m.contextSelected]
}

func (m Model) contextTitle() string {
	if m.contextMode == contextWorkspace {
		return "Workspaces"
	}
	return "Boards"
}

func (m Model) contextNameByID(id string) string {
	if m.contextMode == contextWorkspace {
		return workspaceName(m.workspaces, id)
	}
	return boardName(m.boards, id)
}

func (m *Model) beginContextCreate() {
	if m.contextMode == contextBoard {
		m.startBoardCreateForm()
		return
	}
	m.contextEditMode = contextEditCreate
	m.contextEditInput.SetValue("")
	m.contextEditInput.Placeholder = "New name"
	m.contextEditInput.Focus()
}

func (m *Model) startBoardCreateForm() {
	form := &boardCreateForm{
		boardName: newTaskFormInput("Board name", "", 128),
		columns: []boardColumnFormRow{
			newBoardColumnRow(1, "Todo", "#60A5FA"),
			newBoardColumnRow(2, "Doing", "#F59E0B"),
			newBoardColumnRow(3, "Done", "#22C55E"),
		},
	}
	form.setFocus(0)
	m.boardForm = form
	m.contextEditMode = contextEditNone
	m.contextEditInput.Blur()
	m.contextFilter.Blur()
}

func (m *Model) closeBoardCreateForm() {
	m.boardForm = nil
	m.contextFilter.Focus()
}

func (m *Model) beginBoardColumnsReorder() error {
	if m.contextService == nil {
		return nil
	}
	if m.contextMode != contextBoard {
		return nil
	}

	boardID := m.selectedContextID()
	if boardID == "" {
		return fmt.Errorf("no board selected")
	}

	columns, err := m.contextService.ListColumns(context.Background(), boardID)
	if err != nil {
		return err
	}
	if len(columns) == 0 {
		return fmt.Errorf("selected board has no columns")
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Position < columns[j].Position
	})

	m.boardOrder = &boardColumnsOrderForm{
		boardID:   boardID,
		boardName: m.contextNameByID(boardID),
		columns:   columns,
		selected:  0,
	}
	m.boardForm = nil
	m.contextEditMode = contextEditNone
	m.contextEditInput.Blur()
	m.contextFilter.Blur()
	return nil
}

func (m *Model) closeBoardColumnsReorder() {
	m.boardOrder = nil
	m.contextFilter.Focus()
}

func (m *Model) submitBoardColumnsReorder() error {
	if m.contextService == nil || m.boardOrder == nil {
		return nil
	}
	form := m.boardOrder
	if err := m.contextService.ReorderColumns(context.Background(), form.boardID, form.orderedColumnIDs()); err != nil {
		return err
	}
	if form.boardID == m.boardID {
		if err := m.switchBoard(m.boardID); err != nil {
			return err
		}
	}
	m.closeBoardColumnsReorder()
	return nil
}

func (m *Model) submitBoardCreateForm() error {
	if m.contextService == nil {
		return nil
	}
	if m.boardForm == nil {
		return fmt.Errorf("board form is not active")
	}

	boardName := strings.TrimSpace(m.boardForm.boardName.Value())
	if boardName == "" {
		return fmt.Errorf("board name is required")
	}

	columns := make([]application.CreateBoardColumnInput, 0, len(m.boardForm.columns))
	for i, row := range m.boardForm.columns {
		name := strings.TrimSpace(row.name.Value())
		color := strings.ToUpper(strings.TrimSpace(row.color.Value()))
		if name == "" && color == "" {
			continue
		}
		if name == "" {
			return fmt.Errorf("column %d name is required", i+1)
		}
		if !uiHexColorPattern.MatchString(color) {
			return fmt.Errorf("column %d color must be HEX (#RRGGBB)", i+1)
		}
		columns = append(columns, application.CreateBoardColumnInput{Name: name, Color: color})
	}
	if len(columns) == 0 {
		return fmt.Errorf("add at least one column")
	}

	ctx := context.Background()
	board, err := m.contextService.CreateBoardWithColumns(ctx, m.workspaceID, boardName, columns)
	if err != nil {
		return err
	}

	boards, err := m.contextService.ListBoards(ctx, m.workspaceID)
	if err != nil {
		return err
	}
	m.boards = boards
	if err := m.switchBoard(board.ID); err != nil {
		return err
	}
	m.closeBoardCreateForm()
	return nil
}

func (m *Model) beginContextRename() {
	id := m.selectedContextID()
	if id == "" {
		return
	}
	m.contextEditMode = contextEditRename
	m.contextEditInput.SetValue(m.contextNameByID(id))
	m.contextEditInput.Placeholder = "Rename"
	m.contextEditInput.Focus()
}

func (m *Model) submitContextEdit() error {
	if m.contextService == nil {
		return nil
	}
	value := strings.TrimSpace(m.contextEditInput.Value())
	if value == "" {
		return fmt.Errorf("name is required")
	}
	ctx := context.Background()

	if m.contextEditMode == contextEditCreate {
		if m.contextMode == contextWorkspace {
			workspace, board, err := m.contextService.CreateWorkspace(ctx, m.providerID, value)
			if err != nil {
				return err
			}
			_ = workspace
			if err := m.reloadContextsFromStorage(); err != nil {
				return err
			}
			if err := m.switchWorkspace(board.WorkspaceID); err != nil {
				return err
			}
			if err := m.switchBoard(board.ID); err != nil {
				return err
			}
		} else {
			board, err := m.contextService.CreateBoard(ctx, m.workspaceID, value)
			if err != nil {
				return err
			}
			boards, listErr := m.contextService.ListBoards(ctx, m.workspaceID)
			if listErr != nil {
				return listErr
			}
			m.boards = boards
			if err := m.switchBoard(board.ID); err != nil {
				return err
			}
		}
	} else if m.contextEditMode == contextEditRename {
		id := m.selectedContextID()
		if id == "" {
			return fmt.Errorf("no item selected")
		}
		if m.contextMode == contextWorkspace {
			if err := m.contextService.RenameWorkspace(ctx, id, value); err != nil {
				return err
			}
			workspaces, err := m.contextService.ListWorkspaces(ctx)
			if err != nil {
				return err
			}
			m.workspaces = workspaces
			m.workspaceName = workspaceName(workspaces, m.workspaceID)
		} else {
			if err := m.contextService.RenameBoard(ctx, id, value); err != nil {
				return err
			}
			boards, err := m.contextService.ListBoards(ctx, m.workspaceID)
			if err != nil {
				return err
			}
			m.boards = boards
			m.boardName = boardName(boards, m.boardID)
		}
	}
	m.contextEditMode = contextEditNone
	m.contextEditInput.Blur()
	m.contextFilter.Focus()
	return nil
}

func (m Model) updateContextPanel(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.statusLine = msg.err.Error()
			return m, nil
		}
		m.tasks = m.applyActiveFilters(msg.tasks)
		m.sortTasks(m.tasks)
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.boardOrder != nil {
			switch {
			case key.Matches(msg, m.keys.Cancel):
				m.closeBoardColumnsReorder()
				return m, nil
			case key.Matches(msg, m.keys.Confirm):
				if err := m.submitBoardColumnsReorder(); err != nil {
					m.statusLine = err.Error()
					return m, nil
				}
				return m, m.loadTasksCmd()
			case key.Matches(msg, m.keys.Up):
				m.boardOrder.moveSelection(-1)
				return m, nil
			case key.Matches(msg, m.keys.Down):
				m.boardOrder.moveSelection(1)
				return m, nil
			}
			switch msg.String() {
			case "ctrl+k":
				if m.boardOrder.moveSelected(-1) {
					m.statusLine = ""
				}
				return m, nil
			case "ctrl+j":
				if m.boardOrder.moveSelected(1) {
					m.statusLine = ""
				}
				return m, nil
			}
			return m, nil
		}

		if m.boardForm != nil {
			switch msg.String() {
			case "tab", "down":
				m.boardForm.moveFocus(1)
				return m, textinput.Blink
			case "shift+tab", "up":
				m.boardForm.moveFocus(-1)
				return m, textinput.Blink
			case "ctrl+n":
				m.boardForm.addColumn()
				return m, textinput.Blink
			case "ctrl+d":
				if m.boardForm.removeFocusedColumn() {
					return m, textinput.Blink
				}
				return m, nil
			case "left":
				if m.boardForm.cycleFocusedColor(-1) {
					return m, nil
				}
			case "right":
				if m.boardForm.cycleFocusedColor(1) {
					return m, nil
				}
			case "enter":
				if m.boardForm.focus == len(m.boardForm.fields())-1 {
					if err := m.submitBoardCreateForm(); err != nil {
						m.statusLine = err.Error()
						return m, nil
					}
					return m, m.loadTasksCmd()
				}
				m.boardForm.moveFocus(1)
				return m, textinput.Blink
			}

			switch {
			case key.Matches(msg, m.keys.Cancel):
				m.closeBoardCreateForm()
				return m, nil
			case key.Matches(msg, m.keys.Confirm):
				if err := m.submitBoardCreateForm(); err != nil {
					m.statusLine = err.Error()
					return m, nil
				}
				return m, m.loadTasksCmd()
			}

			field := m.boardForm.currentField()
			var cmd tea.Cmd
			*field, cmd = field.Update(msg)
			if _, ok := m.boardForm.focusedColorColumnIndex(); ok {
				field.SetValue(strings.ToUpper(field.Value()))
			}
			return m, cmd
		}

		if m.contextEditMode != contextEditNone {
			switch {
			case key.Matches(msg, m.keys.Cancel):
				m.contextEditMode = contextEditNone
				m.contextEditInput.Blur()
				m.contextFilter.Focus()
				return m, nil
			case key.Matches(msg, m.keys.Confirm):
				if err := m.submitContextEdit(); err != nil {
					m.statusLine = err.Error()
					return m, nil
				}
				return m, m.loadTasksCmd()
			}
			var cmd tea.Cmd
			m.contextEditInput, cmd = m.contextEditInput.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keys.Cancel):
			m.closeContextPanel()
			return m, nil
		case key.Matches(msg, m.keys.OpenWorkspace):
			m.openContextPanel(contextWorkspace)
			return m, textinput.Blink
		case key.Matches(msg, m.keys.OpenBoardPanel):
			m.openContextPanel(contextBoard)
			return m, textinput.Blink
		case key.Matches(msg, m.keys.Up):
			m.contextSelected--
			m.clampContextSelection()
			return m, nil
		case key.Matches(msg, m.keys.Down):
			m.contextSelected++
			m.clampContextSelection()
			return m, nil
		case msg.String() == "n":
			m.beginContextCreate()
			return m, textinput.Blink
		case msg.String() == "r":
			m.beginContextRename()
			return m, textinput.Blink
		case msg.String() == "o" && m.contextMode == contextBoard:
			if err := m.beginBoardColumnsReorder(); err != nil {
				m.statusLine = err.Error()
				return m, nil
			}
			return m, nil
		case key.Matches(msg, m.keys.Confirm):
			id := m.selectedContextID()
			if id == "" {
				return m, nil
			}
			var err error
			if m.contextMode == contextWorkspace {
				err = m.switchWorkspace(id)
			} else {
				err = m.switchBoard(id)
			}
			if err != nil {
				m.statusLine = err.Error()
				return m, nil
			}
			m.closeContextPanel()
			return m, m.loadTasksCmd()
		}
	}

	var cmd tea.Cmd
	m.contextFilter, cmd = m.contextFilter.Update(msg)
	m.clampContextSelection()
	return m, cmd
}

func (m Model) renderContextPanel(base string) string {
	_ = base
	panelWidth := m.width * 2 / 3
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

	if m.boardOrder != nil {
		panel := m.renderBoardColumnsOrderContextPanel(contentWidth, contentHeight)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
	}

	if m.boardForm != nil {
		panel := m.renderBoardCreateContextPanel(contentWidth, contentHeight)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
	}

	listHeight := max(1, contentHeight-6)

	items := m.contextItems()
	offset := 0
	if m.contextSelected >= listHeight {
		offset = m.contextSelected - listHeight + 1
	}
	if maxOffset := max(0, len(items)-listHeight); offset > maxOffset {
		offset = maxOffset
	}

	helpText := "Type to filter | Enter: switch | n:create | r:rename | Esc:close"
	if m.contextMode == contextBoard {
		helpText = "Type to filter | Enter: switch | n:create (name + columns + colors) | r:rename | o:reorder columns | Esc:close"
	}

	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("151")).Render(m.contextTitle()),
		lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(helpText),
		m.contextFilter.View(),
		"",
	}

	if len(items) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("(empty)"))
	} else {
		end := offset + listHeight
		if end > len(items) {
			end = len(items)
		}
		for i := offset; i < end; i++ {
			id := items[i]
			name := m.contextNameByID(id)
			line := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(name)
			if i == m.contextSelected {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Render(line)
			}
			lines = append(lines, line)
		}
	}

	if m.contextEditMode != contextEditNone {
		label := "Create"
		if m.contextEditMode == contextEditRename {
			label = "Rename"
		}
		lines = append(lines, "", lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Render(label+": "+m.contextEditInput.View()))
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

func (m Model) renderBoardColumnsOrderContextPanel(contentWidth, contentHeight int) string {
	if m.boardOrder == nil {
		return ""
	}
	form := m.boardOrder
	form.clampSelection()

	help := "Up/Down select | Ctrl+K move up | Ctrl+J move down | Enter save | Esc cancel"
	header := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("151")).Render("Reorder Columns"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(help),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render("Board: " + form.boardName),
		"",
	}

	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(fmt.Sprintf("Columns: %d", len(form.columns)))
	availableLines := contentHeight - len(header) - 2 // spacer + footer
	if availableLines < 1 {
		availableLines = 1
	}

	offset := 0
	if form.selected >= availableLines {
		offset = form.selected - availableLines + 1
	}
	maxOffset := max(0, len(form.columns)-availableLines)
	if offset > maxOffset {
		offset = maxOffset
	}
	end := min(len(form.columns), offset+availableLines)

	lines := make([]string, 0, contentHeight+2)
	lines = append(lines, header...)

	rowWidth := max(12, contentWidth-4)
	for i := offset; i < end; i++ {
		col := form.columns[i]
		name := strings.TrimSpace(col.Name)
		if name == "" {
			name = "(unnamed)"
		}
		colorHex := strings.ToUpper(strings.TrimSpace(col.Color))
		if colorHex == "" {
			colorHex = "#9CA3AF"
		}
		swatch := lipgloss.NewStyle().
			Width(2).
			Background(colorFromHexOrDefault(colorHex, "240")).
			Render("  ")
		base := fmt.Sprintf("%2d. %s %s %s", i+1, swatch, truncate(name, max(4, rowWidth-18)), colorHex)
		line := lipgloss.NewStyle().Width(rowWidth).Foreground(lipgloss.Color("252")).Render(base)
		if i == form.selected {
			line = lipgloss.NewStyle().Width(rowWidth).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Render(base)
		}
		lines = append(lines, line)
	}

	lines = append(lines, "", footer)

	return lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("151")).
		Render(strings.Join(lines, "\n"))
}

func (m Model) renderBoardCreateContextPanel(contentWidth, contentHeight int) string {
	if m.boardForm == nil {
		return ""
	}

	focusedColumnIndex, hasFocusedColumn := m.boardForm.focusedColumnIndex()
	headerLines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("151")).Render("Create Board"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("Tab/Shift+Tab focus | Ctrl+N add column | Ctrl+D remove column | \u2190/\u2192 pick color | Enter save | Esc cancel"),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render("Board Name"),
		m.boardForm.boardName.View(),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render(fmt.Sprintf("Columns (%d)", len(m.boardForm.columns))),
	}

	footerLine := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("You can type HEX directly or use \u2190/\u2192 on color fields. Add as many columns as needed.")

	const columnBlockHeight = 4                            // label + input + label + input
	availableLines := contentHeight - len(headerLines) - 2 // reserve spacer + footer
	if availableLines < columnBlockHeight {
		availableLines = columnBlockHeight
	}
	visibleColumns := max(1, availableLines/columnBlockHeight)

	offset := 0
	if hasFocusedColumn && focusedColumnIndex >= visibleColumns {
		offset = focusedColumnIndex - visibleColumns + 1
	}
	maxOffset := max(0, len(m.boardForm.columns)-visibleColumns)
	if offset > maxOffset {
		offset = maxOffset
	}

	end := min(len(m.boardForm.columns), offset+visibleColumns)
	lines := make([]string, 0, contentHeight+4)
	lines = append(lines, headerLines...)

	if offset > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("↑ more columns"))
	}

	for i := offset; i < end; i++ {
		row := m.boardForm.columns[i]
		preview := lipgloss.NewStyle().
			Width(2).
			Background(colorFromHexOrDefault(row.color.Value(), "240")).
			Render("  ")
		lines = append(lines,
			lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render(fmt.Sprintf("Column %d Name", i+1)),
			row.name.View(),
			lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render(fmt.Sprintf("Column %d Color", i+1)),
			lipgloss.JoinHorizontal(lipgloss.Left, preview, " ", row.color.View()),
		)
	}

	if end < len(m.boardForm.columns) {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("↓ more columns"))
	}

	lines = append(lines, "", footerLine)

	return lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("151")).
		Render(strings.Join(lines, "\n"))
}

func (m Model) statusFilterLabel() string {
	if m.filterIndex >= 0 && m.filterIndex < len(m.columns) {
		return m.columns[m.filterIndex].Name
	}
	if strings.TrimSpace(m.columnFilter) != "" {
		return m.columnName(m.columnFilter)
	}
	return "All"
}

func (m Model) dueFilterLabel() string {
	switch m.dueFilter {
	case dueFilterSoon:
		return "Due in 7d"
	case dueFilterOverdue:
		return "Overdue"
	case dueFilterNoDate:
		return "No due date"
	default:
		return "Any"
	}
}

func (m Model) priorityFilterLabel() string {
	switch m.priorityFilter {
	case 0:
		return "Critical (0)"
	case 1:
		return "Urgent (1)"
	case 2:
		return "High (2)"
	case 3:
		return "Medium (3)"
	case 4:
		return "Low (4)"
	case 5:
		return "None (5)"
	default:
		return "All"
	}
}

func (m Model) sortModeLabel() string {
	switch m.sortMode {
	case sortByDueDate:
		return "Due date"
	case sortByTitle:
		return "Title"
	case sortByUpdated:
		return "Updated"
	case sortByCreated:
		return "Created"
	default:
		return "Priority"
	}
}

func (m *Model) setStatusFilterByIndex(index int) {
	m.filterIndex = index
	if index < 0 || index >= len(m.columns) {
		m.filterIndex = -1
		m.columnFilter = ""
		return
	}
	m.columnFilter = m.columns[index].ID
}

func (m *Model) cycleDueFilter() {
	next := int(m.dueFilter) + 1
	if next > int(dueFilterNoDate) {
		next = int(dueFilterAny)
	}
	m.dueFilter = dueFilterMode(next)
}

func (m *Model) cycleSortMode() {
	next := int(m.sortMode) + 1
	if next > int(sortByCreated) {
		next = int(sortByPriority)
	}
	m.sortMode = taskSortMode(next)
}

func workspaceIndexByID(items []domain.Workspace, id string) int {
	for i, item := range items {
		if item.ID == id {
			return i
		}
	}
	return -1
}

func boardIndexByID(items []domain.Board, id string) int {
	for i, item := range items {
		if item.ID == id {
			return i
		}
	}
	return -1
}

func (m *Model) switchBoardByOffset(delta int) (bool, error) {
	if len(m.boards) == 0 {
		return false, nil
	}

	current := boardIndexByID(m.boards, m.boardID)
	if current < 0 {
		current = 0
	}
	next := (current + delta + len(m.boards)) % len(m.boards)
	if m.boards[next].ID == m.boardID {
		return false, nil
	}
	if err := m.switchBoard(m.boards[next].ID); err != nil {
		return false, err
	}
	return true, nil
}

func (m *Model) adjustFilterSelection(delta int) (bool, error) {
	changed := false
	switch m.filterFocus {
	case 0: // workspace
		if len(m.workspaces) == 0 {
			return false, nil
		}
		current := workspaceIndexByID(m.workspaces, m.workspaceID)
		if current < 0 {
			current = 0
		}
		next := (current + delta + len(m.workspaces)) % len(m.workspaces)
		if m.workspaces[next].ID != m.workspaceID {
			if err := m.switchWorkspace(m.workspaces[next].ID); err != nil {
				return false, err
			}
			changed = true
		}
	case 1: // board
		if len(m.boards) == 0 {
			return false, nil
		}
		current := boardIndexByID(m.boards, m.boardID)
		if current < 0 {
			current = 0
		}
		next := (current + delta + len(m.boards)) % len(m.boards)
		if m.boards[next].ID != m.boardID {
			if err := m.switchBoard(m.boards[next].ID); err != nil {
				return false, err
			}
			changed = true
		}
	case 2: // status
		total := len(m.columns) + 1 // + All
		if total <= 0 {
			return false, nil
		}
		current := m.filterIndex + 1
		next := (current + delta + total) % total
		m.setStatusFilterByIndex(next - 1)
		changed = true
	case 3: // due
		total := int(dueFilterNoDate) + 1
		next := (int(m.dueFilter) + delta + total) % total
		m.dueFilter = dueFilterMode(next)
		changed = true
	case 4: // priority
		total := 7 // all + 0..5
		current := m.priorityFilter + 1
		next := (current + delta + total) % total
		m.priorityFilter = next - 1
		changed = true
	case 5: // sort
		total := int(sortByCreated) + 1
		next := (int(m.sortMode) + delta + total) % total
		m.sortMode = taskSortMode(next)
		changed = true
	}
	return changed, nil
}

func (m Model) updateFilterPanel(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.statusLine = msg.err.Error()
			return m, nil
		}
		m.tasks = m.applyActiveFilters(msg.tasks)
		m.sortTasks(m.tasks)
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel), key.Matches(msg, m.keys.ShowFilters):
			m.closeFilterPanel()
			return m, nil
		case key.Matches(msg, m.keys.Up):
			m.filterFocus--
			if m.filterFocus < 0 {
				m.filterFocus = 5
			}
			return m, nil
		case key.Matches(msg, m.keys.Down):
			m.filterFocus++
			if m.filterFocus > 5 {
				m.filterFocus = 0
			}
			return m, nil
		case key.Matches(msg, m.keys.Left):
			if changed, err := m.adjustFilterSelection(-1); err != nil {
				m.statusLine = err.Error()
				return m, nil
			} else if changed {
				return m, m.loadTasksCmd()
			}
			return m, nil
		case key.Matches(msg, m.keys.Right):
			if changed, err := m.adjustFilterSelection(1); err != nil {
				m.statusLine = err.Error()
				return m, nil
			} else if changed {
				return m, m.loadTasksCmd()
			}
			return m, nil
		case key.Matches(msg, m.keys.Confirm):
			m.closeFilterPanel()
			return m, nil
		}
	}
	return m, nil
}

func (m Model) renderFilterPanel(base string) string {
	_ = base
	panelWidth := m.width * 2 / 3
	if panelWidth < 56 {
		panelWidth = 56
	}
	if panelWidth > m.width-2 {
		panelWidth = max(20, m.width-2)
	}
	panelHeight := 12
	if panelHeight > m.height-2 {
		panelHeight = max(8, m.height-2)
	}

	contentWidth := boxContentWidth(panelWidth, 1, true)
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("151")).Render("Filter & Sort"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("↑/↓ select row | ←/→ change value | Enter/Esc close"),
		"",
	}

	row := func(index int, label, value string) string {
		left := lipgloss.NewStyle().Width(12).Foreground(lipgloss.Color("246")).Render(label)
		right := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(value)
		line := left + " " + right
		if m.filterFocus == index {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Render(line)
		}
		return line
	}

	lines = append(lines,
		row(0, "Workspace", m.workspaceName),
		row(1, "Board", m.boardName),
		row(2, "Status", m.statusFilterLabel()),
		row(3, "Due", m.dueFilterLabel()),
		row(4, "Priority", m.priorityFilterLabel()),
		row(5, "Sort", m.sortModeLabel()),
	)

	panel := lipgloss.NewStyle().
		Width(contentWidth).
		Height(boxContentHeight(panelHeight, true)).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("151")).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}

func (m Model) keybindEntries() []keybindEntry {
	entries := []keybindEntry{
		{ID: "new_task", Key: "n", Label: "Create task"},
		{ID: "edit_task", Key: "e", Label: "Edit selected task"},
		{ID: "edit_description", Key: "E", Label: "Edit description"},
		{ID: "add_comment", Key: "c", Label: "Add comment"},
		{ID: "search", Key: "/", Label: "Search"},
		{ID: "open_filters", Key: "f", Label: "Open filter/sort panel"},
		{ID: "open_workspaces", Key: "w", Label: "Open workspace switcher"},
		{ID: "open_board_panel", Key: "b", Label: "Open board manager"},
		{ID: "prev_board", Key: "[", Label: "Previous board"},
		{ID: "next_board", Key: "]", Label: "Next board"},
		{ID: "toggle_details", Key: "d", Label: "Toggle details pane"},
		{ID: "open_move", Key: "Enter", Label: "Open task viewer"},
		{ID: "move_task", Key: "m", Label: "Move task to next status"},
		{ID: "toggle_view", Key: "Tab", Label: "Switch list/kanban"},
		{ID: "cycle_status", Key: "s", Label: "Cycle status filter"},
		{ID: "cycle_due_filter", Key: "z", Label: "Cycle due filter"},
		{ID: "cycle_sort", Key: "o", Label: "Cycle sort mode"},
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
	case "open_filters":
		m.openFilterPanel()
		return m, nil
	case "open_workspaces":
		m.openContextPanel(contextWorkspace)
		return m, textinput.Blink
	case "open_board_panel":
		m.openContextPanel(contextBoard)
		return m, textinput.Blink
	case "prev_board":
		changed, err := m.switchBoardByOffset(-1)
		if err != nil {
			m.statusLine = err.Error()
			return m, nil
		}
		if changed {
			return m, m.loadTasksCmd()
		}
		return m, nil
	case "next_board":
		changed, err := m.switchBoardByOffset(1)
		if err != nil {
			m.statusLine = err.Error()
			return m, nil
		}
		if changed {
			return m, m.loadTasksCmd()
		}
		return m, nil
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
		return m, m.startExternalDescriptionEdit(task)
	case "cycle_status":
		m.cycleColumnFilter()
		return m, m.loadTasksCmd()
	case "cycle_due_filter":
		m.cycleDueFilter()
		return m, m.loadTasksCmd()
	case "cycle_sort":
		m.cycleSortMode()
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
		return m, m.openTaskViewer()
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
	filterParts := []string{fmt.Sprintf("status:%s", m.statusFilterLabel()), fmt.Sprintf("due:%s", strings.ToLower(m.dueFilterLabel()))}
	if m.priorityFilter >= 0 {
		filterParts = append(filterParts, fmt.Sprintf("priority:p%d", m.priorityFilter))
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))

	left := headerStyle.Render(fmt.Sprintf("%s / %s", m.workspaceName, m.boardName))
	right := metaStyle.Render(fmt.Sprintf("view:%s  sort:%s  filter:%s  search:%q", viewLabel, strings.ToLower(m.sortModeLabel()), strings.Join(filterParts, ","), m.titleFilter))
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

	shortcuts := "?:keybinds w:workspaces b:board-manager [ ]:boards f:filters s/z:quick-filter o:sort n:new /:search enter:open-task q:quit"
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
		m.setStatusFilterByIndex(-1)
		return
	}
	m.setStatusFilterByIndex(m.filterIndex + 1)
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
		BoardID:     m.boardID,
		TitleQuery:  m.titleFilter,
		ColumnID:    m.columnFilter,
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
