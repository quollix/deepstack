package deepstack

import (
	"log/slog"
	"sort"
)

type Record struct {
	level      slog.Level
	msg        string
	attributes []slog.Attr
}

func (r *Record) AddAttrs(key string, value any) {
	r.attributes = append(r.attributes, slog.Any(key, value))
}

type contextEntry struct {
	key   string
	value any
}

func sortedContextEntries(context map[string]any) []contextEntry {
	entries := make([]contextEntry, 0, len(context))
	for key, value := range context {
		entries = append(entries, contextEntry{key: key, value: value})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})
	return entries
}
