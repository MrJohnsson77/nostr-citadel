package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/olahol/melody"
	"github.com/spf13/viper"
	"log"
	"nostr-citadel/storage"
)

type nip11Amount []struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}

type SpecialNIP11 struct {
	Contact     string `json:"contact"`
	Description string `json:"description"`
	Fees        struct {
		Admission   nip11Amount   `json:"admission"`
		Publication []interface{} `json:"publication"`
	} `json:"fees"`
	Limitation struct {
		AuthRequired     bool `json:"auth_required"`
		MaxContentLength int  `json:"max_content_length"`
		MaxEventTags     int  `json:"max_event_tags"`
		MaxFilters       int  `json:"max_filters"`
		MaxLimit         int  `json:"max_limit"`
		MaxMessageLength int  `json:"max_message_length"`
		MaxSubidLength   int  `json:"max_subid_length"`
		MaxSubscriptions int  `json:"max_subscriptions"`
		MinPowDifficulty int  `json:"min_pow_difficulty"`
		MinPrefix        int  `json:"min_prefix"`
		PaymentRequired  bool `json:"payment_required"`
	} `json:"limitation"`
	Name          string `json:"name"`
	PaymentsURL   string `json:"payments_url"`
	Pubkey        string `json:"pubkey"`
	Software      string `json:"software"`
	SupportedNips []int  `json:"supported_nips"`
	Version       string `json:"version"`
}

func GetPubKey(pubKey string) (string, string) {
	if len(pubKey) < 4 {
		return "", ""
	}
	var pub, npub string
	if pubKey[:4] == "npub" {
		_, v, err := nip19.Decode(pubKey)
		if err != nil {
			pub = ""
			npub = ""
		} else {
			pub = v.(string)
			npub = pubKey
		}
	} else {
		pub = pubKey
		pk, err := nip19.EncodePublicKey(pub)
		if err != nil {
			pub = ""
			npub = ""
		} else {
			npub = pk
		}
	}
	return pub, npub
}

func acceptEvent(event nostr.Event) bool {

	if viper.GetBool("relay_config.public_relay") {
		return true
	}
	wl := storage.GetWhitelisted(event.PubKey)
	whitelisted := event.PubKey == wl.PubKey
	nPub, _ := nip19.EncodePublicKey(event.PubKey)

	if !whitelisted {
		log.Println("Blocked event from:", nPub)
		return false
	}

	log.Println("Accepted event from:", nPub)
	return true
}

