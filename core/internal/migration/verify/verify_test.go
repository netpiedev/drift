package verify

import (
	"testing"

	"github.com/netpiedev/drift/core/internal/migration/parser"
)

func TestValidateChecksums(t *testing.T) {
	files := []parser.Migration{{Version: "001", Direction: parser.DirectionUp, Checksum: "abc"}}
	applied := []parser.AppliedMigration{{Version: "001", Direction: "up", Checksum: "def", Success: true}}
	issues := ValidateChecksums(files, applied)
	if len(issues) == 0 {
		t.Fatal("expected checksum mismatch issue")
	}
}
