package config

import (
	"encoding/json"

	"github.com/malikbenkirane/yac/internal/commit/scope"
	"github.com/malikbenkirane/yac/internal/commit/wip"
)

func NewJson(s scope.Scope) ConfigJSON {
	return &configJSON{
		Wip:             wip.NewWrap(),
		Logs:            []string{},
		Scope:           s.String(),
		scope:           s,
		AvailableScopes: availableScopes(),
	}
}

func FromJSON(c Config) ConfigJSON {
	return &configJSON{
		Wip:   wip.Wrap{M: c.Wip},
		Logs:  c.Logs,
		Scope: c.Scope,
		scope: c.scope,
	}
}

type ConfigJSON = *configJSON

func (f *configJSON) FlagsLogs() []string                { return f.Logs }
func (f *configJSON) FlagsWip() map[wip.Context][]string { return f.Wip.M }
func (f *configJSON) FlagScope() scope.Scope             { return f.scope }

type configJSON struct {
	Wip             wip.Wrap `yaml:"wip_context"`
	Logs            []string `yaml:"logs"`
	Scope           string   `yaml:"scope"`
	AvailableScopes []string
	scope           scope.Scope
}

var _ json.Marshaler = &configJSON{}

func (f *configJSON) MarshalJSON() ([]byte, error) {
	m := make(map[string][]string)
	for section, notes := range f.Wip.M {
		m[section.String()] = notes
	}
	v := struct {
		Wip             map[string][]string
		Logs            []string
		Scope           string
		AvailableScopes []string `json:"_available_scopes"`
	}{
		Wip:             m,
		Logs:            f.Logs,
		Scope:           f.Scope,
		AvailableScopes: f.AvailableScopes,
	}
	return json.Marshal(v)
}

const DefaultJsonFile = ".prepare.json"
