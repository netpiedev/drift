package parser

import "time"

type Direction string

const (
	DirectionUp   Direction = "up"
	DirectionDown Direction = "down"
)

type Migration struct {
	Version   string
	Name      string
	Direction Direction
	Ext       string
	Path      string
	Checksum  string
	Content   []byte
}

type AppliedMigration struct {
	Version           string
	Name              string
	Direction         string
	Checksum          string
	ExecutionTimeMS   int64
	AppliedBy         string
	AppliedAt         time.Time
	Success           bool
	RollbackSupported bool
}
