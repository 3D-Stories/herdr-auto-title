package resolver

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rivo/uniseg"

	"github.com/kryptamine/herdr-auto-title/internal/herdr/herdrtest"
	"github.com/kryptamine/herdr-auto-title/internal/state"
)

func TestTabPaneIDLeadsTheTitle(t *testing.T) {
	t.Parallel()

	got := NewTabPaneIDs(defaultChain(), DefaultMaxLength, false).Resolve(tabWithCWD(dashboard))
	if got.Name != "wE:p1 · dashboard" {
		t.Errorf("name = %q, want %q", got.Name, "wE:p1 · dashboard")
	}
}

func TestTabPaneIDGoesInFrontOfThePosition(t *testing.T) {
	t.Parallel()

	got := NewTabPaneIDs(defaultChain(), DefaultMaxLength, true).Resolve(atPosition(2, dashboard))
	if got.Name != "wE:p1 · 2 · dashboard" {
		t.Errorf("name = %q, want %q", got.Name, "wE:p1 · 2 · dashboard")
	}
}

func TestTabPaneIDIsThePaneTheTabIsNamedAfter(t *testing.T) {
	t.Parallel()

	// The first pane by ID is not the one the title describes, so naming it
	// would point a script at the wrong pane.
	tab := tabOf([]*state.PaneState{
		{ID: "wE:p1", Dir: api},
		{ID: "wE:p2", Dir: dashboard, Focused: true},
	})

	got := NewTabPaneIDs(defaultChain(), DefaultMaxLength, false).Resolve(tab)
	if got.Name != "wE:p2 · dashboard" {
		t.Errorf("name = %q, want %q", got.Name, "wE:p2 · dashboard")
	}
}

func TestATabWithNoPanesHasNoPaneID(t *testing.T) {
	t.Parallel()

	// A tab read before any of its panes has no context pane to name.
	tab := tabOf(nil)
	tab.Position = 3

	want := NewNumbered(defaultChain(), DefaultMaxLength).Resolve(tab).Name

	got := NewTabPaneIDs(defaultChain(), DefaultMaxLength, true).Resolve(tab)
	if got.Name != want {
		t.Errorf("name = %q, want %q", got.Name, want)
	}
}

func TestTabPaneIDKeepsTheDecisionItWraps(t *testing.T) {
	t.Parallel()

	inner := fixedResolver{decision: Decision{
		Name:       "release notes",
		Confidence: ConfidenceAgent,
		Reason:     "test_source",
	}}

	got := NewTabPaneIDs(inner, DefaultMaxLength, true).Resolve(atPosition(4, api))
	want := Decision{
		Name:       "wE:p1 · 4 · release notes",
		Confidence: ConfidenceAgent,
		Reason:     "test_source",
	}

	if got != want {
		t.Errorf("decision = %+v, want %+v", got, want)
	}
}

func TestTabPaneIDWithoutAWidthTakesTheDefault(t *testing.T) {
	t.Parallel()

	got := NewTabPaneIDs(New(Options{}, NewCWD("")), 0, true).Resolve(atPosition(2, api))
	if got.Name != "wE:p1 · 2 · api" {
		t.Errorf("name = %q, want the ID and the position kept", got.Name)
	}
}

// The ID and the position are fitted as one prefix: fitted one after the other,
// a narrow title ends on the position mark or says nothing but which tab it is.
func TestATabPaneIDNeverLeavesATitleWithoutAName(t *testing.T) {
	t.Parallel()

	long := herdrtest.Dir("work", strings.Repeat("a", 40))

	for maxLength := 1; maxLength <= DefaultMaxLength; maxLength++ {
		for _, withPosition := range []bool{false, true} {
			t.Run(fmt.Sprintf("width %d position %t", maxLength, withPosition), func(t *testing.T) {
				t.Parallel()

				chain := New(Options{MaxLength: maxLength}, NewCWD(""))
				titles := NewTabPaneIDs(chain, maxLength, withPosition)
				got := titles.Resolve(atPosition(2, long)).Name

				if width := uniseg.StringWidth(got); width > maxLength {
					t.Errorf("name %q is %d columns wide, want at most %d", got, width, maxLength)
				}

				rest := strings.TrimPrefix(got, "wE:p1 · ")
				if withPosition {
					rest = strings.TrimPrefix(rest, "2 · ")
				}

				if strings.HasPrefix(got, "wE:p1 · ") && !strings.HasPrefix(rest, "a") {
					t.Errorf("name = %q, want a name after the ID", got)
				}

				if strings.HasSuffix(got, "·") || strings.HasSuffix(got, ":") {
					t.Errorf("name = %q ends on a mark", got)
				}
			})
		}
	}
}

func TestATabPaneIDWithNoRoomLeavesThePosition(t *testing.T) {
	t.Parallel()

	// "wE:p1 · 2 · " takes 12 columns, so 12 leaves no room for a name, and the
	// position is what switches to the tab.
	chain := New(Options{MaxLength: 12}, NewCWD(""))

	got := NewTabPaneIDs(chain, 12, true).Resolve(atPosition(2, dashboard))
	if got.Name != "2 · dashboar" {
		t.Errorf("name = %q, want %q", got.Name, "2 · dashboar")
	}
}

// A name that repeats the prefix it is about to get must still get it: telling a
// kept prefix from a dropped one by comparing names would take one for the other.
func TestATabPaneIDIsKeptWhenTheNameRepeatsIt(t *testing.T) {
	t.Parallel()

	inner := fixedResolver{decision: Decision{Name: "wE:p1 · 2 · wE:p1"}}

	got := NewTabPaneIDs(inner, 17, true).Resolve(atPosition(2, api))
	if got.Name != "wE:p1 · 2 · wE:p1" {
		t.Errorf("name = %q, want %q", got.Name, "wE:p1 · 2 · wE:p1")
	}
}

func TestATabWithNoPanesAndNoPositionKeepsItsName(t *testing.T) {
	t.Parallel()

	want := defaultChain().Resolve(tabOf(nil)).Name

	got := NewTabPaneIDs(defaultChain(), DefaultMaxLength, false).Resolve(tabOf(nil))
	if got.Name != want {
		t.Errorf("name = %q, want %q", got.Name, want)
	}
}
