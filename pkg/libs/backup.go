package libs

import (
	"encoding/csv"
	"fmt"
	"nostr-citadel/pkg/config"
	"nostr-citadel/pkg/models"
	"nostr-citadel/pkg/storage"
	"nostr-citadel/pkg/utils"
	"os"
	"strconv"
	"strings"
	"time"
)

func RunBackup() {
	yesterday := time.Now().Add(-24 * time.Hour)
	fmtDate := yesterday.Format("20060102")
	backupDir := config.Config.Backup.Location

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: Backup directory '%s' does not exist, trying to create it.", backupDir),
			Level:    "ERROR",
		})

		err := os.MkdirAll(backupDir, os.ModePerm)
		if err != nil {
			utils.Logger(utils.LogEvent{
				Datetime: time.Now(),
				Content:  fmt.Sprintf("Backup: Backup directory '%s' could not be created, skipping backup.", backupDir),
				Level:    "ERROR",
			})
			return
		} else {
			utils.Logger(utils.LogEvent{
				Datetime: time.Now(),
				Content:  fmt.Sprintf("Backup: Backup directory '%s' created.", backupDir),
				Level:    "INFO",
			})
		}
	}

	backupWhitelist(backupDir, fmtDate)
	backupEvents(backupDir, fmtDate)
	backupInvoices(backupDir, fmtDate)

}

func removeFile(backupDir string, category string) {
	duration := time.Duration(config.Config.Backup.KeepDays * 72)
	deleteDate := time.Now().Add(-duration * time.Hour)
	fmtDate := deleteDate.Format("20060102")
	utils.Logger(utils.LogEvent{
		Datetime: time.Now(),
		Content:  fmt.Sprintf("Backup: Removing backups older than 72 hours"),
		Level:    "DEBUG",
	})
	backupFile := fmt.Sprintf("%s/citadel-backup-%s-%s.csv", backupDir, category, fmtDate)
	e := os.Remove(backupFile)
	if e != nil {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: No expired %s backup file to remove.", category),
			Level:    "DEBUG",
		})
	}
}

func backupWhitelist(backupDir string, yesterdayMonthDay string) {

	backupFile := fmt.Sprintf("%s/citadel-backup-whitelist-%s.csv", backupDir, yesterdayMonthDay)

	if _, err := os.Stat(backupFile); err == nil {
		return
	}

	utils.Logger(utils.LogEvent{
		Datetime: time.Now(),
		Content:  fmt.Sprintf("Backup: Starting backup for whitelist."),
		Level:    "INFO",
	})

	removeFile(backupDir, "whitelist")

	bwl := models.BackupWhitelist()

	file, err := os.Create(backupFile)
	defer file.Close()
	if err != nil {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: Could not create backup file: %s", backupFile),
			Level:    "ERROR",
		})
		return
	}

	w := csv.NewWriter(file)
	defer w.Flush()
	for _, record := range bwl {
		row := []string{record.PubKey, record.Created.String(), strconv.FormatBool(record.Sync), strconv.FormatBool(record.Admin)}
		if err := w.Write(row); err != nil {
			utils.Logger(utils.LogEvent{
				Datetime: time.Now(),
				Content:  fmt.Sprintf("Backup: Could not write data to CSV file: %s\n%v", backupFile, err),
				Level:    "ERROR",
			})
		}
	}

	utils.Logger(utils.LogEvent{
		Datetime: time.Now(),
		Content:  fmt.Sprintf("Backup: Completed backup of %d whitelisted npubs - %s", len(bwl), backupFile),
		Level:    "INFO",
	})

}

