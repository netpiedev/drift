package verify

import (
	"fmt"

	"github.com/netpiedev/drift/core/internal/migration/parser"
)

type ValidationIssue struct {
	Severity string
	Message  string
}

func ValidateOrdered(migrations []parser.Migration) []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	seen := make(map[string]struct{})
	lastVersion := ""
	for _, m := range migrations {
		k := m.Version + ":" + string(m.Direction)
		if _, ok := seen[k]; ok {
			issues = append(issues, ValidationIssue{Severity: "error", Message: fmt.Sprintf("duplicate migration: %s", k)})
		}
		seen[k] = struct{}{}

		if lastVersion != "" && m.Version < lastVersion {
			issues = append(issues, ValidationIssue{Severity: "error", Message: fmt.Sprintf("out-of-order migration: %s after %s", m.Version, lastVersion)})
		}
		if m.Direction == parser.DirectionUp {
			lastVersion = m.Version
		}
	}
	return issues
}

func ValidateChecksums(migrations []parser.Migration, applied []parser.AppliedMigration) []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	fileByKey := make(map[string]parser.Migration)
	for _, m := range migrations {
		fileByKey[m.Version+":"+string(m.Direction)] = m
	}
	for _, a := range applied {
		k := a.Version + ":" + a.Direction
		current, ok := fileByKey[k]
		if !ok {
			issues = append(issues, ValidationIssue{Severity: "warn", Message: fmt.Sprintf("applied migration missing from disk: %s", k)})
			continue
		}
		if current.Checksum != a.Checksum {
			issues = append(issues, ValidationIssue{Severity: "error", Message: fmt.Sprintf("checksum mismatch for %s", k)})
		}
	}
	return issues
}
