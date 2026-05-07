package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/netpiedev/drift/core/internal/util"
)

var migrationPattern = regexp.MustCompile(`^(\d+)_([a-zA-Z0-9_\-]+)\.(up|down)\.(sql|ts|js|go|py)$`)

func ParseDir(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migration dir: %w", err)
	}

	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		matches := migrationPattern.FindStringSubmatch(name)
		if len(matches) == 0 {
			continue
		}
		full := filepath.Join(dir, name)
		content, err := os.ReadFile(full)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", name, err)
		}

		migrations = append(migrations, Migration{
			Version:   matches[1],
			Name:      matches[2],
			Direction: Direction(matches[3]),
			Ext:       matches[4],
			Path:      full,
			Checksum:  util.SHA256Hex(content),
			Content:   content,
		})
	}

	sort.SliceStable(migrations, func(i, j int) bool {
		if migrations[i].Version == migrations[j].Version {
			if migrations[i].Direction == migrations[j].Direction {
				return migrations[i].Ext < migrations[j].Ext
			}
			return migrations[i].Direction < migrations[j].Direction
		}
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func GroupByVersion(migrations []Migration) map[string][]Migration {
	grouped := make(map[string][]Migration)
	for _, m := range migrations {
		grouped[m.Version] = append(grouped[m.Version], m)
	}
	return grouped
}
