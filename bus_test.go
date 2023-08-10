package goevent

import (
	"sync"
	"testing"
	"time"
)

type simpleEvent struct {
	topic string
}

func (e simpleEvent) Topic() string {
	return e.topic
}

func TestNew(t *testing.T) {
	bus := New()
	if bus == nil {
		t.Log("New EventBus not created!")
		t.Fail()
	}
}

func TestHasCallback(t *testing.T) {
	bus := New()
	bus.Subscribe("topic", func(e Event) {})
	if bus.HasCallback("topic_topic") {
		t.Fail()
	}
	if !bus.HasCallback("topic") {
		t.Fail()
	}
}

func TestSubscribeOnceAndManySubscribe(t *testing.T) {
	bus := New()
	event := "topic"
	flag := 0
	fn := func(e Event) { flag += 1 }
	bus.SubscribeOnce(event, fn)
	bus.Subscribe(event, fn)
	bus.Subscribe(event, fn)
	bus.Dispatch(&simpleEvent{topic: event})

	if flag != 3 {
		t.Fail()
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := New()
	flag := 0
	handler := func(e Event) { flag++ }
	bus.Subscribe("topic", handler)
	bus.Unsubscribe("topic", handler)
	if bus.HasCallback("topic") {
		t.Fail()
	}
	bus.Dispatch(&simpleEvent{topic: "topic"})
	if flag != 0 {
		t.Fail()
	}
}

func TestPublish(t *testing.T) {
	bus := New()
	type Ev struct {
		simpleEvent
		a   int
		err error
	}
	bus.Subscribe("topic", func(e Event) {
		if ev, ok := e.(*Ev); ok {
			if ev.a != 10 {
				t.Fail()
			}

			if ev.err != nil {
				t.Fail()
			}
		} else {
			t.Fail()
		}
	})
	bus.Dispatch(&Ev{simpleEvent{
		topic: "topic",
	}, 10, nil})
}

func TestSubcribeOnceAsync(t *testing.T) {
	results := make([]int, 0)

	bus := New()
	type Ev struct {
		simpleEvent
		a   int
		out *[]int
	}
	bus.SubscribeOnceAsync("topic", func(e Event) {
		ev := e.(*Ev)
		*ev.out = append(*ev.out, ev.a)
	})

	bus.Dispatch(&Ev{simpleEvent{
		topic: "topic",
	}, 10, &results,
	})
	bus.Dispatch(&Ev{simpleEvent{
		topic: "topic",
	}, 10, &results,
	})

	bus.WaitAsync()

	if len(results) != 1 {
		t.Fail()
	}

	if bus.HasCallback("topic") {
		t.Fail()
	}
}

func TestSubscribeAsyncTransactional(t *testing.T) {
	results := make([]int, 0)

	type Ev struct {
		simpleEvent
		a   int
		out *[]int
		dur string
	}
	bus := New()
	bus.SubscribeAsync("topic", func(e Event) {
		ev := e.(*Ev)
		sleep, _ := time.ParseDuration(ev.dur)
		time.Sleep(sleep)
		*ev.out = append(*ev.out, ev.a)
	}, true)

	bus.Dispatch(&Ev{simpleEvent{
		topic: "topic",
	}, 1, &results, "1s"})
	bus.Dispatch(&Ev{simpleEvent{
		topic: "topic",
	}, 2, &results, "0s"})

	bus.WaitAsync()

	if len(results) != 2 {
		t.Fail()
	}

	if results[0] != 1 || results[1] != 2 {
		t.Fail()
	}
}

func TestSubscribeAsync(t *testing.T) {
	results := make(chan int)

	type Ev struct {
		simpleEvent
		a   int
		out chan<- int
	}
	bus := New()
	bus.SubscribeAsync("topic", func(e Event) {
		ev := e.(*Ev)
		ev.out <- ev.a
	}, false)

	bus.Dispatch(&Ev{simpleEvent{
		topic: "topic",
	}, 1, results})
	bus.Dispatch(&Ev{simpleEvent{
		topic: "topic",
	}, 2, results})

	numResults := 0

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range results {
			numResults++
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		bus.WaitAsync()
		close(results)
	}()

	wg.Wait()

	if numResults != 2 {
		t.Fail()
	}
}
