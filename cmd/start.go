package cmd

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/olahol/melody"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"nostr-citadel/pkg/config"
	"nostr-citadel/pkg/controllers"
	"nostr-citadel/pkg/models"
	"nostr-citadel/pkg/routes"
	"nostr-citadel/pkg/storage"
	"nostr-citadel/pkg/workers"
)

var (
	cmd1 = &cobra.Command{
		Use:   "start",
		Short: "Start the relay",
		Long:  ``,
		Run:   cmdStart,
	}
)

func cmdStart(startCmd *cobra.Command, args []string) {
	workers.GetDefaultRelays()

	serverPort, _ := startCmd.Flags().GetString("port")
	err := storage.InitDB()
	if err != nil {
		return
	}

	admPk := config.Config.Admin.Npub
	pubKey, _ := controllers.GetPubKey(admPk)
	if len(pubKey) < 4 {
		fmt.Printf("Enter a valid admin_npub in config.yaml (%s)\n", admPk)
		return
	}
	models.SetAdmin(pubKey)

	e := echo.New()
	m := melody.New()
	m.Config.MessageBufferSize = 512
	m.Config.MaxMessageSize = 512000

	if config.Config.Relay.BehindProxy {
		e.IPExtractor = echo.ExtractIPFromXFFHeader()
	}

	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))
	e.Use(middleware.Recover())

	controllers.SocketMessageHandler(m)

	e.GET("/", func(c echo.Context) error {
		c.Request().RemoteAddr = c.RealIP()
		if c.IsWebSocket() {
			_ = m.HandleRequest(c.Response().Writer, c.Request())
			return nil
		} else {
			nip11Request := c.Request().Header.Get("accept") == "application/nostr+json"
			if nip11Request {
				return c.JSON(http.StatusOK, controllers.NostrNip11())
			} else {
				if config.Config.Dashboard.Enabled {
					return c.HTML(http.StatusOK, SocketHTML)
				} else {
					return c.String(http.StatusOK, "Nostr Citadel")
				}
			}
		}
	})

	if config.Config.Relay.PaidRelay {
		e.GET("/invoices", func(c echo.Context) error {
			return c.HTML(http.StatusOK, InvoiceHTML)
		})
		e.GET("/invoice", func(c echo.Context) error {
			return routes.GetInvoice(c)
		})
	}

	e.GET("/favicon.ico", func(c echo.Context) error {
		return c.Blob(http.StatusOK, "image/ico", FavIcon)
	})

	workers.StartImporter()
	e.Logger.Fatal(e.Start(":" + serverPort))
}

func init() {
	rootCmd.AddCommand(cmd1)
	rootCmd.PersistentFlags().StringP("loglevel", "", "ERROR", "Log level (\"DEBUG,INFO,WARN,ERROR\")")
	_ = viper.BindPFlag("loglevel", rootCmd.PersistentFlags().Lookup("loglevel"))
	cmd1.Flags().StringP("port", "p", "1337", "specify task title / heading")
}
