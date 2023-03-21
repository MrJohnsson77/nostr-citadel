package models

import (
	"fmt"
	"nostr-citadel/pkg/storage"
	"nostr-citadel/pkg/utils"
	"time"
)

type Invoice struct {
	PubKey    string
	Invoice   string
	InvoiceID string
	Paid      bool
	Amount    int64
	Expires   time.Time
	Created   time.Time
}

func GetInvoice(pubKey string) (invoice Invoice) {
	_ = storage.DB.Get(&invoice, "SELECT pubkey, invoice, paid, invoice_id as invoiceid, amount_msat as amount, expires_at as expires, created_at as created from invoice where pubkey = ?", pubKey)
	return invoice
}

func SetInvoicePaid(pubKey string) {
	_, _ = storage.DB.Exec("UPDATE invoice set paid = 1 where pubkey = ?", pubKey)
}

func InsertInvoice(invoice *Invoice) {
	_, _ = storage.DB.Exec(`INSERT into invoice (pubkey, invoice, expires_at, invoice_id, amount_msat, created_at) values ($1,$2,$3,$4,$5,$6)`, invoice.PubKey, invoice.Invoice, invoice.Expires.Unix(), invoice.InvoiceID, invoice.Amount, time.Now().Unix())
}

func GetUnpaidInvoices() (invoice []Invoice) {
	_ = storage.DB.Select(&invoice, "SELECT pubkey, invoice, paid, invoice_id as invoiceid, amount_msat as amount, expires_at as expires, created_at as created from invoice where not paid")
	return invoice
}

func GetAllInvoices() (invoice []Invoice) {
	_ = storage.DB.Select(&invoice, "SELECT pubkey, invoice, paid, invoice_id as invoiceid, amount_msat as amount, expires_at as expires, created_at as created from invoice")
	return invoice
}

func CountInvoicesToExpire() (count int) {
	rows := storage.DB.QueryRow("SELECT COUNT(pubkey) as count FROM invoice WHERE expires_at < UNIXEPOCH() AND NOT paid")
	_ = rows.Scan(&count)
	return count
}

func DeleteExpiredInvoices() {
	_, err := storage.DB.Exec("DELETE from invoice where expires_at < UNIXEPOCH() AND NOT paid")
	if err != nil {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("CitadelKeeper: Failed to delete expired invoices"),
			Level:    "ERROR",
		})
	}
}
