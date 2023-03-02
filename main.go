package main

import (
	_ "embed"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/olahol/melody"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"net/http"
	"nostr-citadel/handlers"
	"nostr-citadel/storage"
	"nostr-citadel/workers"
	"os"
)

var (
	socketHTML string = html
	//go:embed "public/index.html"
	html string

	favIcon = icon
	//go:embed "public/favicon.ico"
	icon []byte
)

func main() {
	viper.AddConfigPath(".")
	viper.SetConfigType("yml")
	viper.SetConfigFile("config.yml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("No config.yml file found!")
		os.Exit(0)
	}

	pflag.String("port", "1337", "Port to listen on")
	pflag.String("whitelist-add", "npub", "Add npub to whitelist")
	pflag.String("whitelist-rem", "npub", "Remove npub from whitelist")

	pflag.Parse()
	_ = viper.BindPFlags(pflag.CommandLine)

	serverPort := viper.GetString("port")
	whitelistAdd := viper.GetString("whitelist-add")
	whitelistRem := viper.GetString("whitelist-rem")

	// Init Storage
	err = storage.InitDB()
	if err != nil {
		return
	}

	admPk := viper.GetString("relay_config.admin_npub")

	pubKey, _ := handlers.GetPubKey(admPk)
	if len(pubKey) < 4 {
		fmt.Printf("Enter a valid admin_npub in config.yml (%s)\n", admPk)
		return
	}
	storage.SetAdmin(pubKey)

	if len(whitelistAdd) > 4 && whitelistAdd != "npub" {
		pubKey, _ := handlers.GetPubKey(whitelistAdd)
		if len(pubKey) > 4 {
			storage.AddWhitelist(pubKey)
			fmt.Printf("Added pubkey %s to whitelist \n", whitelistAdd)
		} else {
			fmt.Printf("Not a valid pubkey: %s\n", whitelistAdd)
		}
		os.Exit(0)
	}

	if len(whitelistRem) > 4 && whitelistRem != "npub" {
		pubKey, _ := handlers.GetPubKey(whitelistRem)
		if len(pubKey) > 4 {
			storage.RemoveWhitelist(pubKey)
			fmt.Printf("Removed pubkey %s from whitelist \n", whitelistRem)
		} else {
			fmt.Printf("Not a valid pubkey: %s\n", whitelistRem)
		}
		os.Exit(0)
	}

	// Start Web Server & Socket
	e := echo.New()
	m := melody.New()
	m.Config.MessageBufferSize = 512
	m.Config.MaxMessageSize = 512000

	if viper.Get("relay_config.behind_proxy").(bool) {
		e.IPExtractor = echo.ExtractIPFromXFFHeader()
	}

	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		if c.IsWebSocket() {
			_ = m.HandleRequest(c.Response().Writer, c.Request())
			return nil
		} else {
			nip11Request := c.Request().Header.Get("accept") == "application/nostr+json"
			if nip11Request {
				return c.JSON(http.StatusOK, handlers.NostrNip11())
			} else {
				if viper.GetBool("relay_config.dashboard") {
					return c.HTML(http.StatusOK, socketHTML)
				} else {
					return c.String(http.StatusOK, "Nostr Citadel")
				}
			}
		}
	})
	e.GET("/invoices", handlers.WebInvoice)
	e.GET("/favicon.ico", func(c echo.Context) error {
		return c.Blob(http.StatusOK, "image/ico", favIcon)
	})

	handlers.SocketMessageHandler(m)
	workers.StartImporter()

	e.Logger.Fatal(e.Start(":" + serverPort))
}
