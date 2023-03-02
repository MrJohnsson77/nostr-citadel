package handlers

import (
	"encoding/json"
	"github.com/nbd-wtf/go-nostr"
	"github.com/olahol/melody"
	"sync"
)

type Subscriber struct {
	filters nostr.Filters
}

var subscribers = make(map[*melody.Session]map[string]*Subscriber)
var subscriberMutex = sync.Mutex{}

func setSubscriber(id string, ws *melody.Session, filters nostr.Filters) {
	subscriberMutex.Lock()
	defer func() {
		subscriberMutex.Unlock()
	}()

	subs, ok := subscribers[ws]
	if !ok {
		subs = make(map[string]*Subscriber)
		subscribers[ws] = subs
	}

	subs[id] = &Subscriber{
		filters: filters,
	}
}

func removeSubscriberId(ws *melody.Session, id string) {
	subscriberMutex.Lock()
	defer func() {
		subscriberMutex.Unlock()
	}()

	subs, ok := subscribers[ws]
	if ok {
		delete(subscribers[ws], id)
		if len(subs) == 0 {
			delete(subscribers, ws)
		}
	}
}

func removeSubscriber(ws *melody.Session) {
	subscriberMutex.Lock()
	defer subscriberMutex.Unlock()

	_, ok := subscribers[ws]
	if ok {
		delete(subscribers, ws)
	}
}

func notifySubscribers(event *nostr.Event) {
	subscriberMutex.Lock()
	defer func() {
		subscriberMutex.Unlock()
	}()

	for ws, subs := range subscribers {
		for id, listener := range subs {
			if !listener.filters.Match(event) {
				continue
			}
			message, _ := json.Marshal([]interface{}{"EVENT", id, event})
			_ = ws.Write(message)
		}
	}
}
