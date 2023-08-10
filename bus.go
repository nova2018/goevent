package goevent

import (
	"reflect"
	"sync"
	"sync/atomic"
)

type eventBus struct {
	wg          *sync.WaitGroup
	lock        *sync.RWMutex
	mapHandlers map[string][]*eventHandler
	mapPriority map[string][2]int
}

func New() Bus {
	return &eventBus{
		wg:          &sync.WaitGroup{},
		lock:        &sync.RWMutex{},
		mapHandlers: make(map[string][]*eventHandler),
		mapPriority: make(map[string][2]int),
	}
}

type eventHandler struct {
	wg            *sync.WaitGroup
	priority      int
	once          *sync.Once
	async         bool
	transactional bool
	removed       int32
	tLock         *sync.Mutex // lock for transactional
	callback      Listener
	callbackValue reflect.Value
}

func (h *eventHandler) remove() {
	h.removed = 1
}

func (h *eventHandler) call(e Event) {
	callback := h.callback
	if h.once != nil {
		callback = func(callback Listener) Listener {
			return func(event Event) {
				h.once.Do(func() {
					if atomic.CompareAndSwapInt32(&h.removed, 0, 1) {
						callback(event)
					}
				})
			}
		}(callback)
	}
	callback = func(callback Listener) Listener {
		return func(event Event) {
			if stop, ok := event.(stoppable); ok && stop.IsPropagationStopped() {
				return
			}
			callback(event)
		}
	}(callback)
	if h.transactional {
		callback = func(callback Listener) Listener {
			h.tLock.Lock()
			return func(event Event) {
				defer h.tLock.Unlock()
				callback(event)
			}
		}(callback)
	}
	if h.async {
		callback = func(callback Listener) Listener {
			return func(event Event) {
				h.wg.Add(1)
				go func() {
					defer h.wg.Done()
					callback(event)
				}()
			}
		}(callback)
	}
	callback = func(callback Listener) Listener {
		return func(event Event) {
			if h.removed == 0 {
				callback(event)
			}
		}
	}(callback)
	callback(e)
}

func getPriority(priority ...int) int {
	if len(priority) == 0 {
		return 0
	}
	return priority[0]
}

func (s *eventBus) Subscribe(topic string, listener Listener, priority ...int) {
	s.doSubscribe(
		topic,
		&eventHandler{
			priority: getPriority(priority...),
			callback: listener,
		},
	)
}

func (s *eventBus) SubscribeAsync(topic string, listener Listener, transactional bool, priority ...int) {
	s.doSubscribe(
		topic,
		&eventHandler{
			priority:      getPriority(priority...),
			callback:      listener,
			async:         true,
			transactional: transactional,
		},
	)
}

func (s *eventBus) SubscribeOnce(topic string, listener Listener, priority ...int) {
	s.doSubscribe(
		topic,
		&eventHandler{
			priority: getPriority(priority...),
			callback: listener,
			once:     &sync.Once{},
		},
	)
}

func (s *eventBus) SubscribeOnceAsync(topic string, listener Listener, priority ...int) {
	s.doSubscribe(
		topic,
		&eventHandler{
			priority: getPriority(priority...),
			callback: listener,
			once:     &sync.Once{},
			async:    true,
		},
	)
}

func (s *eventBus) Unsubscribe(topic string, listener Listener) {
	handler := s.findHandler(topic, listener)
	if handler != nil {
		handler.remove()
	}
}

func (s *eventBus) doSubscribe(topic string, handler *eventHandler) {
	s.lock.Lock()
	defer s.lock.Unlock()
	handler.wg = s.wg
	handler.callbackValue = reflect.ValueOf(handler.callback)
	handler.tLock = &sync.Mutex{}
	s.mapHandlers[topic] = append(s.mapHandlers[topic], handler)
	if priority, ok := s.mapPriority[topic]; ok {
		if priority[0] > handler.priority {
			priority[0] = handler.priority
		}
		if priority[1] < handler.priority {
			priority[1] = handler.priority
		}
		s.mapPriority[topic] = priority
	} else {
		s.mapPriority[topic] = [2]int{handler.priority, handler.priority}
	}
}

func (s *eventBus) findHandler(topic string, listener Listener) *eventHandler {
	callback := reflect.ValueOf(listener)
	if _, ok := s.mapHandlers[topic]; ok {
		for _, handler := range s.mapHandlers[topic] {
			if handler.callbackValue.Type() == callback.Type() &&
				handler.callbackValue.Pointer() == callback.Pointer() {
				return handler
			}
		}
	}
	return nil
}

func (s *eventBus) getHandlersForEvent(topic string) []*eventHandler {
	var cpHandlers []*eventHandler
	s.lock.RLock()
	isPriority := false
	if handlers, ok := s.mapHandlers[topic]; ok && len(handlers) > 0 {
		cpHandlers = make([]*eventHandler, len(handlers))
		copy(cpHandlers, handlers)
		priority := s.mapPriority[topic]
		isPriority = priority[0] != priority[1]
	}
	s.lock.RUnlock()
	if len(cpHandlers) > 0 {
		if !isPriority {
			// 无优先级队列
			return cpHandlers
		}
		queue := getQueue()
		defer freeQueue(queue)
		for _, handler := range cpHandlers {
			queue.Push(handler, handler.priority)
		}
		return queue.ToArray()
	}
	return nil
}

func (s *eventBus) Dispatch(event Event) {
	listHandler := s.getHandlersForEvent(event.Topic())
	if len(listHandler) > 0 {
		for _, h := range listHandler {
			if stop, ok := event.(stoppable); ok && stop.IsPropagationStopped() {
				break
			}
			if h.removed == 1 {
				continue
			}
			h.call(event)
		}
	}
}

func (s *eventBus) WaitAsync() {
	s.wg.Wait()
}

func (s *eventBus) HasCallback(topic string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if handlers, ok := s.mapHandlers[topic]; !ok || len(handlers) < 1 {
		return false
	} else {
		for _, h := range handlers {
			if h.removed == 0 {
				return true
			}
		}
		return false
	}
}
