package cmd

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/mdp/qrterminal/v3"
	"github.com/spf13/cobra"
	"nostr-citadel/pkg/controllers"
	"nostr-citadel/pkg/libs/processing"
	"nostr-citadel/pkg/models"
	"os"
	"time"
)

var (
	invoiceCMD = &cobra.Command{
		Use:   "invoice",
		Short: "Manage Invoices",
		Long:  `Create or verify invoices`,
		Run:   cmdInvoice,
	}
)

func cmdInvoice(iCmd *cobra.Command, args []string) {
	createInvoice, _ := iCmd.Flags().GetString("create")
	verifyInvoice, _ := iCmd.Flags().GetString("verify")
	createQr, _ := iCmd.Flags().GetString("qr")
	listInvoices, _ := iCmd.Flags().GetBool("list")

	if len(createInvoice) > 4 {
		pk, npub := controllers.GetPubKey(createInvoice)
		if len(npub) > 4 {
			invoice, err := processing.CreateInvoice(npub, pk)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
			} else {
				fmt.Printf("Invoice: %s\n", invoice)
			}
		} else {
			fmt.Printf("Not a valid pubkey: %s\n", verifyInvoice)
		}
	} else if len(createQr) > 4 {
		pk, npub := controllers.GetPubKey(createQr)
		if len(npub) > 4 {
			invoice, err := processing.CreateInvoice(npub, pk)
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
		pk, _ := controllers.GetPubKey(verifyInvoice)
		if len(pk) > 4 {

			paid := processing.CheckIfPaid(pk)

			if paid {
				fmt.Println("Invoice is paid")
			} else {
				fmt.Println("Invoice is not paid")
			}

		} else {
			fmt.Printf("Not a valid pubkey: %s\n", verifyInvoice)
		}
	} else if listInvoices {
		invoices := models.GetAllInvoices()
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleLight)
		t.AppendHeader(table.Row{"npub", "amount sats", "date added", "paid"})
		total := int64(0)
		for _, invoice := range invoices {
			total += invoice.Amount / 1000
			_, npub := controllers.GetPubKey(invoice.PubKey)
			t.AppendRows([]table.Row{
				{npub, invoice.Amount / 1000, invoice.Created.Format(time.RFC822), invoice.Paid},
			})
		}
		t.AppendFooter(table.Row{fmt.Sprintf("Total Invoices (%d)", len(invoices)), total, ""})
		t.Render()
	} else {
		_ = iCmd.Help()
		os.Exit(0)
	}

}

func init() {
	rootCmd.AddCommand(invoiceCMD)
	invoiceCMD.Flags().StringP("create", "c", "", "create invoice for npub")
	invoiceCMD.Flags().StringP("verify", "v", "", "verify if npub has paid")
	invoiceCMD.Flags().StringP("qr", "q", "", "create invoice for npub")
	invoiceCMD.Flags().BoolP("list", "l", false, "list invoices")
}
