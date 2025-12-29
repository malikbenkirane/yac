package config

import (
	"github.com/malikbenkirane/yac/internal/commit/scope"
	"github.com/malikbenkirane/yac/internal/commit/wip"

	"gopkg.in/yaml.v3"
)

func NewYaml(s scope.Scope) ConfigYAML {
	return &configYAML{
		Wip:  wip.NewWrap(),
		Logs: []string{},
	}
}

func FromYAML(c Config) ConfigYAML {
	return &configYAML{
		Wip:  wip.Wrap{M: c.Wip},
		Logs: c.Logs,
	}
}

type ConfigYAML = *configYAML

func (f *configYAML) FlagsLogs() []string                { return f.Logs }
func (f *configYAML) FlagsWip() map[wip.Context][]string { return f.Wip.M }

type configYAML struct {
	Wip  wip.Wrap `yaml:"wip_context"`
	Logs []string `yaml:"logs"`
}

var _ yaml.Marshaler = &configYAML{}

func (f *configYAML) MarshalYAML() (interface{}, error) {
	m := make(map[string][]string)
	for section, notes := range f.Wip.M {
		m[section.String()] = notes
	}
	v := struct {
		Wip  map[string][]string `yaml:"wip_context"`
		Logs []string            `yaml:"logs"`
	}{
		Wip:  m,
		Logs: f.Logs,
	}
	return v, nil
}

const DefaultYamlFile = ".prepare.yaml"
