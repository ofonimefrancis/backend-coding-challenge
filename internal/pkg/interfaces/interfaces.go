package interfaces

import "time"

type IDGenerator interface {
	Generate() string
}

type TimeProvider interface {
	Now() time.Time
}
