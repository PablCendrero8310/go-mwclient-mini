package options

import "net/url"

type Options interface {
	ToParams() url.Values
}
type EditOptions struct {
	Summary      string   `json:"summary,omitempty"`
	Minor        bool     `json:"minor,omitempty"`
	Bot          bool     `json:"bot,omitempty"`
	CreateOnly   bool     `json:"createonly,omitempty"`
	NoCreate     bool     `json:"nocreate,omitempty"`
	BaseRevID    int      `json:"baserevid,omitempty"`
	Watchlist    string   `json:"watchlist,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Section      string   `json:"section,omitempty"`
	SectionTitle string   `json:"sectiontitle,omitempty"`
}
