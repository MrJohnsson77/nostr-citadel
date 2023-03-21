package cln

import (
	"encoding/json"
	"fmt"
	lnSocket "github.com/jb55/lnsocket/go"
	"github.com/tidwall/gjson"
	"nostr-citadel/pkg/config"
	"nostr-citadel/pkg/utils"
	"time"
)

type (
	InvoiceRequest struct {
		Amount int64
		Expiry int64
		Memo   string
	}
	InvoiceResponse struct {
		Invoice string
	}
	//InvoiceCheckRequest struct {
	//	Memo string
	//}
)

func GenerateClnInvoice(invoiceRequest *InvoiceRequest) (*InvoiceResponse, error) {
	label := invoiceRequest.Memo
	cln := lnSocket.LNSocket{}
	cln.GenKey()

	err := cln.ConnectAndInit(config.Config.Processing.Cln.Host, config.Config.Processing.Cln.NodeID)
	if err != nil {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Unable to connect to core-lightning: %s", config.Config.Processing.Cln.Host),
			Level:    "ERROR",
		})
		return &InvoiceResponse{}, err
	}
	defer cln.Disconnect()
	// check if there is an invoice already
	jsonParams, _ := json.Marshal(map[string]any{
		"label": label,
	})
	result, _ := cln.Rpc(config.Config.Processing.Cln.Rune, "listinvoices", string(jsonParams))

	if gjson.Get(result, "result.invoices.#").Int() == 1 {
		if CheckClnInvoicePaidOk(label) {
			return &InvoiceResponse{}, fmt.Errorf("this pubkey has already paid")
		}
		timestamp := time.Now().Unix()
		if gjson.Get(result, "result.invoices.0.expires_at").Int() > timestamp {
			return &InvoiceResponse{Invoice: gjson.Get(result, "result.invoices.0.bolt11").String()}, nil
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
		"description": fmt.Sprintf("Write access to %s", config.Config.Relay.Name),
	})
	result, err = cln.Rpc(config.Config.Processing.Cln.Rune, "invoice", string(jsonParams))
	if err != nil {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Unable to create invoice on core-lightning:\n%s", err.Error()),
			Level:    "ERROR",
		})
		return &InvoiceResponse{}, err
	}

	resErr := gjson.Get(result, "error")
	if resErr.Type != gjson.Null {
		if resErr.Type == gjson.JSON {
			utils.Logger(utils.LogEvent{
				Datetime: time.Now(),
				Content:  fmt.Sprintf("Unable to read invoice on core-lightning:\n%s", resErr.Get("message").String()),
				Level:    "ERROR",
			})
			return &InvoiceResponse{}, err
		} else if resErr.Type == gjson.String {
			utils.Logger(utils.LogEvent{
				Datetime: time.Now(),
				Content:  fmt.Sprintf("Unable to read invoice on core-lightning:\n%s", resErr.String()),
				Level:    "ERROR",
			})
			return &InvoiceResponse{}, err
		}
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Unknown commando error on core-lightning:\n%s", resErr.String()),
			Level:    "ERROR",
		})
		return &InvoiceResponse{}, fmt.Errorf("unknown commando error: '%v'", resErr)
	}

	invoice := gjson.Get(result, "result.bolt11")
	if invoice.Type != gjson.String {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("No bolt11 result found in invoice response on core-lightning:\n%s", result),
			Level:    "ERROR",
		})
		return &InvoiceResponse{}, fmt.Errorf("no bolt11 result found in invoice response, got %v", result)
	}

	response := &InvoiceResponse{
		Invoice: invoice.String(),
	}

	return response, nil
}

func CheckClnInvoicePaidOk(label string) bool {
	cln := lnSocket.LNSocket{}
	cln.GenKey()

	err := cln.ConnectAndInit(config.Config.Processing.Cln.Host, config.Config.Processing.Cln.NodeID)
	if err != nil {
		return false
	}
	defer cln.Disconnect()

	jsonParams, _ := json.Marshal(map[string]any{
		"label": label,
	})

	result, _ := cln.Rpc(config.Config.Processing.Cln.Rune, "listinvoices", string(jsonParams))
	return gjson.Get(result, "result.invoices.0.status").String() == "paid"
}
