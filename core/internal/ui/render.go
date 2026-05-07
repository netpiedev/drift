package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/netpiedev/drift/core/internal/migration/executor"
)

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("27")).Padding(0, 1)
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	mutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

func RenderMigrationResults(results []executor.ApplyResult, dryRun bool) string {
	rows := make([][]string, 0, len(results))
	total := time.Duration(0)
	for _, r := range results {
		total += r.Duration
		rows = append(rows, []string{
			r.Migration.Version,
			r.Migration.Name,
			string(r.Migration.Direction),
			r.Duration.String(),
			boolLabel(r.TransactionalSafe),
		})
	}
	t := table.New().
		Headers("Version", "Name", "Dir", "Duration", "Tx").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			if row%2 == 0 {
				return mutedStyle
			}
			return lipgloss.NewStyle()
		})

	state := "APPLIED"
	if dryRun {
		state = "DRY-RUN"
	}
	builder := strings.Builder{}
	builder.WriteString(successStyle.Render(fmt.Sprintf("Migration summary (%s)", state)))
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Total migrations: %d | Total duration: %s\n", len(results), total))
	builder.WriteString(t.String())
	builder.WriteString("\n")

	for _, r := range results {
		if len(r.Warnings) == 0 {
			continue
		}
		builder.WriteString(warnStyle.Render(fmt.Sprintf("Warnings for %s_%s", r.Migration.Version, r.Migration.Name)))
		builder.WriteString("\n")
		for _, w := range r.Warnings {
			builder.WriteString("  - " + w + "\n")
		}
	}
	return strings.TrimSpace(builder.String())
}

func RenderWarnings(lines []string) string {
	if len(lines) == 0 {
		return successStyle.Render("No warnings")
	}
	b := strings.Builder{}
	b.WriteString(warnStyle.Render("Warnings"))
	b.WriteString("\n")
	for _, line := range lines {
		b.WriteString("- " + line + "\n")
	}
	return strings.TrimSpace(b.String())
}

func boolLabel(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}
