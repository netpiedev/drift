package analyzer

import (
	"regexp"
	"strings"
)

type Warning struct {
	Code    string
	Message string
}

type Report struct {
	Warnings        []Warning
	AffectedTables  []string
	EstimatedLocks  []string
	TransactionalOK bool
}

var tableRegex = regexp.MustCompile(`(?i)(?:alter|create|drop)\s+table\s+(?:if\s+exists\s+|if\s+not\s+exists\s+)?([a-zA-Z0-9_\.\"]+)`)

func AnalyzeSQL(sqlText string) Report {
	text := strings.ToLower(sqlText)
	warnings := make([]Warning, 0)
	locks := make([]string, 0)
	tables := make([]string, 0)
	transactionalOK := true

	if strings.Contains(text, "create index") && !strings.Contains(text, "concurrently") {
		warnings = append(warnings, Warning{Code: "INDEX_BLOCKING", Message: "CREATE INDEX without CONCURRENTLY can block writes."})
		locks = append(locks, "ACCESS EXCLUSIVE risk on index target during build")
	}
	if strings.Contains(text, "create index concurrently") {
		transactionalOK = false
	}
	if strings.Contains(text, "alter table") && strings.Contains(text, "type") {
		warnings = append(warnings, Warning{Code: "TYPE_REWRITE", Message: "ALTER TABLE ... TYPE may rewrite full table and block traffic."})
		locks = append(locks, "Potential full table rewrite")
	}
	if strings.Contains(text, "drop column") {
		warnings = append(warnings, Warning{Code: "DROP_COLUMN", Message: "Dropping a column can be destructive if data exists."})
	}
	if strings.Contains(text, "alter table") && strings.Contains(text, "set not null") {
		warnings = append(warnings, Warning{Code: "TABLE_SCAN", Message: "SET NOT NULL may require full table scan."})
	}

	for _, match := range tableRegex.FindAllStringSubmatch(sqlText, -1) {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}

	return Report{
		Warnings:        warnings,
		AffectedTables:  dedupe(tables),
		EstimatedLocks:  dedupe(locks),
		TransactionalOK: transactionalOK,
	}
}

func dedupe(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
