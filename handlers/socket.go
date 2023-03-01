package handlers

import (
	"encoding/json"
	"github.com/olahol/melody"
	"log"
	"time"
)

func SocketMessageHandler(m *melody.Melody) {

	m.HandleConnect(func(s *melody.Session) {
		log.Println("Client Connected:", s.Request.RemoteAddr)
		log.Printf("Connected Clients: %d, subscriptions:%d", m.Len()+1, len(subscribers)+1)

		// Todo NIP-42 challenge
		//nipChallenge := make([]byte, 8)
		//_, err := rand.Read(nipChallenge)
		//if err != nil {
		//	log.Println("NIP-42: Could not create a Rand challenge!")
		//}
		//challenge := hex.EncodeToString(nipChallenge)

		s.Set("events_received", 0)
		s.Set("events_sent", 0)
		s.Set("connected", time.Now())
	})
	m.HandleMessage(func(s *melody.Session, msg []byte) {
		var request []json.RawMessage
		if err := json.Unmarshal(msg, &request); err != nil {
			//log.Println("ERROR: Could not parse JSON!")
			return
		}

		var eventType string
		err := json.Unmarshal(request[0], &eventType)
		if err != nil {
			return
		}

		if len(eventType) > 2 {
			go func() {
				_ = NostrEventHandler(eventType, request, s)
			}()
		}
	})

	m.HandleDisconnect(func(s *melody.Session) {
		removeSubscriber(s)
		evRec, _ := s.Get("events_received")
		evRec = evRec.(int)
		evSen, _ := s.Get("events_sent")
		evSen = evSen.(int)
		connAt, _ := s.Get("connected")
		connected := connAt.(time.Time)
		diff := time.Now().Sub(connected).Seconds()
		log.Printf("Client Disconnected: %s - Sent: %d events, received: %d events, connected: %.2fs", s.Request.RemoteAddr, evSen, evRec, diff)
		log.Printf("Connected Clients: %d, subscriptions: %d", m.Len(), len(subscribers))
	})

}
