package workers

import (
	"encoding/hex"
	"fmt"
	"github.com/nbd-wtf/go-nostr/nip19"
	"nostr-citadel/pkg/config"
	"nostr-citadel/pkg/libs/processing/cln"
	"nostr-citadel/pkg/libs/processing/lnd"
	"nostr-citadel/pkg/models"
	"nostr-citadel/pkg/utils"
	"strings"
	"sync"
	"time"
)

func StartHouseKeeper() {

	go func() {
		time.Sleep(10 * time.Second)
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("CitadelKeeper: Started The CitadelKeeper"),
			Level:    "INFO",
		})
	}()

	if len(config.Config.Processing.Processor) == 3 {
		go func() {
			for {
				time.Sleep(time.Minute * 1)
				cleanExpiredInvoices()
			}
		}()

		go func() {
			for {
				time.Sleep(time.Second * 10)
				checkPayments()
			}
		}()
	}
}

func cleanExpiredInvoices() {
	totalInvoices := models.CountInvoicesToExpire()
	if totalInvoices > 0 {
		models.DeleteExpiredInvoices()
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("CitadelKeeper: Deleted %d expired invoice(s)", totalInvoices),
			Level:    "INFO",
		})
	}
}

func checkPayments() {
	invoices := models.GetUnpaidInvoices()
	const workers = 5
	queue := make(chan models.Invoice)
	wg := &sync.WaitGroup{}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go checkPaymentWorker(wg, queue)
	}

	for _, invoice := range invoices {
		queue <- invoice
	}
	close(queue)
	wg.Wait()
}

func checkPaymentWorker(wg *sync.WaitGroup, queue chan models.Invoice) {
	defer wg.Done()

	for invoice := range queue {

		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("CitadelKeeper: Checking Payment on Invoice:\n%v", invoice),
			Level:    "DEBUG",
		})

		var paid = false
		switch config.Config.Processing.Processor {
		case "cln":
			paid = cln.CheckClnInvoicePaidOk(invoice.InvoiceID)
			break
		case "lnd":
			invoiceHash, err := hex.DecodeString(invoice.InvoiceID)
			if err != nil {
				fmt.Println("Unable to get LND RHASH ", err)
			}
			invoiceCheck := &lnd.CheckInvoiceRequest{RHASH: invoiceHash}
			paid = lnd.CheckInvoicePaid(invoiceCheck)
			break
		}

		if paid {
			npub, _ := nip19.EncodePublicKey(invoice.PubKey)
			utils.Logger(utils.LogEvent{
				Datetime: time.Now(),
				Content:  fmt.Sprintf("CitadelKeeper: %s Invoice - %d sats paid by %s", strings.ToUpper(config.Config.Processing.Processor), invoice.Amount/1000, npub),
				Level:    "INFO",
			})
			models.SetInvoicePaid(invoice.PubKey)
			models.AddWhitelist(invoice.PubKey)
		}

	}
}
