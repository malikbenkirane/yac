package snake

import "testing"

func TestSnake(t *testing.T) {
	for _, test := range []struct {
		expected string
		input    string
	}{
		{"other", "Other"},
		{"known-issue", "KnownIssue"},
	} {
		got := Case(test.input)
		if got != test.expected {
			t.Logf("expected %q got %q", test.expected, got)
			t.Fail()
		}
	}
}

func TestSnakeSpaceKeepCase(t *testing.T) {
	for _, test := range []struct {
		expected string
		input    string
	}{
		{"Other", "Other"},
		{"Known Issue", "KnownIssue"},
	} {
		got := Case(test.input, WithKeepCase(), WithSpace())
		if got != test.expected {
			t.Logf("expected %q got %q", test.expected, got)
			t.Fail()
		}
	}
}

func TestSnakePlural(t *testing.T) {
	for _, test := range []struct {
		input    string
		expected string
	}{
		{"KnownIssue", "Known Issues"},
		{"cat", "cats"},
		{"book", "books"},
		{"apple", "apples"},
		{"bus", "buses"},
		{"glass", "glasses"},
		{"wish", "wishes"},
		{"church", "churches"},
		{"box", "boxes"},
		{"quiz", "quizzes"},
		{"baby", "babies"},
		{"city", "cities"},
		{"story", "stories"},
		{"day", "days"},
		{"key", "keys"},
		{"boy", "boys"},
		{"other", "other"},
		{"Other", "Other"},
		{"MultipleOther", "Multiple Other"},
		{"ErrorMessage", "Error Messages"},
	} {
		got := Case(test.input, WithPlural(), WithKeepCase(), WithSpace(), WithPlural("other"))
		if got != test.expected {
			t.Logf("expected %q got %q", test.expected, got)
			t.Fail()
		}
	}
}
