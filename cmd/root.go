package cmd

import "github.com/spf13/cobra"

var (
	FavIcon     []byte
	SocketHTML  string
	InvoiceHTML string
	rootCmd     = &cobra.Command{
		Use:           "nostr-citadel",
		Short:         "Nostr Citadel – The Sovereign Relay",
		Long:          `Operate your citadel`,
		SilenceErrors: false,
		SilenceUsage:  false,
	}
)

func Execute(socketHTML string, favIcon []byte, invoiceHTML string) error {
	FavIcon = favIcon
	SocketHTML = socketHTML
	InvoiceHTML = invoiceHTML
	return rootCmd.Execute()
}
