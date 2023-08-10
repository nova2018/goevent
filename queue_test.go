package goevent

import (
	"reflect"
	"testing"
)

func newQueue[T any]() *priorityQueue[T] {
	return &priorityQueue[T]{
		queueTop: make(map[int]*priorityElem),
		queueEnd: make(map[int]*priorityElem),
	}
}

func TestPriorityQueue(t *testing.T) {
	queue := newQueue[int]()
	queue.Push(1, 1)
	queue.Push(2, 2)
	queue.Push(3, 1)
	queue.Push(4, 5)
	queue.Push(5, 4)

	if !reflect.DeepEqual(queue.ToArray(), []int{4, 5, 2, 1, 3}) {
		t.Fail()
	}
}

func BenchmarkPriorityQueue100(b *testing.B) {
	b.StopTimer()
	queue := newQueue[int]()
	initAdd(queue, 100)
	b.StartTimer()
	benchmarkAdd(b, queue, 100)
}

func BenchmarkPriorityQueue1000(b *testing.B) {
	b.StopTimer()
	queue := newQueue[int]()
	initAdd(queue, 1000)
	b.StartTimer()
	benchmarkAdd(b, queue, 1000)
}

func BenchmarkPriorityQueue10000(b *testing.B) {
	b.StopTimer()
	queue := newQueue[int]()
	initAdd(queue, 10000)
	b.StartTimer()
	benchmarkAdd(b, queue, 10000)
}

func BenchmarkPriorityQueue100000(b *testing.B) {
	b.StopTimer()
	queue := newQueue[int]()
	initAdd(queue, 100000)
	b.StartTimer()
	benchmarkAdd(b, queue, 100000)
}

func initAdd(queue *priorityQueue[int], size int) *priorityQueue[int] {
	for i := 0; i < size; i++ {
		queue.Push(i, i%40)
	}
	return queue
}

func benchmarkAdd(b *testing.B, queue *priorityQueue[int], size int) {
	for i := 0; i < b.N; i++ {
		for n := 0; n < size; n++ {
			queue.Push(n, n%40)
		}
	}
}

func BenchmarkPriorityQueue2(b *testing.B) {
	queue := &priorityQueue2[int]{
		listPriority: make([]int, 0),
		queue:        make(map[int][]int),
	}
	for i := 0; i < b.N; i++ {
		queue.Push(i, i%40)
	}
}

func BenchmarkArray(b *testing.B) {
	queue := make([]int, 0)
	for i := 0; i < b.N; i++ {
		queue = append(queue, i)
	}
}
