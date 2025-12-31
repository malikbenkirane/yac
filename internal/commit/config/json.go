package config

import (
	"encoding/json"

	"github.com/4sp1/yac/internal/commit/scope"
	"github.com/4sp1/yac/internal/commit/wip"
)

func NewJson(s scope.Scope) ConfigJSON {
	return &configJSON{
		Wip:  wip.NewWrap(),
		Logs: []string{},
	}
}

func FromJSON(c Config) ConfigJSON {
	return &configJSON{
		Wip:  wip.Wrap{M: c.Wip},
		Logs: c.Logs,
	}
}

type ConfigJSON = *configJSON

func (f *configJSON) FlagsLogs() []string                { return f.Logs }
func (f *configJSON) FlagsWip() map[wip.Context][]string { return f.Wip.M }

type configJSON struct {
	Wip  wip.Wrap `yaml:"wip_context"`
	Logs []string `yaml:"logs"`
}

var _ json.Marshaler = &configJSON{}

func (f *configJSON) MarshalJSON() ([]byte, error) {
	m := make(map[string][]string)
	for section, notes := range f.Wip.M {
		m[section.String()] = notes
	}
	v := struct {
		Wip  map[string][]string
		Logs []string
	}{
		Wip:  m,
		Logs: f.Logs,
	}
	return json.Marshal(v)
}

const DefaultJsonFile = ".prepare.json"
