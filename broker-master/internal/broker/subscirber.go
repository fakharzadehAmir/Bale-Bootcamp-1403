package broker

import (
	"therealbroker/pkg/broker"
)

var idGenerator = GenId{}

type Subscriber struct {
	id      int
	chanMsg chan broker.Message
}
