package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

type CitadelConfig struct {
	Relay          Relay      `mapstructure:"relay"`
	Admin          Admin      `mapstructure:"admin"`
	Dashboard      Dashboard  `mapstructure:"dashboard"`
	Processing     Processing `mapstructure:"processing"`
	Importer       Importer   `mapstructure:"importer"`
	Database       Database   `mapstructure:"database"`
	BootstrapRelay string     `mapstructure:"bootstrap_relay"`
}
type Limits struct {
	ResponseEventLimit int `mapstructure:"response_event_limit"`
}
type Relay struct {
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
	RelayURL    string `mapstructure:"relay_url"`
	BehindProxy bool   `mapstructure:"behind_proxy"`
	PublicRelay bool   `mapstructure:"public_relay"`
	PaidRelay   bool   `mapstructure:"paid_relay"`
	TicketPrice int    `mapstructure:"ticket_price"`
	Limits      Limits `mapstructure:"limits"`
}
type Admin struct {
	Npub  string `mapstructure:"npub"`
	Email string `mapstructure:"email"`
}
type Dashboard struct {
	Enabled bool `mapstructure:"enabled"`
}
type Cln struct {
	Active bool   `mapstructure:"active"`
	NodeID string `mapstructure:"node_id"`
	Rune   string `mapstructure:"rune"`
	Host   string `mapstructure:"host"`
}
type Lnd struct {
	Active   bool   `mapstructure:"active"`
	NodeID   string `mapstructure:"node_id"`
	Macaroon string `mapstructure:"macaroon"`
	Host     string `mapstructure:"host"`
}
type Processing struct {
	Cln Cln `mapstructure:"cln"`
	Lnd Lnd `mapstructure:"lnd"`
}
type Importer struct {
	Workers           int  `mapstructure:"workers"`
	Fetchers          int  `mapstructure:"fetchers"`
	ImportDaysOnInit  int  `mapstructure:"import_days_on_init"`
	ImportWhitelisted bool `mapstructure:"import_whitelisted"`
}
type Database struct {
	Name string `mapstructure:"name"`
}

var (
	Config *CitadelConfig
)

func SetConf() {
	conf := &CitadelConfig{}
	err := viper.Unmarshal(conf)
	if err != nil {
		fmt.Printf("Error: unable to decode the config, %v", err)
		os.Exit(1)
	}
	Config = conf
}
