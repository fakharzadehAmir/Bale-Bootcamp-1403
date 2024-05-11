package broker

import "sync"

var GlobalID = new(int)

type GenId struct {
	id int
	sync.Mutex
}

func (gi *GenId) NewID() int {
	gi.Lock()
	defer gi.Unlock()
	*GlobalID += 1
	gi.id = *GlobalID
	return gi.id
}
