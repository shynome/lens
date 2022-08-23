package sdk

import "github.com/donovanhide/eventsource"

var _ eventsource.Event = &Task{}
