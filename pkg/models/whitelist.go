package models

import (
	"nostr-citadel/pkg/config"
	"nostr-citadel/pkg/storage"
	"time"
)

type Whitelisted struct {
	PubKey  string
	Admin   bool
	Created time.Time
}

func GetWhitelisted(pubKey string) (wl Whitelisted) {
	_ = storage.DB.Get(&wl, "SELECT pubkey,admin,created_at as created FROM whitelist WHERE pubkey = ?", pubKey)
	return wl
}

func AddWhitelist(pubKey string) {
	_, _ = storage.DB.Exec(`INSERT into whitelist (pubkey, created_at, sync, last_synced, admin) values ($1,$2,$3,$4,$5) on conflict (pubkey) DO NOTHING`,
		pubKey, time.Now().Unix(), 0, time.Now().Add(time.Duration(-24*config.Config.Importer.ImportDaysOnInit)*time.Hour).Unix(), 0)
}

func RemoveWhitelist(pubKey string) {
	_, _ = storage.DB.Exec(`DELETE from whitelist where pubkey = $1 and not admin`, pubKey)
	_, _ = storage.DB.Exec(`DELETE from event where pubkey = $1`, pubKey)
}

func GetWhitelist() (wl []Whitelisted) {
	_ = storage.DB.Select(&wl, "SELECT pubkey,admin,created_at as created FROM whitelist")
	return wl
}
