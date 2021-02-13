package queue

type Queue interface {
	Push(...interface{})
	Pop() (interface{}, bool)
}

type queue struct {
	items []interface{}
}

func (q *queue) Push(items ...interface{}) {
	q.items = append(q.items, items...)
}

func (q *queue) Pop() (interface{}, bool) {
	if len(q.items) > 0 {
		item := q.items[0]
		q.items = q.items[1:]
		return item, true
	}
	return nil, false
}

func New() Queue {
	return &queue{
		items: []interface{}{},
	}
}