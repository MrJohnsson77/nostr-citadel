package cmd

import (
	"github.com/spf13/cobra"
	"nostr-citadel/pkg/libs"
	"os"
)

var (
	backupCMD = &cobra.Command{
		Use:   "backup",
		Short: "Backup and restore citadel data",
		Long:  `Perform backups and restore citadel data`,
		Run:   cmdBackup,
	}
)

func cmdBackup(iCmd *cobra.Command, args []string) {
	rFileName, _ := iCmd.Flags().GetString("restore")
	backup, _ := iCmd.Flags().GetBool("all")

	if len(rFileName) > 0 {
		libs.RestoreBackup(rFileName)
	} else if backup {
		libs.RunBackup()
	} else {
		_ = iCmd.Help()
		os.Exit(0)
	}
}

func init() {
	rootCmd.AddCommand(backupCMD)
	backupCMD.Flags().BoolP("all", "a", false, "backup everything to files")
	backupCMD.Flags().StringP("restore", "r", "", "restore backup from file")
}
