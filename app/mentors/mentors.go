// Package mentors implements the mentor-facing pages on Skylab
package mentors

import (
	"github.com/bokwoon95/nusskylabx/app/db"
	"github.com/bokwoon95/nusskylabx/app/skylab"
)

type Mentors struct {
	skylb skylab.Skylab
	d     db.DB
}

func New(skylb skylab.Skylab) Mentors {
	return Mentors{
		skylb: skylb,
		d:     db.New(skylb),
	}
}
