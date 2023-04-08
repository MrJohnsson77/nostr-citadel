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
	Admin      bool
}

func SetLastSync(pubKey string) {
	_, _ = storage.DB.Exec(`UPDATE whitelist set last_synced = $1 WHERE pubkey = $2`, time.Now().Unix(), pubKey)
}

func GetPlebsToSync() (plebs []Pleb, err error) {
	wlSync := config.Config.Importer.ImportWhitelisted
	query := `SELECT wl.pubkey as pubkey,coalesce(content,'') as relays,
    coalesce(last_synced,1672527600000) as lastsynced, admin from whitelist wl LEFT JOIN (SELECT * from event WHERE kind = 3) as e on wl.pubkey = e.pubkey`
	if wlSync {
		err = storage.DB.Select(&plebs, query)
	} else {
		err = storage.DB.Select(&plebs, query+" WHERE wl.admin = 1")
	}

	return plebs, err
}
