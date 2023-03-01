package storage

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"github.com/spf13/viper"
	"log"
	"strconv"
	"strings"
	"time"
)

type Pleb struct {
	PubKey     string
	Content    string
	LastSynced int64
}

type Whitelisted struct {
	PubKey  string
	Admin   bool
	Created time.Time
}

type Admin struct {
	PubKey string
}

func SetAdmin(pubKey string) {
	admin := Admin{}
	_ = DB.Get(&admin, "SELECT pubkey FROM whitelist WHERE pubkey = ? and admin = 1", pubKey)

	if admin.PubKey == pubKey {
		return
	} else {
		_, _ = DB.Exec(`DELETE FROM whitelist where admin = 1`)
		_, _ = DB.Exec(`DELETE FROM event where kind = 0 and pubkey = ?`, admin.PubKey)
		_, _ = DB.Exec(`INSERT into whitelist (pubkey, created_at, sync, last_synced, admin) values ($1,$2,$3,$4,$5)`,
			pubKey, time.Now().Unix(), 1, time.Now().Add(time.Duration(-24*viper.GetInt("sync_config.sync_days_on_init"))*time.Hour).Unix(), 1)
	}
}

func AddWhitelist(pubKey string) {
	_, _ = DB.Exec(`INSERT into whitelist (pubkey, created_at, sync, last_synced, admin) values ($1,$2,$3,$4,$5) on conflict (pubkey) DO NOTHING`,
		pubKey, time.Now().Unix(), 0, time.Now().Add(time.Duration(-24*viper.GetInt("sync_config.sync_days_on_init"))*time.Hour).Unix(), 0)
}

func RemoveWhitelist(pubKey string) {
	_, _ = DB.Exec(`DELETE from whitelist where pubkey = $1 and not admin`, pubKey)
	_, _ = DB.Exec(`DELETE from event where pubkey = $1`, pubKey)
}

//func GetWhitelist() []Whitelisted {
//	var p []Whitelisted
//	_ = DB.Select(&p, "SELECT pubkey,admin,created_at as created FROM whitelist")
//	fmt.Printf("%v", p)
//	return p
//}

func SetLastSync(pubKey string) {
	_, _ = DB.Exec(`UPDATE whitelist set last_synced = $1 WHERE pubkey = $2`, time.Now().Unix(), pubKey)
}

func GetWhitelisted(pubKey string) (wl Whitelisted) {
	p := Whitelisted{}
	_ = DB.Get(&p, "SELECT pubkey,admin,created_at as created FROM whitelist WHERE pubkey = ?", pubKey)
	return p
}

// GetPlebsToSync Todo: Move default relays to config.
func GetPlebsToSync() (plebs []Pleb, err error) {
	var wl []Pleb

	wlSync := viper.GetBool("sync_config.sync_whitelisted")
	query := `SELECT wl.pubkey as pubkey,coalesce(content,
    '{"wss://nostr.fmt.wiz.biz":{"write":true,"read":true},"wss://relay.snort.social":{"write":true,"read":true},"wss://puravida.nostr.land":{"write":true,"read":true},"wss://nos.lol":{"write":true,"read":true},"wss://relay.damus.io":{"write":true,"read":true},"wss://nostr.milou.lol":{"write":true,"read":true},"wss://relay.nostr.bg":{"write":true,"read":true},"wss://bitcoiner.social":{"write":true,"read":true},"wss://relay.current.fyi":{"write":true,"read":true},"wss://nostr.zebedee.cloud":{"write":true,"read":true},"wss://eden.nostr.land":{"write":true,"read":true}}') as content,
    coalesce(last_synced,1672527600000) as lastsynced from whitelist wl LEFT JOIN (SELECT * from event WHERE kind = 3) as e on wl.pubkey = e.pubkey`
	if wlSync {
		err = DB.Select(&wl, query)
	} else {
		err = DB.Select(&wl, query+" WHERE wl.sync = 1")
	}

	return wl, err
}

