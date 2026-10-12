package resolver

import (
	"strconv"

	"github.com/kryptamine/herdr-auto-title/internal/state"
)

// TabPaneIDs prefixes every tab title with the Herdr ID of the pane the tab is
// named after, so the pane a script or another agent should address is in the
// tab bar. It wraps a resolver for the reason Numbered does.
type TabPaneIDs struct {
	inner TitleResolver
	// maxLength is the bound inner was built with; the ID is counted against
	// it, as Numbered counts the position.
	maxLength int
	// withPosition puts the tab's position after the ID. It is fitted here
	// rather than by wrapping Numbered: see Resolve.
	withPosition bool
}

var _ TitleResolver = (*TabPaneIDs)(nil)

// NewTabPaneIDs wraps inner, which must have been given the same maxLength.
// Zero or less takes the default, as New does.
func NewTabPaneIDs(inner TitleResolver, maxLength int, withPosition bool) *TabPaneIDs {
	if maxLength <= 0 {
		maxLength = DefaultMaxLength
	}

	return &TabPaneIDs{inner: inner, maxLength: maxLength, withPosition: withPosition}
}

// Resolve names the tab and puts its pane's ID, then its position, in front.
// The two are fitted as one prefix: fitted one after the other, a title too
// narrow for both would end on the position mark with no name left.
func (p *TabPaneIDs) Resolve(tab state.TabState) Decision {
	decision := p.inner.Resolve(tab)

	position := ""
	if p.withPosition {
		position = strconv.Itoa(tab.Position) + positionMark
	}

	// A tab read before any of its panes has no pane to name.
	if tab.Context != nil {
		named := withPrefix("["+tab.Context.ID+"] "+position, decision.Name, p.maxLength)
		if named != decision.Name {
			decision.Name = named

			return decision
		}
	}

	// The ID goes first when there is no room for both: the position is the
	// key that switches to the tab.
	if position != "" {
		decision.Name = withPrefix(position, decision.Name, p.maxLength)
	}

	return decision
}
