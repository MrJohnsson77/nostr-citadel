package routes

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"nostr-citadel/pkg/controllers"
	"nostr-citadel/pkg/libs"
)

type relayInvoice struct {
	Bolt11 string `json:"bolt11"`
	Error  string `json:"error"`
}

func GetInvoice(c echo.Context) error {
	pubKey := c.QueryParam("pubkey")
	_, npub := controllers.GetPubKey(pubKey)
	if len(npub) < 30 {
		payment := &relayInvoice{Bolt11: "", Error: "Invalid Npub"}
		return c.JSON(http.StatusOK, payment)
	} else {
		invoice, err := libs.GenerateClnInvoice(npub)
		if err != nil {
			payment := &relayInvoice{Bolt11: "", Error: err.Error()}
			return c.JSON(http.StatusOK, payment)
		}
		payment := &relayInvoice{Bolt11: invoice, Error: ""}
		return c.JSON(http.StatusOK, payment)
	}
}
