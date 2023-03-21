package cmd

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"nostr-citadel/pkg/controllers"
	"nostr-citadel/pkg/models"
	"os"
	"time"
)

var (
	cmd2 = &cobra.Command{
		Use:    "whitelist",
		Short:  "Manage whitelist",
		Long:   `Add Npub or PubKey Hex to the whitelist`,
		Run:    cmdWhitelist,
		Hidden: false,
	}
)

func cmdWhitelist(wlCmd *cobra.Command, args []string) {
	whitelistAdd, _ := wlCmd.Flags().GetString("add")
	whitelistRem, _ := wlCmd.Flags().GetString("remove")
	whitelist, _ := wlCmd.Flags().GetBool("list")

	if len(whitelistAdd) > 4 || len(whitelistRem) > 4 || whitelist {
		if len(whitelistAdd) > 4 && whitelistAdd != "npub" {
			pubKey, _ := controllers.GetPubKey(whitelistAdd)
			if len(pubKey) > 4 {
				models.AddWhitelist(pubKey)
				fmt.Printf("Added pubkey %s to whitelist \n", whitelistAdd)
			} else {
				fmt.Printf("Not a valid pubkey: %s\n", whitelistAdd)
			}
		}

		if len(whitelistRem) > 4 && whitelistRem != "npub" {
			pubKey, _ := controllers.GetPubKey(whitelistRem)
			if len(pubKey) > 4 {
				models.RemoveWhitelist(pubKey)
				fmt.Printf("Removed pubkey %s from whitelist \n", whitelistRem)
			} else {
				fmt.Printf("Not a valid pubkey: %s\n", whitelistRem)
			}
		}

		if whitelist {
			whitelist := models.GetWhitelist()
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.SetStyle(table.StyleLight)
			t.AppendHeader(table.Row{"npub", "date added", "admin"})
			for _, whitelisted := range whitelist {
				_, npub := controllers.GetPubKey(whitelisted.PubKey)
				t.AppendRows([]table.Row{
					{npub, whitelisted.Created.Format(time.RFC822), whitelisted.Admin},
				})
			}
			t.AppendFooter(table.Row{fmt.Sprintf("Total Whitelisted (%d)", len(whitelist)), "", ""})
			t.Render()
		}

	} else {
		_ = wlCmd.Help()
		os.Exit(0)
	}

}

func init() {
	rootCmd.AddCommand(cmd2)
	cmd2.Flags().StringP("add", "a", "", "npub / hex to add")
	cmd2.Flags().StringP("remove", "r", "", "npub / hex to remove")
	cmd2.Flags().BoolP("list", "l", false, "list whitelist")
}
