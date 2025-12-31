package agent

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/4sp1/yac/internal/commit/wip"
)

func prepareExpected(t *testing.T) []string {
	t.Helper()
	f, err := os.Open("user_expected.txt")
	if err != nil {
		t.Fatalf("open user_expected: %s", err.Error())
	}
	var b bytes.Buffer
	if _, err = io.Copy(&b, f); err != nil {
		t.Fatalf("copy user_expected: %s", err.Error())
	}
	return strings.Split(b.String(), "\n---\n")
}

func TestFill(t *testing.T) {
	expectedSet := prepareExpected(t)
	for i, tf := range []templateFiller{
		{
			Scope:       "",
			GitLog:      "",
			Diff:        "some diff",
			IdealFuture: nil,
		},
		{
			Scope:       "api",
			GitLog:      "",
			Diff:        "some diff",
			IdealFuture: []idealSection{},
		},
		{
			Scope:  "api",
			GitLog: "some log\n\nand other logs",
			Diff:   "some diff",
			IdealFuture: []idealSection{
				{Name: "Other", Notes: []string{"Wonderful", "Things"}},
				{Name: "Known Issues", Notes: []string{"We acknowledge that it is unfortunate"}},
			},
		},
	} {
		t.Run(fmt.Sprintf("golden file section %d", i), func(t *testing.T) {
			var b bytes.Buffer
			tf.IdealSeparator = DefaultIdealSeparator
			if err := tf.Fill(&b); err != nil {
				t.Fatalf("templateFiller:  %s", err.Error())
			}
			expected := strings.TrimSpace(expectedSet[i])
			got := strings.TrimSpace(b.String())
			if got != expected {
				t.Fatalf("expected\n%q\ngot\n%q", expected, got)
			}
		})
	}
}

func TestAgentFill(t *testing.T) {
	expectedSet := prepareExpected(t)
	for i, ctx := range []AgentContext{
		{
			diff: "some diff",
		},
		{
			diff:  "some diff",
			scope: "api",
		},
		{
			diff:  "some diff",
			scope: "api",
			logs: []string{
				"some log",
				"and other logs",
			},
			wip: map[wip.Context][]string{
				wip.Other:      {"Wonderful", "Things"},
				wip.KnownIssue: {"We acknowledge that it is unfortunate"},
			},
		},
	} {
		t.Run(fmt.Sprintf("golden file section %d", i), func(t *testing.T) {
			user, err := agent{context: ctx}.UserPrompt()
			if err != nil {
				t.Fatalf("agent: user prompt: %s", err.Error())
			}
			got := strings.TrimSpace(user)
			expected := strings.TrimSpace(expectedSet[i])
			if got != expected {
				t.Fatalf("expected\n%q\ngot\n%q", expected, got)
			}
		})
	}
}
