package goevent

import (
	"sort"
	"sync"
)

type priorityQueue2[T any] struct {
	listPriority []int
	queue        map[int][]T
}

func (p *priorityQueue2[T]) Push(value T, priority int) {
	if _, ok := p.queue[priority]; !ok {
		p.listPriority = append(p.listPriority, priority)
		p.queue[priority] = make([]T, 0)
	}
	p.queue[priority] = append(p.queue[priority], value)
}

type priorityQueue[T any] struct {
	listPriority []int
	queueTop     map[int]*priorityElem
	queueEnd     map[int]*priorityElem
}

type priorityElem struct {
	next *priorityElem
	elem any
	num  int
}

func (p *priorityQueue[T]) Push(value T, priority int) {
	elem := getPriorityElem()
	elem.elem = value
	if _, ok := p.queueTop[priority]; !ok {
		p.listPriority = append(p.listPriority, priority)
		top := getPriorityElem()
		p.queueTop[priority] = top
		p.queueEnd[priority] = top
	}
	p.queueEnd[priority].next = elem
	p.queueEnd[priority] = elem
	p.queueTop[priority].num++
}

func (p *priorityQueue[T]) Free() {
	p.listPriority = nil
	for _, elem := range p.queueTop {
		setPriorityElem(elem)
	}
}

func (p *priorityQueue[T]) ToArray() []T {
	l := 0
	for _, x := range p.queueTop {
		l += x.num
	}
	listPriority := append([]int{}, p.listPriority...)
	sort.Sort(sort.Reverse(sort.IntSlice(listPriority)))
	r := make([]T, 0, l)
	for _, priority := range listPriority {
		cur := p.queueTop[priority]
		for cur.next != nil {
			cur = cur.next
			r = append(r, cur.elem.(T))
		}
	}
	return r
}

func getQueue() *priorityQueue[*eventHandler] {
	elem := queuePool.Get().(*priorityQueue[*eventHandler])
	elem.queueTop = make(map[int]*priorityElem)
	elem.queueEnd = make(map[int]*priorityElem)
	return elem
}

func freeQueue(q *priorityQueue[*eventHandler]) {
	q.Free()
	queuePool.Put(q)
}

func getPriorityElem() *priorityElem {
	item := priorityElemPool.Get()
	elem := item.(*priorityElem)
	elem.next = nil
	elem.elem = nil
	return elem
}

func setPriorityElem(elem *priorityElem) {
	priorityElemPool.Put(elem)
}

var (
	priorityElemPool = &sync.Pool{New: func() any {
		return &priorityElem{}
	}}
	queuePool = &sync.Pool{
		New: func() any {
			return &priorityQueue[*eventHandler]{}
		},
	}
)
