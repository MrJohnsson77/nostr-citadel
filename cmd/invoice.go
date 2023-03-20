package cmd

import (
	"fmt"
	"github.com/mdp/qrterminal/v3"
	"github.com/spf13/cobra"
	"nostr-citadel/pkg/controllers"
	"nostr-citadel/pkg/libs"
	"os"
)

var (
	invoiceCMD = &cobra.Command{
		Use:   "invoice",
		Short: "Manage Invoices",
		Long:  `Create or verify invoices`,
		Run:   cmdInvoice,
	}
)

func cmdInvoice(wlCmd *cobra.Command, args []string) {
	createInvoice, _ := wlCmd.Flags().GetString("create")
	verifyInvoice, _ := wlCmd.Flags().GetString("verify")
	createQr, _ := wlCmd.Flags().GetString("qr")

	if len(createInvoice) > 4 {
		_, npub := controllers.GetPubKey(createInvoice)
		if len(npub) > 4 {
			invoice, err := libs.GenerateClnInvoice(npub)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
			} else {
				fmt.Printf("Invoice: %s\n", invoice)
			}
		} else {
			fmt.Printf("Not a valid pubkey: %s\n", verifyInvoice)
		}
	} else if len(createQr) > 4 {
		_, npub := controllers.GetPubKey(createQr)
		if len(npub) > 4 {
			invoice, err := libs.GenerateClnInvoice(npub)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
			}
			config := qrterminal.Config{
				Level:          qrterminal.L,
				Writer:         os.Stdout,
				HalfBlocks:     true,
				BlackChar:      " ",
				BlackWhiteChar: "▄",
				WhiteChar:      "█",
				WhiteBlackChar: "▀",
				QuietZone:      2,
			}
			qrterminal.GenerateWithConfig(invoice, config)
			fmt.Printf("Invoice: %s\n", invoice)
		}
	} else if len(verifyInvoice) > 4 {
		_, npub := controllers.GetPubKey(verifyInvoice)
		if len(npub) > 4 {
			if libs.CheckClnInvoicePaidOk(npub) {
				fmt.Println("Invoice is paid")
			} else {
				fmt.Println("Invoice is not paid")
			}
		} else {
			fmt.Printf("Not a valid pubkey: %s\n", verifyInvoice)
		}
	} else {
		_ = wlCmd.Help()
		os.Exit(0)
	}

}

func init() {
	rootCmd.AddCommand(invoiceCMD)
	invoiceCMD.Flags().StringP("create", "c", "", "create invoice for npub")
	invoiceCMD.Flags().StringP("verify", "v", "", "verify if npub has paid")
	invoiceCMD.Flags().StringP("qr", "q", "", "create invoice for npub")
}
