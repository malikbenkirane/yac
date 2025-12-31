package scope

import (
	"fmt"

	"github.com/4sp1/yac/internal/snake"
)

//go:generate stringer -type=Scope
type Scope int

const (
	Other Scope = iota
	Accessibility
	Alpha
	Analytics
	Android
	Api
	Auth
	Backend
	Benchmark
	Beta
	Build
	Cd
	Chore
	Ci
	Cli
	Compliance
	Components
	Config
	Database
	Db
	Dependencies
	Deps
	Devops
	Docs
	Drizzle
	E2e
	Feat
	Fix
	Infra
	Internationalization
	Ios
	Kubernetes
	Kustomize
	Legal
	Localization
	Logging
	Mobile
	Perf
	Plumbing
	PreRelease
	Prisma
	Profiling
	Refactor
	Release
	ReleaseCandidate
	Scripts
	SearchEngineOptimization
	Security
	Server
	Testing
	Tests
	Tracking
	Translation
	Types
	Ui
	UnitTests
	Ux
	Vendor
	UpperBound // only use in for loop
)

func (s Scope) Flag() string {
	return fmt.Sprintf("scope-%s", snake.Case(s.String()))
}

func (s Scope) Label() string {
	return fmt.Sprintf("%q related commit", snake.Case(s.String(), snake.WithSpace()))
}
