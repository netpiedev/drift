package lint

import (
	"fmt"
	"strings"

	"github.com/netpiedev/drift/core/internal/migration/analyzer"
	"github.com/netpiedev/drift/core/internal/migration/parser"
)

type Issue struct {
	Severity string
	Message  string
	File     string
}

func Analyze(migrations []parser.Migration) []Issue {
	issues := make([]Issue, 0)
	for _, m := range migrations {
		if m.Ext != "sql" || m.Direction != parser.DirectionUp {
			continue
		}
		report := analyzer.AnalyzeSQL(string(m.Content))
		for _, w := range report.Warnings {
			issues = append(issues, Issue{Severity: "warn", Message: w.Message, File: m.Path})
		}
		if strings.Contains(strings.ToLower(string(m.Content)), "select *") {
			issues = append(issues, Issue{Severity: "warn", Message: "SELECT * detected in migration", File: m.Path})
		}
		if strings.Contains(strings.ToLower(string(m.Content)), "drop table") {
			issues = append(issues, Issue{Severity: "error", Message: "DROP TABLE requires manual approval", File: m.Path})
		}
	}
	return issues
}

func FormatIssues(issues []Issue) string {
	if len(issues) == 0 {
		return "no lint issues"
	}
	lines := make([]string, 0, len(issues))
	for _, i := range issues {
		lines = append(lines, fmt.Sprintf("[%s] %s (%s)", i.Severity, i.Message, i.File))
	}
	return strings.Join(lines, "\n")
}
