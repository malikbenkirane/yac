package wip

import "testing"

func TestCase(t *testing.T) {
	for _, test := range []struct {
		input          Context
		expectedHeader string
		expectedFlag   string
	}{
		// {Blocker, "Blockers", "wip-blocker"},
		// {Configuration, "Configurations", "wip-configuration"},
		// {DesignReview, "Design Reviews", "wip-design-review"},
		// {DevDocumentation, "Dev Documentation", "wip-dev-documentation"},
		// {ElaboratingTest, "Elaborating Tests", "wip-elaborating-test"},
		// {ErrorMessage, "Error Messages", "wip-error-message"},
		// {ErrorDebugging, "Error Debugging", "wip-error-debugging"},
		// {FixAttempt, "Fix Attempts", "wip-fix-attempt"},
		// {Implementation, "Implementations", "wip-implementation"},
		// {Integration, "Integrations", "wip-integration"},
		// {KnownIssue, "Known Issues", "wip-known-issue"},
		// {NeedTesting, "Need Testing", "wip-need-testing"},
		// {OptimizationGuess, "Optimization Guesses", "wip-optimization-guess"},
		// {OptimizationInsight, "Optimization Insights", "wip-optimization-insight"},
		// {Other, "Other", "wip-other"},
		// {PlannedFix, "Planned Fixes", "wip-planned-fix"},
		// {PlannedImprovement, "Planned Improvements", "wip-planned-improvement"},
		// {PlannedIntegration, "Planned Integrations", "wip-planned-integration"},
		// {PlannedOptimization, "Planned Optimizations", "wip-planned-optimization"},
		// {PlannedRefactoring, "Planned Refactoring", "wip-planned-refactoring"},
		// {Prototype, "Prototypes", "wip-prototype"},
		// {Refactoring, "Refactoring", "wip-refactoring"},
		// {RefactoringIdea, "Refactoring Ideas", "wip-refactoring-idea"},
		// {SeparationOfConcern, "Separation Of Concerns", "wip-separation-of-concern"},
		// {TemporaryWorkaround, "Temporary Workarounds", "wip-temporary-workaround"},
		// {Test, "Tests", "wip-test"},
		// {TestRefinement, "Test Refinements", "wip-test-refinement"},
		// {Todo, "Todos", "wip-todo"},
		// {UnitTest, "Unit Tests", "wip-unit-test"},
		// {UpdateDependency, "Update Dependencies", "wip-update-dependency"},
		// {UserDocumentation, "User Documentation", "wip-user-documentation"},
		// {UserExperienceTest, "User Experience Tests", "wip-user-experience-test"},
		// {UserInterfaceTest, "User Interface Tests", "wip-user-interface-test"},
	} {
		gotHeader := test.input.Header()
		gotFlag := test.input.Flag()
		if gotHeader != test.expectedHeader {
			t.Logf("got header %q expected %q", gotHeader, test.expectedHeader)
			t.Fail()
		}
		if gotFlag != test.expectedFlag {
			t.Logf("got flag %q expected %q", gotFlag, test.expectedFlag)
			t.Fail()
		}
	}
}
