package ui

import "testing"

func TestOverlayActiveOverlay_None(t *testing.T) {
	o := overlayState{}
	if got := o.activeOverlay(); got != overlayNone {
		t.Errorf("activeOverlay() = %v, want overlayNone", got)
	}
}

func TestOverlayActiveOverlay_PriorityOrder(t *testing.T) {
	// taskView dominates over keybinds
	o := overlayState{showTaskView: true, showKeybinds: true}
	if got := o.activeOverlay(); got != overlayTaskView {
		t.Errorf("activeOverlay() = %v, want overlayTaskView", got)
	}

	// keybinds dominates over filters
	o = overlayState{showKeybinds: true, showFilters: true}
	if got := o.activeOverlay(); got != overlayKeybinds {
		t.Errorf("activeOverlay() = %v, want overlayKeybinds", got)
	}

	// filters dominates over contexts
	o = overlayState{showFilters: true, showContexts: true}
	if got := o.activeOverlay(); got != overlayFilters {
		t.Errorf("activeOverlay() = %v, want overlayFilters", got)
	}

	// contexts dominates over input
	o = overlayState{showContexts: true, inputMode: inputSearch}
	if got := o.activeOverlay(); got != overlayContexts {
		t.Errorf("activeOverlay() = %v, want overlayContexts", got)
	}

	// input alone
	o = overlayState{inputMode: inputSearch}
	if got := o.activeOverlay(); got != overlayInput {
		t.Errorf("activeOverlay() = %v, want overlayInput", got)
	}
}

func TestOverlayOpenKeybinds(t *testing.T) {
	o := overlayState{}
	o.openKeybinds()
	if !o.showKeybinds {
		t.Error("expected showKeybinds to be true")
	}
	if o.keySelected != 0 {
		t.Errorf("keySelected = %d, want 0", o.keySelected)
	}
	if o.activeOverlay() != overlayKeybinds {
		t.Error("expected activeOverlay to be overlayKeybinds")
	}
}

func TestOverlayOpenKeybinds_PreservesOthers(t *testing.T) {
	o := overlayState{
		showTaskView: true,
		showFilters:  true,
		inputMode:    inputSearch,
		taskForm:     &taskForm{},
	}
	o.openKeybinds()
	if !o.showTaskView {
		t.Error("expected showTaskView to be preserved")
	}
	if !o.showFilters {
		t.Error("expected showFilters to be preserved")
	}
	if o.inputMode != inputSearch {
		t.Errorf("inputMode = %v, want inputSearch", o.inputMode)
	}
	if o.taskForm == nil {
		t.Error("expected taskForm to be preserved")
	}
}

func TestOverlayCloseKeybinds(t *testing.T) {
	o := overlayState{showKeybinds: true, keySelected: 5}
	o.closeKeybinds()
	if o.showKeybinds {
		t.Error("expected showKeybinds to be false")
	}
	if o.keySelected != 0 {
		t.Errorf("keySelected = %d, want 0", o.keySelected)
	}
}

func TestOverlayOpenFilters(t *testing.T) {
	o := overlayState{}
	o.openFilters()
	if !o.showFilters {
		t.Error("expected showFilters to be true")
	}
	if o.filterFocus != 0 {
		t.Errorf("filterFocus = %d, want 0", o.filterFocus)
	}
	if o.activeOverlay() != overlayFilters {
		t.Error("expected activeOverlay to be overlayFilters")
	}
}

func TestOverlayOpenFilters_PreservesOthers(t *testing.T) {
	o := overlayState{showKeybinds: true, showTaskView: true}
	o.openFilters()
	if !o.showKeybinds {
		t.Error("expected showKeybinds to be preserved")
	}
	if !o.showTaskView {
		t.Error("expected showTaskView to be preserved")
	}
}

func TestOverlayCloseFilters(t *testing.T) {
	o := overlayState{showFilters: true, filterFocus: 3}
	o.closeFilters()
	if o.showFilters {
		t.Error("expected showFilters to be false")
	}
}

func TestOverlayOpenContexts(t *testing.T) {
	o := overlayState{}
	o.openContexts(contextBoard)
	if !o.showContexts {
		t.Error("expected showContexts to be true")
	}
	if o.contextMode != contextBoard {
		t.Errorf("contextMode = %v, want contextBoard", o.contextMode)
	}
	if o.contextSelected != 0 {
		t.Errorf("contextSelected = %d, want 0", o.contextSelected)
	}
	if o.contextEditMode != contextEditNone {
		t.Errorf("contextEditMode = %v, want contextEditNone", o.contextEditMode)
	}
	if o.boardForm != nil {
		t.Error("expected boardForm to be nil")
	}
	if o.boardOrder != nil {
		t.Error("expected boardOrder to be nil")
	}
}

func TestOverlayOpenContexts_PreservesOthers(t *testing.T) {
	o := overlayState{showKeybinds: true, inputMode: inputSearch, taskForm: &taskForm{}}
	o.openContexts(contextWorkspace)
	if !o.showKeybinds {
		t.Error("expected showKeybinds to be preserved")
	}
	if o.inputMode != inputSearch {
		t.Errorf("inputMode = %v, want inputSearch", o.inputMode)
	}
	if o.taskForm == nil {
		t.Error("expected taskForm to be preserved")
	}
}

