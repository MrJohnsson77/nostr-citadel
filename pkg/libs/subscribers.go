package libs

import (
	"encoding/json"
	"github.com/nbd-wtf/go-nostr"
	"github.com/olahol/melody"
	"sync"
)

type Subscriber struct {
	filters nostr.Filters
}

var Subscribers = make(map[*melody.Session]map[string]*Subscriber)
var subscriberMutex = sync.Mutex{}

func SetSubscriber(id string, ws *melody.Session, filters nostr.Filters) {
	subscriberMutex.Lock()
	defer func() {
		subscriberMutex.Unlock()
	}()

	subs, ok := Subscribers[ws]
	if !ok {
		subs = make(map[string]*Subscriber)
		Subscribers[ws] = subs
	}

	subs[id] = &Subscriber{
		filters: filters,
	}
}

func RemoveSubscriberId(ws *melody.Session, id string) {
	subscriberMutex.Lock()
	defer func() {
		subscriberMutex.Unlock()
	}()

	subs, ok := Subscribers[ws]
	if ok {
		delete(Subscribers[ws], id)
		if len(subs) == 0 {
			delete(Subscribers, ws)
		}
	}
}

func RemoveSubscriber(ws *melody.Session) {
	subscriberMutex.Lock()
	defer subscriberMutex.Unlock()

	_, ok := Subscribers[ws]
	if ok {
		delete(Subscribers, ws)
	}
}

func NotifySubscribers(event *nostr.Event) {
	subscriberMutex.Lock()
	defer func() {
		subscriberMutex.Unlock()
	}()

	for ws, subs := range Subscribers {
		for id, listener := range subs {
			if !listener.filters.Match(event) {
				continue
			}
			message, _ := json.Marshal([]interface{}{"EVENT", id, event})
			_ = ws.Write(message)
		}
	}
}
