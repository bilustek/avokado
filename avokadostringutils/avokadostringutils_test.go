package avokadostringutils_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/bilustek/avokado/avokadostringutils"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty string", input: "", expected: ""},
		{name: "whitespace only", input: "   ", expected: ""},
		{name: "already snake_case", input: "hello_world", expected: "hello_world"},
		{name: "camelCase", input: "helloWorld", expected: "hello_world"},
		{name: "PascalCase", input: "HelloWorld", expected: "hello_world"},
		{name: "spaces", input: "hello world", expected: "hello_world"},
		{name: "hyphens", input: "hello-world", expected: "hello_world"},
		{name: "mixed separators", input: "hello-world foo", expected: "hello_world_foo"},
		{name: "single word lowercase", input: "hello", expected: "hello"},
		{name: "single word uppercase", input: "Hello", expected: "hello"},
		{name: "multiple uppercase", input: "HTMLParser", expected: "htmlparser"},
		{name: "uppercase to lower transition", input: "userID", expected: "user_id"},
		{name: "leading spaces trimmed", input: "  HelloWorld  ", expected: "hello_world"},
		{name: "single char", input: "A", expected: "a"},
		{name: "all lowercase", input: "alllowercase", expected: "alllowercase"},
		{name: "consecutive underscores preserved", input: "hello__world", expected: "hello__world"},
		{name: "multiple spaces collapsed", input: "Foo  bar", expected: "foo_bar"},
		{name: "many spaces collapsed", input: "Foo    Bar", expected: "foo_bar"},
		{name: "multiple hyphens collapsed", input: "foo--bar", expected: "foo_bar"},
		{name: "mixed consecutive separators", input: "foo - bar", expected: "foo_bar"},
		{name: "trailing separators trimmed", input: "Foo   - Bar     Baz-", expected: "foo_bar_baz"},
		{name: "leading separators trimmed", input: "--foo", expected: "foo"},
		{name: "both ends trimmed", input: "- -foo bar- -", expected: "foo_bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := avokadostringutils.ToSnakeCase(tt.input)
			if got != tt.expected {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractFlag(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		flagName  string
		expected  string
		wantError bool
	}{
		{name: "space separated", args: []string{"--name", "foo"}, flagName: "name", expected: "foo"},
		{name: "equals separated", args: []string{"--name=foo"}, flagName: "name", expected: "foo"},
		{
			name:     "flag among others",
			args:     []string{"--verbose", "--name", "foo", "--debug"},
			flagName: "name",
			expected: "foo",
		},
		{
			name:     "equals among others",
			args:     []string{"--verbose", "--name=foo", "--debug"},
			flagName: "name",
			expected: "foo",
		},
		{name: "missing flag", args: []string{"--other", "value"}, flagName: "name", wantError: true},
		{name: "empty args", args: []string{}, flagName: "name", wantError: true},
		{name: "nil args", args: nil, flagName: "name", wantError: true},
		{name: "flag at end without value", args: []string{"--name"}, flagName: "name", wantError: true},
		{name: "empty value space separated", args: []string{"--name", ""}, flagName: "name", wantError: true},
		{name: "empty value equals", args: []string{"--name="}, flagName: "name", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := avokadostringutils.ExtractFlag(tt.args, tt.flagName)

			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil with value %q", got)
				}

				var avErr *avokadoerror.Error
				if !errors.As(err, &avErr) {
					t.Errorf("expected *avokadoerror.Error, got %T", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf("ExtractFlag(%v, %q) = %q, want %q", tt.args, tt.flagName, got, tt.expected)
			}
		})
	}
}

func ExampleExtractFlag() {
	args := []string{"--verbose", "--name", "avokado", "--debug"}

	val, err := avokadostringutils.ExtractFlag(args, "name")
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(val)
	// Output:
	// avokado
}

func ExampleExtractFlag_equals() {
	args := []string{"--name=avokado"}

	val, err := avokadostringutils.ExtractFlag(args, "name")
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(val)
	// Output:
	// avokado
}

func ExampleToSnakeCase() {
	fmt.Println(avokadostringutils.ToSnakeCase("helloWorld"))
	fmt.Println(avokadostringutils.ToSnakeCase("HelloWorld"))
	fmt.Println(avokadostringutils.ToSnakeCase("hello-world"))
	fmt.Println(avokadostringutils.ToSnakeCase("hello world"))
	fmt.Println(avokadostringutils.ToSnakeCase("userID"))
	// Output:
	// hello_world
	// hello_world
	// hello_world
	// hello_world
	// user_id
}
