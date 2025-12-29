package config

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/malikbenkirane/yac/internal/commit/wip"
)

type Flags interface {
	FlagsLogs() []string
	FlagsWip() map[wip.Context][]string
}

var _ Flags = &configJSON{}
var _ Flags = &configYAML{}

//go:generate stringer -type=Mode
type Mode int

const (
	ModeJSON Mode = iota
	ModeYAML
	ModeNone
)

type Config = *config

type config struct {
	Wip  map[wip.Context][]string
	Logs []string
}

var _ yaml.Unmarshaler = &config{}
var _ json.Unmarshaler = &config{}

func (f *config) UnmarshalYAML(n *yaml.Node) error {
	v := configYAML{
		Wip: wip.NewWrap(),
	}
	if err := n.Decode(&v); err != nil {
		return fmt.Errorf("unmarshal flag json: %w", err)
	}
	*f = config{
		Wip:  v.Wip.M,
		Logs: v.Logs,
	}
	return nil
}

func (f *config) UnmarshalJSON(data []byte) error {
	v := configJSON{
		Wip: wip.NewWrap(),
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("unmarshal flag json: %w", err)
	}
	*f = config{
		Wip:  v.Wip.M,
		Logs: v.Logs,
	}
	return nil
}

func (m Mode) File(path string) string {
	if m == ModeYAML {
		return path + ".flags.yaml"
	}
	return path + ".flags.json"
}
