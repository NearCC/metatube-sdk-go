// Package multiplexer maps high-level search "systems" (e.g. "AV", "FC2")
// to the underlying SDK provider names.
package multiplexer

import (
	"sort"
	"strings"
)

// System is a named search system that maps to one or more SDK provider names.
type System struct {
	ID           string   `yaml:"id" json:"id"`
	Name         string   `yaml:"name" json:"name"`
	Providers    []string `yaml:"providers" json:"providers"`
	Disabled     bool     `yaml:"disabled,omitempty" json:"disabled"`
	DisabledHint string   `yaml:"disabled_hint,omitempty" json:"disabled_hint,omitempty"`
}

// Multiplexer routes a system ID to its providers.
type Multiplexer struct {
	systems map[string]System
	order   []string
}

// Default returns the built-in default system set.
// Users can override these via config.yaml; unconfigured systems remain available.
func Default() *Multiplexer {
	return &Multiplexer{
		systems: map[string]System{
			"AV": {
				ID:        "AV",
				Name:      "AV",
				Providers: []string{"JavBus"},
			},
			"FC2": {
				ID:        "FC2",
				Name:      "FC2",
				Providers: []string{"FC2", "fc2hub"},
			},
			"FANZA": {
				ID:        "FANZA",
				Name:      "FANZA/DMM",
				Providers: []string{"FANZA"},
			},
			"Madou": {
				ID:           "Madou",
				Name:         "Madou",
				Providers:    []string{"MadouQu"},
				Disabled:     true,
				DisabledHint: "需要启用 MadouQu provider (设置 MT_MOVIE_PROVIDER_MADOUQU__PRIORITY=1000)",
			},
			"ThePornDB": {
				ID:           "ThePornDB",
				Name:         "ThePornDB",
				Providers:    []string{"ThePornDB Movie", "ThePornDB Scene"},
				Disabled:     true,
				DisabledHint: "需要配置 ThePornDB access token",
			},
		},
		order: []string{"AV", "FC2", "FANZA", "Madou", "ThePornDB"},
	}
}

// New builds a multiplexer with defaults, then layers user-provided overrides
// on top. Overrides only add/overwrite systems; the default ordering is preserved
// with new entries appended.
func New(overrides []System) *Multiplexer {
	m := Default()
	known := make(map[string]bool, len(m.order))
	for _, id := range m.order {
		known[id] = true
	}
	for _, s := range overrides {
		id := strings.TrimSpace(s.ID)
		if id == "" {
			continue
		}
		if !known[id] {
			m.order = append(m.order, id)
			known[id] = true
		}
		m.systems[id] = System{
			ID:           id,
			Name:         firstNonEmpty(s.Name, id),
			Providers:    normalizeProviders(s.Providers),
			Disabled:     s.Disabled,
			DisabledHint: s.DisabledHint,
		}
	}
	return m
}

// Systems returns all known systems in display order.
func (m *Multiplexer) Systems() []System {
	out := make([]System, 0, len(m.order))
	for _, id := range m.order {
		if s, ok := m.systems[id]; ok {
			out = append(out, s)
		}
	}
	return out
}

// EnabledSystems returns the systems that are not marked disabled.
func (m *Multiplexer) EnabledSystems() []System {
	out := make([]System, 0, len(m.order))
	for _, id := range m.order {
		if s, ok := m.systems[id]; ok && !s.Disabled {
			out = append(out, s)
		}
	}
	return out
}

// ProvidersFor returns the provider names for a given system ID,
// or nil if the system is unknown.
func (m *Multiplexer) ProvidersFor(systemID string) []string {
	s, ok := m.systems[strings.TrimSpace(systemID)]
	if !ok {
		return nil
	}
	return s.Providers
}

// IsKnown reports whether the system ID exists.
func (m *Multiplexer) IsKnown(systemID string) bool {
	_, ok := m.systems[strings.TrimSpace(systemID)]
	return ok
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func normalizeProviders(providers []string) []string {
	if len(providers) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(providers))
	out := make([]string, 0, len(providers))
	for _, p := range providers {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil
	}
	sort.Strings(out)
	return out
}
