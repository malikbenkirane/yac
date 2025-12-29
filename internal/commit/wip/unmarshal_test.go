package wip

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	j := []byte(`
	{
		"KnownIssue": ["some issue"]
	}
	`)
	m := NewWrap()
	if err := json.Unmarshal(j, &m); err != nil {
		t.Fatalf("Unmarshal: %s", err)
	}
	notes, ok := m.M[KnownIssue]
	if !ok {
		t.Fatal("expected KnownIssue key")
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note got %d", len(notes))
	}
	if notes[0] != "some issue" {
		t.Fatalf("expected note \"some issue\" got %q", notes[0])
	}

	for _, test := range []struct {
		name          string
		input         string
		expected      map[Context][]string
		expectedError error
	}{
		{
			name: "single note",
			input: `
{
	"KnownIssue": ["some issue"]
}
		`,
			expected: map[Context][]string{
				KnownIssue: {"some issue"},
			},
		},
		{
			name: "two notes one context",
			input: `
{
	"KnownIssue": ["some issue", "other issue"]
}
		`,
			expected: map[Context][]string{
				KnownIssue: {"some issue", "other issue"},
			},
		},
		{
			name: "unknown context",
			input: `
{
	"KnownIssue": ["some issue", "other issue"],
	"Unknown": ["other issue"]
}
		`,
			expectedError: ErrUnknownContext,
		},
		{
			name: "two contexts",
			input: `
{
	"KnownIssue": ["some issue", "other issue"],
	"Testing": ["some missing test"]
}
		`,
			expected: map[Context][]string{
				KnownIssue: {"some issue", "other issue"},
				Testing:    {"some missing test"},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			w := NewWrap()
			if err := json.Unmarshal([]byte(test.input), &w); err != nil {
				if errors.Is(err, test.expectedError) {
					return
				}
				t.Fatalf("unmarshal error: %s", err.Error())
			}
			unexpectedContexts := []Context{}
			for i := Other; i < UpperBound; i++ {
				if _, ok := test.expected[i]; !ok {
					unexpectedContexts = append(unexpectedContexts, i)
				}
			}
			for _, i := range unexpectedContexts {
				if _, ok := w.M[i]; ok {
					t.Fatalf("unexpected context %q", i)
				}
			}
			for i, expectedNotes := range test.expected {
				notes, ok := w.M[i]
				if !ok {
					t.Fatalf("expected key %q", i)
				}
				if len(notes) != len(expectedNotes) {
					t.Fatalf("expected notes %s got %s",
						fmtStringList(expectedNotes), fmtStringList(notes))
				}
				notesMap := map[string]struct{}{}
				for _, note := range notes {
					notesMap[note] = struct{}{}
				}
				expectedNotesMap := map[string]struct{}{}
				for _, note := range expectedNotes {
					if _, ok := notesMap[note]; !ok {
						t.Logf("expected note %q", note)
						t.Fail()
					}
					expectedNotesMap[note] = struct{}{}
				}
				for _, note := range notes {
					if _, ok := expectedNotesMap[note]; !ok {
						t.Logf("unexpected note %q", note)
						t.Fail()
					}
				}
			}
		})
	}
}

func fmtStringList(l []string) string {
	for i, s := range l {
		l[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(l, ", ") + "]"
}
