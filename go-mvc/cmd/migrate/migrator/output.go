package migrator

import (
	"fmt"
	"strings"
)

// MigrationStatus represents status of a single migration
type MigrationStatus struct {
	Ran       bool
	Migration string
	Batch     *int
}

// RenderStatusTable formats migration statuses into an aligned ASCII table like Laravel
func RenderStatusTable(statuses []MigrationStatus) string {
	if len(statuses) == 0 {
		return "No migrations found."
	}

	maxNameLen := len("Migration")
	for _, s := range statuses {
		if len(s.Migration) > maxNameLen {
			maxNameLen = len(s.Migration)
		}
	}

	border := fmt.Sprintf("+------+-%s-+---------+", strings.Repeat("-", maxNameLen))
	header := fmt.Sprintf("| Ran? | %-*s | Batch   |", maxNameLen, "Migration")

	var sb strings.Builder
	sb.WriteString(border + "\n")
	sb.WriteString(header + "\n")
	sb.WriteString(border + "\n")

	for _, s := range statuses {
		ranStr := "No"
		batchStr := "Pending"
		if s.Ran {
			ranStr = "Yes"
			if s.Batch != nil {
				batchStr = fmt.Sprintf("%d", *s.Batch)
			}
		}
		sb.WriteString(fmt.Sprintf("| %-4s | %-*s | %-7s |\n", ranStr, maxNameLen, s.Migration, batchStr))
	}

	sb.WriteString(border)
	return sb.String()
}
