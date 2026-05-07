package plan

import (
	"fmt"

	"github.com/netpiedev/drift/core/internal/migration/parser"
)

func PendingUp(migrations []parser.Migration, applied []parser.AppliedMigration) []parser.Migration {
	appliedUp := make(map[string]struct{})
	for _, a := range applied {
		if a.Direction == string(parser.DirectionUp) && a.Success {
			appliedUp[a.Version] = struct{}{}
		}
	}
	out := make([]parser.Migration, 0)
	for _, m := range migrations {
		if m.Direction != parser.DirectionUp {
			continue
		}
		if _, ok := appliedUp[m.Version]; ok {
			continue
		}
		out = append(out, m)
	}
	return out
}

func NextDown(migrations []parser.Migration, applied []parser.AppliedMigration, steps int) ([]parser.Migration, error) {
	if steps <= 0 {
		return nil, fmt.Errorf("steps must be > 0")
	}

	downByVersion := make(map[string]parser.Migration)
	for _, m := range migrations {
		if m.Direction == parser.DirectionDown {
			downByVersion[m.Version] = m
		}
	}

	appliedOrder := make([]string, 0)
	appliedSeen := make(map[string]struct{})
	for i := len(applied) - 1; i >= 0; i-- {
		a := applied[i]
		if a.Direction != string(parser.DirectionUp) || !a.Success {
			continue
		}
		if _, ok := appliedSeen[a.Version]; ok {
			continue
		}
		appliedSeen[a.Version] = struct{}{}
		appliedOrder = append(appliedOrder, a.Version)
	}

	if len(appliedOrder) == 0 {
		return nil, nil
	}

	if steps > len(appliedOrder) {
		steps = len(appliedOrder)
	}

	out := make([]parser.Migration, 0, steps)
	for _, version := range appliedOrder[:steps] {
		down, ok := downByVersion[version]
		if !ok {
			return nil, fmt.Errorf("missing down migration for applied version %s", version)
		}
		out = append(out, down)
	}
	return out, nil
}
