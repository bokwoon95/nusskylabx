// Package admins implements the admin-facing pages on Skylab
package admins

import (
	"github.com/bokwoon95/nusskylabx/app/db"
	"github.com/bokwoon95/nusskylabx/app/skylab"
)

type Admins struct {
	skylb skylab.Skylab
	d     db.DB
}

func New(skylb skylab.Skylab) Admins {
	return Admins{
		skylb: skylb,
		d:     db.New(skylb),
	}
}

func removeEmptyStrings(values []string) (purged []string) {
	for _, value := range values {
		if value == "" {
			continue
		}
		purged = append(purged, value)
	}
	return purged
}

func dedupStrings(values []string) (deduped []string) {
	uniq := make(map[string]bool)
	for _, value := range values {
		if !uniq[value] {
			deduped = append(deduped, value)
			uniq[value] = true
		}
	}
	return deduped
}
