package ui

// overlayKind identifies which overlay is currently active.
type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayKeybinds
	overlayFilters
	overlayContexts
	overlayTaskView
	overlayInput
)

// overlayState encapsulates the overlay and form workflow state.
// It manages which panel is active and enforces transitions between them.
type overlayState struct {
	showKeybinds bool
	showFilters  bool
	showContexts bool
	showTaskView bool
	inputMode    inputMode
	taskForm     *taskForm

	keySelected     int
	filterFocus     int
	contextSelected int
	contextMode     contextMode
	contextEditMode contextEditMode
	boardForm       *boardCreateForm
	boardOrder      *boardColumnsOrderForm

	viewTaskID     string
	viewDescScroll int
	returnTaskView bool
	returnTaskID   string
}

// activeOverlay returns the currently active overlay based on priority order.
func (o *overlayState) activeOverlay() overlayKind {
	if o.showTaskView {
		return overlayTaskView
	}
	if o.showKeybinds {
		return overlayKeybinds
	}
	if o.showFilters {
		return overlayFilters
	}
	if o.showContexts {
		return overlayContexts
	}
	if o.inputMode != inputNone {
		return overlayInput
	}
	return overlayNone
}

func (o *overlayState) openKeybinds() {
	o.showKeybinds = true
	o.keySelected = 0
}

func (o *overlayState) closeKeybinds() {
	o.showKeybinds = false
	o.keySelected = 0
}

func (o *overlayState) openFilters() {
	o.showFilters = true
	o.filterFocus = 0
}

func (o *overlayState) closeFilters() {
	o.showFilters = false
}

func (o *overlayState) openContexts(mode contextMode) {
	o.showContexts = true
	o.contextMode = mode
	o.contextSelected = 0
	o.contextEditMode = contextEditNone
	o.boardForm = nil
	o.boardOrder = nil
}

func (o *overlayState) closeContexts() {
	o.showContexts = false
	o.contextEditMode = contextEditNone
	o.boardForm = nil
	o.boardOrder = nil
}

func (o *overlayState) openTaskView(taskID string) {
	o.showTaskView = true
	o.viewTaskID = taskID
	o.viewDescScroll = 0
}

func (o *overlayState) closeTaskView() {
	o.showTaskView = false
	o.viewTaskID = ""
	o.viewDescScroll = 0
}

func (o *overlayState) setTaskViewerReturn(taskID string) {
	o.returnTaskView = true
	o.returnTaskID = taskID
}

func (o *overlayState) clearTaskViewerReturn() {
	o.returnTaskView = false
	o.returnTaskID = ""
}

func (o *overlayState) startTaskForm(form *taskForm) {
	o.inputMode = inputTaskForm
	o.taskForm = form
}

func (o *overlayState) closeTaskForm() {
	o.inputMode = inputNone
	o.taskForm = nil
}
