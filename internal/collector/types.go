package collector

import "phant/internal/dump"

const DefaultBufferSize = 2000

type Decoder func(line string) (*dump.Event, error)

type Event = dump.Event
