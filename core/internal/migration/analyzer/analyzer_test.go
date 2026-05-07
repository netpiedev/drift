package analyzer

import "testing"

func TestAnalyzeSQL(t *testing.T) {
	report := AnalyzeSQL("CREATE INDEX idx_users_name ON users(name);")
	if len(report.Warnings) == 0 {
		t.Fatal("expected warning for non-concurrent index")
	}
	if report.TransactionalOK != true {
		t.Fatal("expected transactional safe for regular create index")
	}

	report2 := AnalyzeSQL("CREATE INDEX CONCURRENTLY idx_users_name ON users(name);")
	if report2.TransactionalOK {
		t.Fatal("expected non-transactional for concurrently")
	}
}
