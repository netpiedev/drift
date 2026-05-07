package status

import "github.com/netpiedev/drift/core/internal/migration/parser"

type Row struct {
	Version string
	Name    string
	Applied bool
	Dir     string
}

func BuildRows(files []parser.Migration, applied []parser.AppliedMigration) []Row {
	appliedUp := map[string]bool{}
	for _, a := range applied {
		if a.Direction == string(parser.DirectionUp) && a.Success {
			appliedUp[a.Version] = true
		}
	}
	rows := make([]Row, 0)
	seenVersion := map[string]bool{}
	for _, f := range files {
		if f.Direction != parser.DirectionUp {
			continue
		}
		if seenVersion[f.Version] {
			continue
		}
		seenVersion[f.Version] = true
		rows = append(rows, Row{Version: f.Version, Name: f.Name, Applied: appliedUp[f.Version], Dir: f.Path})
	}
	return rows
}
