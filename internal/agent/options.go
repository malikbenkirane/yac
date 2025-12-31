package agent

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/4sp1/yac/internal/commit/scope"
	"github.com/4sp1/yac/internal/commit/wip"
	"github.com/4sp1/yac/internal/snake"

	"go.uber.org/zap"
)

type Option interface {
	Apply(AgentContext) (AgentContext, error)
	fmt.Stringer
}

var ErrNoDiff = errors.New("no diff")

type option struct {
	description string
	apply       optionFn
}

var _ Option = &option{}

func WithLogger(log *zap.Logger) Option {
	return &option{
		apply: func(ac AgentContext) (AgentContext, error) {
			ac.logger = log
			return ac, nil
		},
		description: "set logger",
	}
}

func WithGitDiff() Option {
	return &option{
		apply: func(ac AgentContext) (AgentContext, error) {
			comb, err := exec.Command("git", "diff", "--cached", "-u").CombinedOutput()
			if err != nil {
				return ac, err
			}
			if len(comb) == 0 {
				return ac, ErrNoDiff
			}
			var status string
			{
				comb, err := exec.Command("git", "status", "-uno", "-s").CombinedOutput()
				if err != nil {
					return ac, err
				}
				var in, out bytes.Buffer
				if _, err := in.Write(comb); err != nil {
					return ac, err
				}
				cmd := exec.Command("sed", "-n", "/^[^ ]/p")
				cmd.Stdin = &in
				cmd.Stdout = &out
				if err := cmd.Run(); err != nil {
					return ac, err
				}
				status = out.String()
			}
			ac.diff = fmt.Sprintf("%s\n\n----------------\n\ngit status -s\n\n%s",
				string(comb), status)
			return ac, nil
		},
		description: "git diff",
	}
}

func WithJujutsuGitDiff() Option {
	return &option{
		apply: func(ac AgentContext) (AgentContext, error) {
			var stdout, stderr bytes.Buffer
			cmd := exec.Command("jj", "diff", "--git")
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				if _, err := io.Copy(os.Stdout, &stdout); err != nil {
					return ac, err
				}

				if _, err := io.Copy(os.Stderr, &stderr); err != nil {
					return ac, err
				}
				return ac, err
			}
			if len(stdout.String()) == 0 {
				return ac, ErrNoDiff
			}
			var status string
			{
				comb, err := exec.Command("jj", "st").CombinedOutput()
				if err != nil {
					return ac, err
				}
				status = string(comb)
			}
			ac.diff = fmt.Sprintf("%s\n\n----------------\n\ngit status -s\n\n%s",
				stdout.String(), status)
			return ac, nil
		},
		description: "jj diff --git",
	}
}

func WithScope(sc scope.Scope) Option {
	return &option{
		apply: func(ac AgentContext) (AgentContext, error) {
			if sc == scope.Other {
				ac.scope = ""
			} else {
				ac.scope = snake.Case(sc.String())
			}
			return ac, nil
		},
		description: fmt.Sprintf("set scope %q", sc),
	}
}

func WithGitLog(hash string) Option {
	return &option{
		apply: func(ac AgentContext) (AgentContext, error) {
			comb, err := exec.Command("git", "log", "-1", hash).CombinedOutput()
			if err != nil {
				ac.logger.Error("git log", zap.String("hash", hash), zap.Error(err))
				return ac, err
			}
			ac.logs = append(ac.logs, string(comb))
			return ac, nil
		},
		description: fmt.Sprintf("git log %q", hash),
	}
}

// WithJujutsuLog configure the agent to take in consideration a [jj show] output
// [id] can be either a commit id prefix or a change id prefix.
func WithJujutsuLog(id string) Option {
	return &option{
		apply: func(ac AgentContext) (AgentContext, error) {
			o, err := exec.Command("jj", "show", id).CombinedOutput()
			if err != nil {
				return ac, fmt.Errorf("exec: %w", err)
			}
			ac.logs = append(ac.logs, string(o))
			return ac, nil
		},
		description: fmt.Sprintf("jj show %q", id),
	}
}

func WithNote(note string, kind wip.Context) Option {
	return &option{
		apply: func(ac AgentContext) (AgentContext, error) {
			if _, ok := ac.wip[kind]; !ok {
				ac.wip[kind] = []string{}
			}
			ac.wip[kind] = append(ac.wip[kind], note)
			return ac, nil
		},
		description: "wip note",
	}
}
