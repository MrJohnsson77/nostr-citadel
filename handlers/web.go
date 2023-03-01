package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func WebInvoice(c echo.Context) error {
	return c.String(http.StatusOK, "Coming soon.")
}
