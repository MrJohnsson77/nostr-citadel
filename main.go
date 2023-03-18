package main

import (
	_ "embed"
	"fmt"
	"github.com/spf13/viper"
	"nostr-citadel/cmd"
	"nostr-citadel/pkg/config"
	"nostr-citadel/pkg/storage"
	"os"
)

var (
	invoiceHTML string = inv
	//go:embed "public/invoice.html"
	inv string

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
	viper.SetConfigFile("config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("No config.yaml file found!")
		os.Exit(0)
	}

	config.SetConf()
	_ = storage.InitDB()

	err = cmd.Execute(socketHTML, favIcon, invoiceHTML)
	if err != nil && err.Error() != "" {
		fmt.Println(err)
	}

	os.Exit(0)
}
