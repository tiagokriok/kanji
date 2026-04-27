package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tiagokriok/kanji/internal/application"
	"github.com/tiagokriok/kanji/internal/domain"
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
