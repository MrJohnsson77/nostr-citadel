package processing

import (
	"encoding/hex"
	"fmt"
	"nostr-citadel/pkg/config"
	"nostr-citadel/pkg/libs/processing/cln"
	"nostr-citadel/pkg/libs/processing/lnd"
	"nostr-citadel/pkg/models"
	"nostr-citadel/pkg/utils"
	"strings"
	"time"
)

type (
	QueryCheckInvoice struct {
		Npub string
	}
	QueryCreateInvoice struct {
		Npub string
	}
	ResponseCreateInvoice struct {
		Invoice string
	}
	ResponseCheckInvoice struct {
		Paid bool
	}
)

func generateMemo(npub string) (memo string) {
	label := strings.ToLower(strings.ReplaceAll(config.Config.Relay.Name, " ", "-"))
	memo = fmt.Sprintf("%s:ticket:%s", label, npub)
	return memo
}

func CreateInvoice(npub string, pubKey string) (string, error) {

	// Todo: Check if already exists on npub
	invoice := models.GetInvoice(pubKey)

	if len(invoice.Invoice) > 0 {
		if invoice.Paid {
			return "", fmt.Errorf("this pubkey has already paid")
		} else {
			return invoice.Invoice, nil
		}
	}

	ticketPrice := config.Config.Relay.TicketPrice * 1000
	ticketExpiry := config.Config.Relay.TicketExpiry

	memo := generateMemo(npub)

	switch config.Config.Processing.Processor {
	case "cln":
		invoiceReq := &cln.InvoiceRequest{
			Amount: ticketPrice,
			Expiry: ticketExpiry,
			Memo:   memo,
		}
		clnInvoice, err := cln.GenerateClnInvoice(invoiceReq)
		if err != nil {
			return "", err
		}

		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("CLN: Issuing %d sats invoice for: %s", config.Config.Relay.TicketPrice, npub),
			Level:    "INFO",
		})

		inv := &models.Invoice{
			PubKey:    pubKey,
			Invoice:   clnInvoice.Invoice,
			InvoiceID: memo,
			Paid:      false,
			Amount:    ticketPrice,
			Expires:   time.Now().Add(time.Duration(ticketExpiry) * time.Second),
			Created:   time.Now(),
		}

		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("CLN Invoice:\n%v", inv),
			Level:    "DEBUG",
		})

		models.InsertInvoice(inv)
		return clnInvoice.Invoice, nil
	case "lnd":
		invoiceReq := &lnd.InvoiceRequest{
			Amount: ticketPrice,
			Expiry: ticketExpiry,
			Memo:   memo,
		}
		lndInvoice, err := lnd.CreateInvoice(invoiceReq)
		if err != nil {
			return "", err
		}

		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("LND: Issuing %d sats invoice for: %s", config.Config.Relay.TicketPrice, npub),
			Level:    "INFO",
		})

		invoiceID := hex.EncodeToString(lndInvoice.Invoice.RHash)

		inv := &models.Invoice{
			PubKey:    pubKey,
			Invoice:   lndInvoice.Invoice.PaymentRequest,
			InvoiceID: invoiceID,
			Paid:      false,
			Amount:    ticketPrice,
			Expires:   time.Now().Add(time.Duration(ticketExpiry) * time.Second),
			Created:   time.Now(),
		}

		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("LND Invoice:\n%v", inv),
			Level:    "DEBUG",
		})

		models.InsertInvoice(inv)
		return lndInvoice.Invoice.PaymentRequest, nil
	default:
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("No payment processor configured"),
			Level:    "ERROR",
		})
		return "", fmt.Errorf("no payment processor configured")
	}
}

func CheckIfPaid(pubKey string) bool {
	invoice := models.GetInvoice(pubKey)
	return invoice.Paid
}
