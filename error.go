package deepstack

import (
	"fmt"
	"strings"
)

type DeepStackError struct {
	Message    string
	StackTrace string
	Context    map[string]any
}

// Error returns all fields, so the logs clearly show when an error has been wrapped incorrectly multiple times by different layers.
func (d *DeepStackError) Error() string {
	return fmt.Sprintf("message: %s; context: %s; stack: %s", d.Message, formatContextEntries(d.Context), d.StackTrace)
}

func formatContextEntries(context map[string]any) string {
	entries := sortedContextEntries(context)
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		parts = append(parts, fmt.Sprintf("%s=%s", entry.key, formatDisplayValue(entry.value)))
	}
	return "[" + strings.Join(parts, " ") + "]"
}