func GetEventsQuery(filter *nostr.Filter) (events []nostr.Event, err error) {
	var conditions []string
	var params []any

	if filter == nil {
		err = errors.New("filter cannot be null")
		return
	}
	if filter.IDs != nil {
		if len(filter.IDs) > 500 {
			// too many ids, fail everything
			return
		}

		likeIds := make([]string, 0, len(filter.IDs))
		for _, id := range filter.IDs {
			// to prevent sql attack here we will check if
			// these ids are valid 32byte hex
			parsed, err := hex.DecodeString(id)
			if err != nil || len(parsed) != 32 {
				continue
			}
			//likeIds = append(likeIds, fmt.Sprintf("id LIKE '%x%%'", parsed))
			likeIds = append(likeIds, fmt.Sprintf("id = '%x'", parsed))
		}
		if len(likeIds) == 0 {
			// ids being [] mean you won't get anything
			return
		}
		conditions = append(conditions, "("+strings.Join(likeIds, " OR ")+")")
	}

	if filter.Authors != nil {
		if len(filter.Authors) > 500 {
			// too many authors, fail everything
			return
		}

		likeKeys := make([]string, 0, len(filter.Authors))
		for _, key := range filter.Authors {
			// to prevent sql attack here we will check if
			// these keys are valid 32byte hex
			parsed, err := hex.DecodeString(key)
			if err != nil || len(parsed) != 32 {
				continue
			}
			//likeKeys = append(likeKeys, fmt.Sprintf("pubkey LIKE '%x%%'", parsed))
			likeKeys = append(likeKeys, fmt.Sprintf("pubkey = '%x'", parsed))
		}
		if len(likeKeys) == 0 {
			// authors being [] mean you won't get anything
			return
		}
		conditions = append(conditions, "("+strings.Join(likeKeys, " OR ")+")")
	}

	if filter.Kinds != nil {
		if len(filter.Kinds) > 10 {
			// too many kinds, fail everything
			return
		}

		if len(filter.Kinds) == 0 {
			// kinds being [] mean you won't get anything
			return
		}
		// no sql injection issues since these are ints
		inKinds := make([]string, len(filter.Kinds))
		for i, kind := range filter.Kinds {
			inKinds[i] = strconv.Itoa(kind)
		}
		conditions = append(conditions, `kind IN (`+strings.Join(inKinds, ",")+`)`)
	}

	tagQuery := make([]string, 0, 1)
	for _, values := range filter.Tags {
		if len(values) == 0 {
			// any tag set to [] is wrong
			return
		}

		// add these tags to the query
		tagQuery = append(tagQuery, values...)

		if len(tagQuery) > 10 {
			// too many tags, fail everything
			return
		}
	}

	if len(tagQuery) > 0 {
		arrayBuild := make([]string, len(tagQuery))
		for i, tagValue := range tagQuery {
			arrayBuild[i] = "?"
			params = append(params, tagValue)
		}
		// Works for now...
		conditions = append(conditions, `tags LIKE '%' || ? || '%'`)
	}

	if filter.Since != nil {
		conditions = append(conditions, "created_at > ?")
		params = append(params, filter.Since.Unix())
	}

	if filter.Until != nil {
		conditions = append(conditions, "created_at < ?")
		params = append(params, filter.Until.Unix())
	}

	if len(conditions) == 0 {
		// fallback
		conditions = append(conditions, "true")
	}

	// Limit how many events we respond with.
	respLimit := viper.GetInt("relay_config.limits.response_event_limit")
	if filter.Limit < 1 || filter.Limit > respLimit {
		params = append(params, respLimit)
	} else {
		params = append(params, filter.Limit)
	}

	query := DB.Rebind(`SELECT
	 id, pubkey, created_at, kind, tags, content, sig
	FROM event WHERE ` +
		strings.Join(conditions, " AND ") +
		" ORDER BY created_at DESC LIMIT ?")

	fmt.Printf("%v\n", query)
	fmt.Printf("%v\n", params)

	rows, err := DB.Query(query, params...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch events using query %q: %w", query, err)
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var evt nostr.Event
		var timestamp time.Time
		err := rows.Scan(&evt.ID, &evt.PubKey, &timestamp,
			&evt.Kind, &evt.Tags, &evt.Content, &evt.Sig)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		evt.CreatedAt = timestamp //time.Unix(timestamp, 0)
		events = append(events, evt)
	}

	return events, nil
}

func WriteEvent(event *nostr.Event) (bool, string) {

	switch event.Kind {

	case nostr.KindDeletion: // 5
		for _, tag := range event.Tags {
			if len(tag) >= 2 && tag[0] == "e" {
				eventID, err := hex.DecodeString(event.ID)
				if err != nil || len(eventID) != 32 {
					continue
				}
				_, err = DB.Exec(`DELETE FROM event WHERE id = $1 AND pubkey = $2`, eventID, event.PubKey)
				if err != nil {
					log.Printf("Failed to delete event: %v", event)
					return false, fmt.Sprintf("error: failed to delete: %s", err.Error())
				}
			}
		}
		return true, ""

	case nostr.KindSetMetadata: //0
		// delete past set_metadata events from this user
		_, _ = DB.Exec(`DELETE FROM event WHERE pubkey = $1 AND kind = 0`, event.PubKey)
	case nostr.KindTextNote: //1
		// do nothing
	case nostr.KindRecommendServer: //2
		// delete past recommend_server events equal to this one
		_, _ = DB.Exec(`DELETE FROM event WHERE pubkey = $1 AND kind = 2 AND content = $2`,
			event.PubKey, event.Content)
	case nostr.KindContactList: //3
		// delete past contact lists from this same pubkey
		_, _ = DB.Exec(`DELETE FROM event WHERE pubkey = $1 AND kind = 3`, event.PubKey)
	}

	if 20000 <= event.Kind && event.Kind < 30000 {
		// do not store ephemeral events
	} else {
		tagsList, _ := json.Marshal(event.Tags)
		_, err := DB.Exec(
			`INSERT INTO event (id, pubkey, created_at, kind, tags, content, sig) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (id) do nothing`,
			event.ID,
			event.PubKey,
			event.CreatedAt,
			event.Kind,
			tagsList,
			event.Content,
			event.Sig,
		)
		if err != nil {
			return false, "failed to save event"
		}
	}
	return true, ""
}
