package models

import (
	"nostr-citadel/pkg/config"
	"nostr-citadel/pkg/storage"
	"time"
)

type BWL struct {
	PubKey  string
	Admin   bool
	Sync    bool
	Created time.Time
}

type BEV struct {
	ID      string
	PubKey  string
	Kind    string
	Content string
	Tags    string
	Sig     string
	Created time.Time
}

type BIN struct {
	PubKey    string
	Invoice   string
	InvoiceID string
	Paid      bool
	Amount    int64
	Expires   time.Time
	Created   time.Time
}

func BackupWhitelist() (bl []BWL) {
	_ = storage.DB.Select(&bl, "SELECT pubkey,admin,sync,created_at as created FROM whitelist")
	return bl
}

func BackupEvents() (bl []BEV) {
	if config.Config.Backup.AdminOnly {
		_ = storage.DB.Select(&bl, "SELECT e.id,e.pubkey,e.kind,e.tags,e.content,e.sig,e.created_at as created FROM event e LEFT JOIN whitelist w on e.pubkey = w.pubkey where w.admin")
	} else {
		_ = storage.DB.Select(&bl, "SELECT id,pubkey,kind,tags,content,sig,created_at as created FROM event")
	}
	return bl
}

func BackupInvoices() (bl []BIN) {
	_ = storage.DB.Select(&bl, "SELECT pubkey,invoice,invoice_id as invoiceid,paid,amount_msat as amount,expires_at as expires,created_at as created FROM invoice")
	return bl
}
