package wip

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/malikbenkirane/yac/internal/snake"
	"gopkg.in/yaml.v3"
)

//go:generate stringer -type=Context
type Context int

const (
	Other         Context = iota
	Blocker               // cannot proceed without resolving these
	InProgress            // active work, partially complete
	NeedsDecision         // design/architecture choices blocking progress
	Idea                  // future improvements, non-critical
	Question              // need input from team/community
	FixAttempt            // unsure about fix effects
	// Implementation
	ImplementationTodo
	ImplementationParial
	ImplementationNeedsReview
	// Tesing
	Testing
	TestingUntested
	TestingFailing
	TestingMissingCoverage
	TestingIntegrationNeeded
	TestingMissingUnitTest
	TestingMissingEndToEndTest
	TestingStress
	TestingUserInterface
	TestingUserExperience
	// Benchmark
	Benchmark
	BenchmarkIntegration
	BenchmarkCost
	// TechnicalDebt
	TechincalDebt
	TechnicalDebtWorkaround
	TechnicalDebtRefactoringNeeded
	TechnicalDebtOptimizationOpportunity
	// Dependencies
	Dependency
	DependenciesExternalBlocker
	DependenciesVeresionUpdate
	DependenciesIntegrationPoint
	// Documentation
	Documentation
	DocumentationCodeComment
	DocumentationApiDoc
	DocumentationUserDoc
	DocumentationMigrationGuide
	DoucmentationBuildGuide
	DocumentationDesignDocument
	DocumentationInfrastructureDocument
	DocumentationArchitectureDocument
	// KnownIssue
	KnownIssue
	KnownIssueError
	KnownIssueUpdate
	UpperBound // only for range upper bound in for loop
)

func (c Context) Flag() string {
	return "wip-" + snake.Case(c.String())
}

func (c Context) Header() string {
	return snake.Case(c.String(),
		snake.WithKeepCase(),
		snake.WithSpace(),
		snake.WithPlural(
			"other",
			"further",
			"documentation",
		))
}

var _ json.Unmarshaler = &Wrap{}
var _ yaml.Unmarshaler = &Wrap{}

type Wrap struct {
	M     map[Context][]string
	index map[string]Context
}

func NewWrap() Wrap {
	index := map[string]Context{}
	m := make(map[Context][]string)
	for i := Other; i < UpperBound; i++ {
		index[i.String()] = i
		m[i] = []string{}
	}
	return Wrap{
		M:     m,
		index: index,
	}
}

func (c *Wrap) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	for section, notesInterface := range m {
		notes, ok := notesInterface.([]interface{})
		if !ok {
			return fmt.Errorf("%T not []string", notesInterface)
		}
		i, ok := c.index[section]
		if !ok {
			return fmt.Errorf("%w %q", ErrUnknownContext, section)
		}
		c.M[i] = make([]string, len(notes))
		for k, noteInterface := range notes {
			note, ok := noteInterface.(string)
			if !ok {
				return fmt.Errorf("%T not string", noteInterface)
			}
			c.M[i][k] = note
		}
	}
	return nil
}

func (c *Wrap) UnmarshalYAML(n *yaml.Node) error {
	var m map[string]interface{}
	if err := n.Decode(&m); err != nil {
		return err
	}
	for section, notesInterface := range m {
		notes, ok := notesInterface.([]interface{})
		if !ok {
			return fmt.Errorf("%T not []string", notesInterface)
		}
		i, ok := c.index[section]
		if !ok {
			return fmt.Errorf("%w %q", ErrUnknownContext, section)
		}
		c.M[i] = make([]string, len(notes))
		for k, noteInterface := range notes {
			note, ok := noteInterface.(string)
			if !ok {
				return fmt.Errorf("%T not string", noteInterface)
			}
			c.M[i][k] = note
		}
	}
	return nil
}

var ErrUnknownContext = errors.New("unknown wip context")