func NostrEventHandler(eventType string, requestData []json.RawMessage, s *melody.Session) error {

	switch eventType {
	case "EVENT":

		var event nostr.Event
		if err := json.Unmarshal(requestData[1], &event); err != nil {
			//notice = "failed to decode event: " + err.Error()
			return nil
		}

		serialized := event.Serialize()

		// Assign ID
		hash := sha256.Sum256(serialized)
		event.ID = hex.EncodeToString(hash[:])

		// check signature (requires the ID to be set)
		if ok, err := event.CheckSignature(); err != nil {
			log.Println("NOTICE", "error: failed to verify signature")
			sn, _ := json.Marshal([]interface{}{"OK", event.ID, false, "error: failed to verify signature"})
			_ = s.Write(sn)
			return nil
		} else if !ok {
			log.Println("NOTICE", "signature is invalid")
			sn, _ := json.Marshal([]interface{}{"OK", event.ID, false, "invalid: signature is invalid"})
			_ = s.Write(sn)
			return nil
		}

		// Check if event is accepted
		if !acceptEvent(event) {
			// NIP-20
			sn, _ := json.Marshal([]interface{}{"OK", event.ID, false, "blocked: no event posting access, contact admin."})
			_ = s.Write(sn)
			return nil
		}

		ok, message := storage.WriteEvent(&event)

		if ok {
			evr, _ := s.Get("events_sent")
			evRec := evr.(int)
			s.Set("events_sent", evRec+1)
			notifySubscribers(&event)
		}
		msg, _ := json.Marshal([]interface{}{"OK", event.ID, ok, message})
		_ = s.Write(msg)
		return nil
	case "REQ":
		var id string
		_ = json.Unmarshal(requestData[1], &id)
		if id == "" {
			log.Println("NOTICE", "invalid: Request has no ID")
			sn, _ := json.Marshal([]interface{}{"NOTICE", "invalid: Request has no ID"})
			_ = s.Write(sn)
			return nil
		}
		filters := make(nostr.Filters, len(requestData)-2)
		for i, filterReq := range requestData[2:] {
			if err := json.Unmarshal(
				filterReq,
				&filters[i],
			); err != nil {
				log.Println("NOTICE", "Failed to decode filter")
				sn, _ := json.Marshal([]interface{}{"NOTICE", "Failed to decode filter"})
				_ = s.Write(sn)
				return nil
			}

			filter := &filters[i]

			// Todo: Implement NIP-42 - Only allow authed users to get their private messages (kind-4)

			events, err := storage.GetEventsQuery(filter)
			if err != nil {
				log.Printf("Req Error: %v", err)
				continue
			}

			// Backup if query is broken.
			if filter.Limit > 0 && len(events) > filter.Limit {
				events = events[0:filter.Limit]
			}

			evr, _ := s.Get("events_received")
			evRec := evr.(int)
			s.Set("events_received", len(events)+evRec)

			for _, event := range events {
				sn, _ := json.Marshal([]interface{}{"EVENT", id, event})
				_ = s.Write(sn)
			}

		}
		setSubscriber(id, s, filters)
		sn, _ := json.Marshal([]interface{}{"EOSE", id})
		_ = s.Write(sn)
		return nil
	case "CLOSE":
		var id string
		_ = json.Unmarshal(requestData[1], &id)
		if id == "" {
			return nil
		}
		removeSubscriberId(s, id)
		return nil
	case "AUTH":
		// Todo: NIP-42
		return nil
	default:
		//reqData, _ := json.MarshalIndent(&requestData, "", "   ")
		//_ = SocketSend(reqData)
		return nil
	}
}

func NostrNip11() interface{} {

	supportedNIPs := []int{9, 11, 12, 15, 16, 20}

	//if authNip42 {
	//	supportedNIPs = append(supportedNIPs, 42)
	//}

	pubKey, _ := GetPubKey(viper.GetString("relay_config.admin_npub"))
	relayUrl := viper.GetString("relay_config.relay_url")

	if len(relayUrl) > 1 {
		relayUrl = relayUrl + "/invoices"
	}

	nip11Info := &SpecialNIP11{
		Name:          viper.GetString("relay_config.name"),
		Description:   viper.GetString("relay_config.description"),
		Pubkey:        pubKey,
		Contact:       viper.GetString("relay_config.admin_email"),
		SupportedNips: supportedNIPs,
		Software:      "git+https://github.com/MrJohnsson77/nostr-citadel.git",
		Version:       "0.0.1-alpha",
		PaymentsURL:   relayUrl,
	}

	// Todo: Connect it..
	nip11Info.Limitation.MaxMessageLength = 262144
	nip11Info.Limitation.MaxSubscriptions = 10
	nip11Info.Limitation.MaxFilters = 10
	nip11Info.Limitation.MaxLimit = 5000
	nip11Info.Limitation.MaxSubidLength = 500
	nip11Info.Limitation.MinPrefix = 4
	nip11Info.Limitation.MaxEventTags = 2500
	nip11Info.Limitation.MaxContentLength = 102400
	nip11Info.Limitation.MinPowDifficulty = 0
	nip11Info.Limitation.AuthRequired = false
	nip11Info.Limitation.PaymentRequired = true
	nip11Info.Fees.Admission = nip11Amount{
		{
			Amount: 50000,
			Unit:   "msats",
		},
	}
	nip11Info.Fees.Publication = []interface{}{}

	return nip11Info
}
