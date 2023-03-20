package models

import (
	"nostr-citadel/pkg/config"
	"nostr-citadel/pkg/storage"
	"time"
)

type Pleb struct {
	PubKey     string
	Relays     string
	LastSynced int64
}

func SetLastSync(pubKey string) {
	_, _ = storage.DB.Exec(`UPDATE whitelist set last_synced = $1 WHERE pubkey = $2`, time.Now().Unix(), pubKey)
}

func GetPlebsToSync() (plebs []Pleb, err error) {
	var wl []Pleb

	wlSync := config.Config.Importer.ImportWhitelisted
	query := `SELECT wl.pubkey as pubkey,coalesce(content,'') as relays,
    coalesce(last_synced,1672527600000) as lastsynced from whitelist wl LEFT JOIN (SELECT * from event WHERE kind = 3) as e on wl.pubkey = e.pubkey`
	if wlSync {
		err = storage.DB.Select(&wl, query)
	} else {
		err = storage.DB.Select(&wl, query+" WHERE wl.sync = 1")
	}

	return wl, err
}
