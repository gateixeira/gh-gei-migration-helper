package worker

type Processor interface {
	Process(entity interface{}) error
}

type Error struct {
	err    error
	entity interface{}
}

func Start(id int, jobs <-chan Processor, results chan<- Error)
