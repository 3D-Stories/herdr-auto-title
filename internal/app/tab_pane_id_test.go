package app

import (
	"testing"

	"github.com/kryptamine/herdr-auto-title/internal/herdr"
	"github.com/kryptamine/herdr-auto-title/internal/herdr/herdrtest"
)

// startTabPaneIDs drives the resolvers Resolvers builds with the tab pane ID
// turned on, which the plain harness's chain does not.
func startTabPaneIDs(t *testing.T, tabs []herdr.TabInfo, panes []herdr.PaneInfo) *harness {
	t.Helper()

	cfg := testConfig()
	cfg.ShowTabPaneID = true
	cfg.Home = testHome(t)

	titles, paneNames, topics := Resolvers(cfg)
	instance := &fakeInstance{}

	return &harness{
		t:        t,
		app:      New(cfg, discardLogger(), titles, paneNames, topics, instance),
		client:   herdrtest.New(tabs, panes),
		instance: instance,
	}
}

// splitTab is one tab holding two panes in the same directory, so the pane ID
// is the only thing that tells their titles apart.
func splitTab(focused string) []herdr.PaneInfo {
	return []herdr.PaneInfo{
		{PaneID: "wE:p1", TabID: "wE:t1", CWD: dashboard, Focused: focused == "wE:p1"},
		{PaneID: "wE:p2", TabID: "wE:t1", CWD: dashboard, Focused: focused == "wE:p2"},
	}
}

func moveFocus(h *harness, focused string, revision uint64) {
	for _, pane := range splitTab(focused) {
		pane.Revision = revision
		h.client.SetPane(pane)
	}
}

func TestMovingFocusRetitlesTheTabWithTheNewPaneID(t *testing.T) {
	t.Parallel()

	// The label the plugin set for the first pane must not be taken for one
	// the user typed when the second pane takes the title over.
	h := startTabPaneIDs(t, []herdr.TabInfo{{TabID: "wE:t1", Label: "1"}}, splitTab("wE:p1"))
	h.poll()

	moveFocus(h, "wE:p2", 2)
	h.poll()

	moveFocus(h, "wE:p1", 3)
	h.poll()

	want := []string{"[wE:p1] dashboard", "[wE:p2] dashboard", "[wE:p1] dashboard"}

	renames := h.client.Renames()
	if len(renames) != len(want) {
		t.Fatalf("renames = %+v, want %d", renames, len(want))
	}

	for i, label := range want {
		if renames[i].Label != label {
			t.Errorf("rename %d = %q, want %q", i, renames[i].Label, label)
		}
	}
}

func TestATabRenamedByHandStaysSoWithTabPaneIDsOn(t *testing.T) {
	t.Parallel()

	h := startTabPaneIDs(t, []herdr.TabInfo{{TabID: "wE:t1", Label: "1"}}, splitTab("wE:p1"))
	h.poll()

	h.client.SetTab(herdr.TabInfo{TabID: "wE:t1", Label: "release"})
	h.poll()

	moveFocus(h, "wE:p2", 2)
	h.poll()

	if renames := h.client.Renames(); len(renames) != 1 {
		t.Errorf("renames = %+v, want only the first", renames)
	}
}
