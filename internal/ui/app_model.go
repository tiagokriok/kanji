package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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
	m.overlayState.openContexts(mode)
	m.contextEditInput.SetValue("")
	m.contextFilter.SetValue("")
	m.contextFilter.Focus()
}

func (m *Model) closeContextPanel() {
	m.overlayState.closeContexts()
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
		return m.handleTasksLoaded(msg, false, true)
	case commentsLoadedMsg:
		return m.handleCommentsLoaded(msg)
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
	flow := m.taskFlow
	return func() tea.Msg {
		result, err := flow.MoveTaskAdjacent(context.Background(), task.ID, m.columns, task.ColumnID, 1)
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: result.Message, taskID: result.TaskID, columnID: result.ColumnID}
	}
}

func (m Model) moveToPrevColumnCmd(task domain.Task) tea.Cmd {
	if len(m.columns) == 0 {
		return nil
	}
	flow := m.taskFlow
	return func() tea.Msg {
		result, err := flow.MoveTaskAdjacent(context.Background(), task.ID, m.columns, task.ColumnID, -1)
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: result.Message, taskID: result.TaskID, columnID: result.ColumnID}
	}
}

func (m Model) deleteTaskCmd(id string) tea.Cmd {
	service := m.taskService
	return func() tea.Msg {
		err := service.DeleteTask(context.Background(), id)
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: "task deleted"}
	}
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
