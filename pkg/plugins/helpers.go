package plugins

import (
	"fmt"
	"strings"
)

// ValidateAgainstDefinition checks that all required fields declared in def.Parameters.Required
// are present and non-empty in req.Arguments.
func ValidateAgainstDefinition(def ToolDefinition, req ExecuteRequest) *ValidateResponse {
	var errs []string
	for _, key := range def.Parameters.Required {
		v, exists := req.Arguments[key]
		if !exists {
			errs = append(errs, fmt.Sprintf("%q is required", key))
			continue
		}
		s, ok := v.(string)
		if !ok || s == "" {
			errs = append(errs, fmt.Sprintf("%q is required", key))
		}
	}
	return &ValidateResponse{Valid: len(errs) == 0, Errors: errs}
}

func MatchesToolsFilter(def ToolDefinition, f ListToolsFilter) bool {
	if def.Deprecated && !f.Deprecated {
		return false
	}
	if f.Prefix != "" && !strings.HasPrefix(def.Name, f.Prefix) {
		return false
	}
	if len(f.Tags) > 0 {
		for _, want := range f.Tags {
			for _, have := range def.Tags {
				if have == want {
					goto tagMatched
				}
			}
		}
		return false
	tagMatched:
	}
	return true
}
