package libs

import (
	"encoding/json"
	"errors"
	"fmt"
	lnSocket "github.com/jb55/lnsocket/go"
	"github.com/tidwall/gjson"
	"nostr-citadel/pkg/config"
	"time"
)

func generateLabel(pubkey string) string {

	return fmt.Sprintf("citadel:ticket:%s", pubkey)
}

func GenerateClnInvoice(pubkey string) (string, error) {
	label := generateLabel(pubkey)
	cln := lnSocket.LNSocket{}
	cln.GenKey()

	err := cln.ConnectAndInit(config.Config.Processing.Cln.Host, config.Config.Processing.Cln.NodeID)
	if err != nil {
		return "", err
	}
	defer cln.Disconnect()

	// check if there is an invoice already
	jsonParams, _ := json.Marshal(map[string]any{
		"label": label,
	})
	result, _ := cln.Rpc(config.Config.Processing.Cln.Rune, "listinvoices", string(jsonParams))

	if gjson.Get(result, "result.invoices.#").Int() == 1 {
		if CheckClnInvoicePaidOk(pubkey) {
			return "", fmt.Errorf("this npub has already paid")
		}
		timestamp := time.Now().Unix()
		if gjson.Get(result, "result.invoices.0.expires_at").Int() > timestamp {
			return gjson.Get(result, "result.invoices.0.bolt11").String(), nil
		}
		jsonParams, _ := json.Marshal(map[string]any{
			"label":  label,
			"status": "expired",
		})
		_, _ = cln.Rpc(config.Config.Processing.Cln.Rune, "delinvoice", string(jsonParams))
	}

	jsonParams, _ = json.Marshal(map[string]any{
		"amount_msat": config.Config.Relay.TicketPrice * 1000,
		"label":       label,
		"description": fmt.Sprintf("%s's ticket for writing to %s", pubkey, config.Config.Relay.Name),
	})
	result, err = cln.Rpc(config.Config.Processing.Cln.Rune, "invoice", string(jsonParams))
	if err != nil {
		return "", err
	}

	resErr := gjson.Get(result, "error")
	if resErr.Type != gjson.Null {
		if resErr.Type == gjson.JSON {
			return "", errors.New(resErr.Get("message").String())
		} else if resErr.Type == gjson.String {
			return "", errors.New(resErr.String())
		}
		return "", fmt.Errorf("unknown commando error: '%v'", resErr)
	}

	invoice := gjson.Get(result, "result.bolt11")
	if invoice.Type != gjson.String {
		return "", fmt.Errorf("no bolt11 result found in invoice response, got %v", result)
	}

	return invoice.String(), nil
}

func CheckClnInvoicePaidOk(pubkey string) bool {
	cln := lnSocket.LNSocket{}
	cln.GenKey()

	err := cln.ConnectAndInit(config.Config.Processing.Cln.Host, config.Config.Processing.Cln.NodeID)
	if err != nil {
		return false
	}
	defer cln.Disconnect()

	jsonParams, _ := json.Marshal(map[string]any{
		"label": generateLabel(pubkey),
	})

	result, _ := cln.Rpc(config.Config.Processing.Cln.Rune, "listinvoices", string(jsonParams))
	return gjson.Get(result, "result.invoices.0.status").String() == "paid"
}