func TestOverlayCloseContexts(t *testing.T) {
	o := overlayState{
		showContexts:    true,
		contextEditMode: contextEditCreate,
		boardForm:       &boardCreateForm{},
		boardOrder:      &boardColumnsOrderForm{},
	}
	o.closeContexts()
	if o.showContexts {
		t.Error("expected showContexts to be false")
	}
	if o.contextEditMode != contextEditNone {
		t.Errorf("contextEditMode = %v, want contextEditNone", o.contextEditMode)
	}
	if o.boardForm != nil {
		t.Error("expected boardForm to be nil")
	}
	if o.boardOrder != nil {
		t.Error("expected boardOrder to be nil")
	}
}

func TestOverlayOpenTaskView(t *testing.T) {
	o := overlayState{}
	o.openTaskView("task-123")
	if !o.showTaskView {
		t.Error("expected showTaskView to be true")
	}
	if o.viewTaskID != "task-123" {
		t.Errorf("viewTaskID = %q, want %q", o.viewTaskID, "task-123")
	}
	if o.viewDescScroll != 0 {
		t.Errorf("viewDescScroll = %d, want 0", o.viewDescScroll)
	}
	if o.activeOverlay() != overlayTaskView {
		t.Error("expected activeOverlay to be overlayTaskView")
	}
}

func TestOverlayOpenTaskView_PreservesOthers(t *testing.T) {
	o := overlayState{showKeybinds: true, inputMode: inputTaskForm, taskForm: &taskForm{}}
	o.openTaskView("task-1")
	if !o.showKeybinds {
		t.Error("expected showKeybinds to be preserved")
	}
	if o.inputMode != inputTaskForm {
		t.Errorf("inputMode = %v, want inputTaskForm", o.inputMode)
	}
	if o.taskForm == nil {
		t.Error("expected taskForm to be preserved")
	}
}

func TestOverlayCloseTaskView(t *testing.T) {
	o := overlayState{
		showTaskView:   true,
		viewTaskID:     "task-1",
		viewDescScroll: 5,
	}
	o.closeTaskView()
	if o.showTaskView {
		t.Error("expected showTaskView to be false")
	}
	if o.viewTaskID != "" {
		t.Errorf("viewTaskID = %q, want empty", o.viewTaskID)
	}
	if o.viewDescScroll != 0 {
		t.Errorf("viewDescScroll = %d, want 0", o.viewDescScroll)
	}
}

func TestOverlayTaskViewerReturn(t *testing.T) {
	o := overlayState{}
	o.setTaskViewerReturn("task-42")
	if !o.returnTaskView {
		t.Error("expected returnTaskView to be true")
	}
	if o.returnTaskID != "task-42" {
		t.Errorf("returnTaskID = %q, want %q", o.returnTaskID, "task-42")
	}

	o.clearTaskViewerReturn()
	if o.returnTaskView {
		t.Error("expected returnTaskView to be false")
	}
	if o.returnTaskID != "" {
		t.Errorf("returnTaskID = %q, want empty", o.returnTaskID)
	}
}

func TestOverlayStartTaskForm(t *testing.T) {
	form := &taskForm{mode: taskFormCreate}
	o := overlayState{}
	o.startTaskForm(form)
	if o.inputMode != inputTaskForm {
		t.Errorf("inputMode = %v, want inputTaskForm", o.inputMode)
	}
	if o.taskForm != form {
		t.Error("expected taskForm to be set")
	}
	if o.activeOverlay() != overlayInput {
		t.Error("expected activeOverlay to be overlayInput")
	}
}

func TestOverlayStartTaskForm_PreservesOthers(t *testing.T) {
	form := &taskForm{}
	o := overlayState{showTaskView: true, showKeybinds: true}
	o.startTaskForm(form)
	if !o.showTaskView {
		t.Error("expected showTaskView to be preserved")
	}
	if !o.showKeybinds {
		t.Error("expected showKeybinds to be preserved")
	}
}

func TestOverlayCloseTaskForm(t *testing.T) {
	o := overlayState{
		inputMode: inputTaskForm,
		taskForm:  &taskForm{},
	}
	o.closeTaskForm()
	if o.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", o.inputMode)
	}
	if o.taskForm != nil {
		t.Error("expected taskForm to be nil")
	}
}

func TestOverlayStartTaskForm_PreservesReturnState(t *testing.T) {
	o := overlayState{returnTaskView: true, returnTaskID: "task-99"}
	o.startTaskForm(&taskForm{})
	if !o.returnTaskView {
		t.Error("expected returnTaskView to be preserved")
	}
	if o.returnTaskID != "task-99" {
		t.Errorf("returnTaskID = %q, want %q", o.returnTaskID, "task-99")
	}
}

func TestOverlayOpenTaskView_PreservesReturnState(t *testing.T) {
	o := overlayState{returnTaskView: true, returnTaskID: "task-99"}
	o.openTaskView("task-1")
	if !o.returnTaskView {
		t.Error("expected returnTaskView to be preserved")
	}
	if o.returnTaskID != "task-99" {
		t.Errorf("returnTaskID = %q, want %q", o.returnTaskID, "task-99")
	}
}
