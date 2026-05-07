package parser

import "fmt"

func ValidatePairs(migrations []Migration) error {
	grouped := GroupByVersion(migrations)
	for version, items := range grouped {
		hasUp := false
		hasDown := false
		name := ""
		for _, m := range items {
			if name == "" {
				name = m.Name
			}
			if m.Name != name {
				return fmt.Errorf("version %s has mixed names", version)
			}
			switch m.Direction {
			case DirectionUp:
				hasUp = true
			case DirectionDown:
				hasDown = true
			}
		}
		if !hasUp || !hasDown {
			return fmt.Errorf("version %s must include both up and down migration files", version)
		}
	}
	return nil
}
