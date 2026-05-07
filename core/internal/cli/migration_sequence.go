package cli

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	SequenceTypeTimestamp = "timestamp"
	SequenceTypeDate      = "date"
	SequenceTypeSerial    = "serial"
	SequenceTypeNone      = "none"
)

var serialPrefixPattern = regexp.MustCompile(`^(\d+)_`)

func normalizeSequenceType(value string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return SequenceTypeTimestamp, nil
	}
	switch v {
	case SequenceTypeTimestamp, SequenceTypeDate, SequenceTypeSerial, SequenceTypeNone:
		return v, nil
	default:
		return "", fmt.Errorf("unsupported migrations.sequence_type %q (supported: timestamp, date, serial, none)", value)
	}
}

func buildMigrationBaseName(sequenceType string, migrationsDir string, migrationName string) (string, error) {
	switch sequenceType {
	case SequenceTypeTimestamp:
		return fmt.Sprintf("%s_%s", time.Now().UTC().Format("200601021504"), migrationName), nil
	case SequenceTypeDate:
		return fmt.Sprintf("%s_%s", time.Now().UTC().Format("20060102"), migrationName), nil
	case SequenceTypeSerial:
		next, err := nextSerialVersion(migrationsDir)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s_%s", next, migrationName), nil
	case SequenceTypeNone:
		return migrationName, nil
	default:
		return "", fmt.Errorf("unsupported sequence type: %s", sequenceType)
	}
}

func nextSerialVersion(migrationsDir string) (string, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "001", nil
		}
		return "", fmt.Errorf("read migrations dir for serial sequence: %w", err)
	}

	maxValue := 0
	width := 3
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := serialPrefixPattern.FindStringSubmatch(entry.Name())
		if len(matches) == 0 {
			continue
		}
		raw := matches[1]
		value, parseErr := strconv.Atoi(raw)
		if parseErr != nil {
			continue
		}
		if value > maxValue {
			maxValue = value
		}
		if len(raw) > width {
			width = len(raw)
		}
	}

	next := maxValue + 1
	return fmt.Sprintf("%0*d", width, next), nil
}
