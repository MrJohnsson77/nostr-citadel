package models

import (
	"nostr-citadel/pkg/config"
	"nostr-citadel/pkg/storage"
	"time"
)

type Admin struct {
	PubKey string
}

func SetAdmin(pubKey string) {
	admin := Admin{}
	_ = storage.DB.Get(&admin, "SELECT pubkey FROM whitelist WHERE pubkey = ? and admin = 1", pubKey)

	if admin.PubKey == pubKey {
		return
	} else {
		_, _ = storage.DB.Exec(`DELETE FROM whitelist where admin = 1`)
		_, _ = storage.DB.Exec(`DELETE FROM event where kind = 0 and pubkey = ?`, admin.PubKey)
		_, _ = storage.DB.Exec(`INSERT into whitelist (pubkey, created_at, sync, last_synced, admin) values ($1,$2,$3,$4,$5)`,
			pubKey, time.Now().Unix(), 1, time.Now().Add(time.Duration(-24*config.Config.Importer.ImportDaysOnInit)*time.Hour).Unix(), 1)
	}
}
