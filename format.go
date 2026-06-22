package deepstack

import (
	"fmt"
	"strconv"
	"strings"
)

func formatDisplayValue(value any) string {
	if s, ok := value.(string); ok {
		if s == "" || strings.ContainsAny(s, " \t\n\r") {
			return strconv.Quote(s)
		}
		return s
	}
	return fmt.Sprint(value)
}
