package avokadostringutils

import (
	"strings"
	"unicode"

	"github.com/bilustek/avokado/avokadoerror"
)

// ToSnakeCase converts a string to snake_case.
// It handles camelCase, PascalCase, spaces, and hyphens as word separators.
// Consecutive separators are collapsed into a single underscore.
func ToSnakeCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	var result strings.Builder

	for i, r := range s {
		if r == ' ' || r == '-' {
			if result.Len() > 0 && !strings.HasSuffix(result.String(), "_") {
				result.WriteRune('_')
			}

			continue
		}

		if unicode.IsUpper(r) {
			if i > 0 {
				prev := rune(s[i-1])
				if prev != '_' && prev != ' ' && prev != '-' && !unicode.IsUpper(prev) {
					result.WriteRune('_')
				}
			}

			result.WriteRune(unicode.ToLower(r))

			continue
		}

		result.WriteRune(r)
	}

	out := strings.ToLower(result.String())
	out = strings.TrimFunc(out, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	return out
}

// ExtractFlag extracts a named flag value from a list of command-line arguments.
// It supports both "--flag value" and "--flag=value" formats.
// Returns an error if the flag is missing, has no value, or has an empty value.
func ExtractFlag(args []string, flagName string) (string, error) {
	lookup := "--" + flagName
	lookupWithEq := lookup + "="

	for i, arg := range args {
		if arg == lookup {
			if i+1 >= len(args) {
				return "", avokadoerror.New(lookup + " flag requires a value")
			}
			found := args[i+1]
			if found == "" {
				return "", avokadoerror.New(lookup + " value cannot be empty")
			}
			return found, nil
		}

		if strings.HasPrefix(arg, lookupWithEq) {
			found := strings.TrimPrefix(arg, lookupWithEq)
			if found == "" {
				return "", avokadoerror.New(lookupWithEq + " value cannot be empty")
			}
			return found, nil
		}
	}

	return "", avokadoerror.New(lookup + " or " + lookupWithEq + " flag is required")
}
