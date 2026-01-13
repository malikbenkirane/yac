package agent

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/4sp1/yac/internal/commit/wip"

	"go.uber.org/zap"
)

func New(opts ...Option) (Agent, error) {
	ctx := AgentContext{
		wip: make(map[wip.Context][]string),
	}
	var err error
	for _, opt := range opts {
		ctx, err = opt.Apply(ctx)
		if err != nil {
			return nil, fmt.Errorf("option %q: %w", opt, err)
		}
	}
	return &agent{
		context: ctx,
	}, nil
}

type Agent interface {
	UserPrompt(f Filler) (string, error)
}

type AgentContext struct {
	diff  string
	scope string
	logs  []string
	wip   map[wip.Context][]string

	logger *zap.Logger
}

type agent struct {
	context AgentContext
}

func DefaultFiller(a agent) Filler {
	ideal := []idealSection{}
	for i := range wip.UpperBound {
		if notes, ok := a.context.wip[i]; ok {
			if len(notes) == 0 {
				continue
			}
			ideal = append(ideal, idealSection{
				Name:  i.Header(),
				Notes: notes,
			})
		}
	}
	return templateFiller{
		GitLog:         strings.TrimSpace(strings.Join(a.context.logs, "\n\n")),
		Scope:          a.context.scope,
		Diff:           a.context.diff,
		IdealFuture:    ideal,
		IdealSeparator: DefaultIdealSeparator,
	}
}

func (a agent) UserPrompt(f Filler) (string, error) {
	var b bytes.Buffer
	err := f.Fill(&b)
	if err != nil {
		return "", fmt.Errorf("templateFiller: %w", err)
	}
	return b.String(), nil
}

type optionFn func(AgentContext) (AgentContext, error)

func (opt option) Apply(ac AgentContext) (AgentContext, error) {
	return opt.apply(ac)
}

func (opt option) String() string {
	return opt.description
}
