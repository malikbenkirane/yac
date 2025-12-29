package config

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/malikbenkirane/yac/internal/commit/scope"
	"github.com/malikbenkirane/yac/internal/commit/wip"
)

type Flags interface {
	FlagsLogs() []string
	FlagsWip() map[wip.Context][]string
	FlagScope() scope.Scope
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
	Wip   map[wip.Context][]string
	Logs  []string
	Scope string
	scope scope.Scope
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
	_scope := scope.Scope(-1)
	for i := scope.Other; i < scope.UpperBound; i++ {
		if i.String() == v.Scope {
			_scope = i
		}
	}
	if _scope == -1 {
		_scope = scope.Other
	}
	*f = config{
		Wip:   v.Wip.M,
		Logs:  v.Logs,
		Scope: _scope.String(),
		scope: _scope,
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
	_scope := scope.Scope(-1)
	for i := scope.Other; i < scope.UpperBound; i++ {
		if i.String() == v.Scope {
			_scope = i
		}
	}
	if _scope == -1 {
		_scope = scope.Other
	}
	*f = config{
		Wip:   v.Wip.M,
		Logs:  v.Logs,
		Scope: _scope.String(),
		scope: _scope,
	}
	return nil
}

func (m Mode) File(path string) string {
	if m == ModeYAML {
		return path + ".flags.yaml"
	}
	return path + ".flags.json"
}

func availableScopes() []string {
	scopt := make([]string, scope.UpperBound)
	for i := scope.Other; i < scope.UpperBound; i++ {
		scopt[i] = i.String()
	}
	return scopt
}