func backupEvents(backupDir string, yesterdayMonthDay string) {
	backupFile := fmt.Sprintf("%s/citadel-backup-events-%s.csv", backupDir, yesterdayMonthDay)

	if _, err := os.Stat(backupFile); err == nil {
		return
	}

	utils.Logger(utils.LogEvent{
		Datetime: time.Now(),
		Content:  fmt.Sprintf("Backup: Starting backup for events."),
		Level:    "INFO",
	})

	removeFile(backupDir, "events")

	bwl := models.BackupEvents()

	file, err := os.Create(backupFile)
	defer file.Close()
	if err != nil {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: Could not create backup file: %s", backupFile),
			Level:    "ERROR",
		})
		return
	}

	w := csv.NewWriter(file)
	defer w.Flush()
	for _, record := range bwl {
		row := []string{record.ID, record.PubKey, record.Kind, record.Content, record.Tags, record.Sig, record.Created.String()}
		if err := w.Write(row); err != nil {
			utils.Logger(utils.LogEvent{
				Datetime: time.Now(),
				Content:  fmt.Sprintf("Backup: Could not write data to CSV file: %s\n%v", backupFile, err),
				Level:    "ERROR",
			})
		}
	}

	utils.Logger(utils.LogEvent{
		Datetime: time.Now(),
		Content:  fmt.Sprintf("Backup: Completed backup of %d events - %s", len(bwl), backupFile),
		Level:    "INFO",
	})

}
func backupInvoices(backupDir string, yesterdayMonthDay string) {
	backupFile := fmt.Sprintf("%s/citadel-backup-invoices-%s.csv", backupDir, yesterdayMonthDay)

	if _, err := os.Stat(backupFile); err == nil {
		return
	}

	utils.Logger(utils.LogEvent{
		Datetime: time.Now(),
		Content:  fmt.Sprintf("Backup: Starting backup for invoices."),
		Level:    "INFO",
	})

	removeFile(backupDir, "invoices")

	bwl := models.BackupInvoices()

	file, err := os.Create(backupFile)
	defer file.Close()
	if err != nil {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: Could not create backup file: %s", backupFile),
			Level:    "ERROR",
		})
		return
	}

	w := csv.NewWriter(file)
	defer w.Flush()
	for _, record := range bwl {
		row := []string{record.PubKey, record.Invoice, record.InvoiceID, strconv.FormatBool(record.Paid), strconv.FormatInt(record.Amount, 10), record.Expires.String(), record.Created.String()}
		if err := w.Write(row); err != nil {
			utils.Logger(utils.LogEvent{
				Datetime: time.Now(),
				Content:  fmt.Sprintf("Backup: Could not write data to CSV file: %s\n%v", backupFile, err),
				Level:    "ERROR",
			})
		}
	}

	utils.Logger(utils.LogEvent{
		Datetime: time.Now(),
		Content:  fmt.Sprintf("Backup: Completed backup of %d invoices - %s", len(bwl), backupFile),
		Level:    "INFO",
	})

}

func RestoreBackup(backupFile string) {
	file, err := os.Open(backupFile)
	if err != nil {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: Can't open the backupfile '%s'", backupFile),
			Level:    "INFO",
		})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: Can't read the backupfile '%s'", backupFile),
			Level:    "INFO",
		})
		return
	}

	if strings.Contains(backupFile, "-events-") {

		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: Preparing to restore %d events from '%s'", len(records), backupFile),
			Level:    "INFO",
		})

		stmt, err := storage.DB.Prepare("INSERT INTO event(id,pubkey,kind,tags,content,sig,created_at) values(?,?,?,?,?,?,?) ON CONFLICT DO NOTHING")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer stmt.Close()
		for _, record := range records {
			_, err = stmt.Exec(record[0], record[1], record[2], record[3], record[4], record[5], record[6])
			if err != nil {
				utils.Logger(utils.LogEvent{
					Datetime: time.Now(),
					Content:  fmt.Sprintf("Backup: Failed to restore events:\n%v.", err),
					Level:    "ERROR",
				})
				return
			}
		}

		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: Restored %d events sucessfully.", len(records)),
			Level:    "INFO",
		})

	} else if strings.Contains(backupFile, "-whitelist-") {

		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: Preparing to restore %d whitelists from '%s'", len(records), backupFile),
			Level:    "INFO",
		})

		stmt, err := storage.DB.Prepare("INSERT INTO whitelist(pubkey,admin,sync,created_at) values(?,?,?,?) ON CONFLICT DO NOTHING")
		if err != nil {
			utils.Logger(utils.LogEvent{
				Datetime: time.Now(),
				Content:  fmt.Sprintf("Backup: Failed to restore whitelists:\n%v.", err),
				Level:    "ERROR",
			})
			return
		}
		defer stmt.Close()
		for _, record := range records {
			_, err = stmt.Exec(record[0], record[1], record[2], record[3])
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: Restored %d whitelisted npubs sucessfully.", len(records)),
			Level:    "INFO",
		})

	} else if strings.Contains(backupFile, "-invoices-") {

		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: Preparing to restore %d invoices from '%s'", len(records), backupFile),
			Level:    "INFO",
		})

		stmt, err := storage.DB.Prepare("INSERT INTO invoice(pubkey,invoice,invoice_id,paid,amount_msat,expires_at,created_at) values(?,?,?,?,?,?,?) ON CONFLICT DO NOTHING")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer stmt.Close()
		for _, record := range records {
			_, err = stmt.Exec(record[0], record[1], record[2], record[3], record[4], record[5], record[6])
			if err != nil {
				utils.Logger(utils.LogEvent{
					Datetime: time.Now(),
					Content:  fmt.Sprintf("Backup: Failed to restore invoices:\n%v.", err),
					Level:    "ERROR",
				})
				return
			}
		}

		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: Restored %d invoices sucessfully.", len(records)),
			Level:    "INFO",
		})

	} else {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Backup: file '%s' doesnt match any supported backup types", backupFile),
			Level:    "ERROR",
		})
	}
}
