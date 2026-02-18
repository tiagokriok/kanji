package ui

import (
	"context"
	"fmt"
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
	taskFormStepTitle = iota
	taskFormStepDescription
	taskFormStepPriority
	taskFormStepStatus
	taskFormSteps = 4
)

type taskForm struct {
	mode        taskFormMode
	taskID      string
	step        int
	title       string
	description string
	priority    string
	status      string
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

type Model struct {
	taskService    *application.TaskService
	commentService *application.CommentService

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

	viewMode    viewMode
	showDetails bool
	inputMode   inputMode
	taskForm    *taskForm

	textInput textinput.Model
	textArea  textarea.Model

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

	cols := make([]domain.Column, 0, len(setup.Columns))
	cols = append(cols, setup.Columns...)
	sort.Slice(cols, func(i, j int) bool {
		return cols[i].Position < cols[j].Position
	})

	return Model{
		taskService:    taskService,
		commentService: commentService,
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
		keys:           newKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadTasksCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.inputMode != inputNone {
		return m.updateInputMode(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textArea.SetWidth(max(20, msg.Width/2-6))
		return m, nil
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

	if m.viewMode == viewList {
		return m.renderListScreen()
	}

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
	return lipgloss.NewStyle().Padding(0, 1).Render(content)
}

func (m Model) updateInputMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	mode := m.inputMode
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keys.Cancel):
			m.inputMode = inputNone
			m.taskForm = nil
			m.textInput.Blur()
			m.textArea.Blur()
			m.statusLine = ""
			return m, nil
		case keyMsg.String() == "ctrl+s" && mode == inputEditDescription:
			task, ok := m.currentTask()
			if !ok {
				m.inputMode = inputNone
				return m, nil
			}
			description := m.textArea.Value()
			m.inputMode = inputNone
			m.textArea.Blur()
			return m, m.updateTaskDescriptionCmd(task.ID, description)
		case key.Matches(keyMsg, m.keys.Confirm) && mode == inputTaskForm:
			return m.submitOrAdvanceTaskForm()
		case key.Matches(keyMsg, m.keys.Confirm) && mode != inputEditDescription:
			value := strings.TrimSpace(m.textInput.Value())
			m.inputMode = inputNone
			m.textInput.Blur()
			switch mode {
			case inputSearch:
				m.titleFilter = value
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

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.inputMode = mode
	return m, cmd
}

func (m *Model) startCreateTaskForm() {
	defaultStatus := ""
	if len(m.columns) > 0 {
		defaultStatus = m.columns[0].Name
	}
	m.taskForm = &taskForm{
		mode:     taskFormCreate,
		step:     taskFormStepTitle,
		priority: "0",
		status:   defaultStatus,
	}
	m.inputMode = inputTaskForm
	m.loadCurrentTaskFormStep()
	m.textInput.Focus()
}

func (m *Model) startEditTaskForm(task domain.Task) {
	status := ""
	if task.ColumnID != nil {
		status = m.columnName(*task.ColumnID)
	} else if task.Status != nil {
		status = *task.Status
	}
	m.taskForm = &taskForm{
		mode:        taskFormEdit,
		taskID:      task.ID,
		step:        taskFormStepTitle,
		title:       task.Title,
		description: task.DescriptionMD,
		priority:    strconv.Itoa(task.Priority),
		status:      status,
	}
	m.inputMode = inputTaskForm
	m.loadCurrentTaskFormStep()
	m.textInput.Focus()
}

func (m *Model) submitOrAdvanceTaskForm() (tea.Model, tea.Cmd) {
	if m.taskForm == nil {
		m.inputMode = inputNone
		return m, nil
	}

	value := strings.TrimSpace(m.textInput.Value())
	switch m.taskForm.step {
	case taskFormStepTitle:
		m.taskForm.title = value
	case taskFormStepDescription:
		m.taskForm.description = value
	case taskFormStepPriority:
		m.taskForm.priority = value
	case taskFormStepStatus:
		m.taskForm.status = value
	}

	if m.taskForm.step < taskFormSteps-1 {
		m.taskForm.step++
		m.loadCurrentTaskFormStep()
		m.textInput.Focus()
		return m, textinput.Blink
	}

	cmd, err := m.submitTaskFormCmd()
	if err != nil {
		m.statusLine = err.Error()
		m.loadCurrentTaskFormStep()
		m.textInput.Focus()
		return m, textinput.Blink
	}

	m.inputMode = inputNone
	m.taskForm = nil
	m.textInput.Blur()
	return m, cmd
}

func (m *Model) loadCurrentTaskFormStep() {
	if m.taskForm == nil {
		return
	}

	modeLabel := "Create"
	if m.taskForm.mode == taskFormEdit {
		modeLabel = "Edit"
	}

	suffix := fmt.Sprintf("%s task (%d/%d) - ", modeLabel, m.taskForm.step+1, taskFormSteps)
	switch m.taskForm.step {
	case taskFormStepTitle:
		m.textInput.Placeholder = "Title"
		m.textInput.SetValue(m.taskForm.title)
		m.statusLine = suffix + "title"
	case taskFormStepDescription:
		m.textInput.Placeholder = "Description (markdown)"
		m.textInput.SetValue(m.taskForm.description)
		m.statusLine = suffix + "description"
	case taskFormStepPriority:
		m.textInput.Placeholder = "Priority (0-5)"
		m.textInput.SetValue(m.taskForm.priority)
		m.statusLine = suffix + "priority"
	case taskFormStepStatus:
		m.textInput.Placeholder = "Status/column (Todo|Doing|Done)"
		m.textInput.SetValue(m.taskForm.status)
		m.statusLine = suffix + "status/column"
	}
}

func (m Model) submitTaskFormCmd() (tea.Cmd, error) {
	if m.taskForm == nil {
		return nil, fmt.Errorf("task form is not active")
	}

	title := strings.TrimSpace(m.taskForm.title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	priorityRaw := strings.TrimSpace(m.taskForm.priority)
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

	columnID, status := m.resolveStatusInput(strings.TrimSpace(m.taskForm.status), m.taskForm.mode)
	description := m.taskForm.description

	if m.taskForm.mode == taskFormCreate {
		boardID := m.boardID
		return m.createTaskWithDetailsCmd(title, description, priority, &boardID, columnID, status), nil
	}

	return m.updateTaskWithDetailsCmd(
		m.taskForm.taskID,
		&title,
		&description,
		&priority,
		columnID,
		status,
	), nil
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

func (m Model) createTaskWithDetailsCmd(title, description string, priority int, boardID, columnID, status *string) tea.Cmd {
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
			Labels:        []string{},
		})
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: "task created"}
	}
}

func (m Model) updateTaskWithDetailsCmd(taskID string, title, description *string, priority *int, columnID, status *string) tea.Cmd {
	service := m.taskService
	return func() tea.Msg {
		err := service.UpdateTask(context.Background(), taskID, application.UpdateTaskInput{
			Title:         title,
			DescriptionMD: description,
			Priority:      priority,
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
	case inputSearch, inputAddComment, inputTaskForm:
		inputLine = lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Render(m.textInput.View())
	case inputEditDescription:
		inputLine = lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Render(m.textArea.View())
	}

	shortcuts := "n:new e:edit c:comment /:search tab:view d:details s:column z:due soon"
	if strings.TrimSpace(m.titleFilter) != "" {
		shortcuts += " x:clear-search"
	}
	lines := []string{}

	if strings.TrimSpace(m.statusLine) != "" {
		status := m.statusLine
		if m.inputMode == inputTaskForm {
			status += " | enter:next/save esc:cancel"
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("222")).Render(status))
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
