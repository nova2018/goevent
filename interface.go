package goevent

type Dispatcher interface {
	Dispatch(event Event)
}

type Subscriber interface {
	Subscribe(topic string, listener Listener, priority ...int)
	SubscribeAsync(topic string, listener Listener, transactional bool, priority ...int)
	SubscribeOnce(topic string, listener Listener, priority ...int)
	SubscribeOnceAsync(topic string, listener Listener, priority ...int)
	Unsubscribe(topic string, listener Listener)
}

type Controller interface {
	WaitAsync()
	HasCallback(topic string) bool
}

type Bus interface {
	Dispatcher
	Subscriber
	Controller
}
