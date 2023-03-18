package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"github.com/orcaman/concurrent-map/v2"
	"nostr-citadel/pkg/config"
	"nostr-citadel/pkg/models"
	"nostr-citadel/pkg/utils"
	"strings"
	"sync"
	"time"
)

var eventIndex = cmap.New[*nostr.Event]()

type RelaySettings struct {
	URL        string
	LastSynced time.Time
}

var (
	numberOfWorkers    = 3
	numberOfSubWorkers = 5
)

func ImportData(npub string) {

	utils.Logger(utils.LogEvent{
		Datetime: time.Now(),
		Content:  "Worker: Running Import Jobs",
		Level:    "INFO",
	})

	plebs, err := models.GetPlebsToSync()

	if err != nil {
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Error getting pleb to sync %s - %s", npub, err.Error()),
			Level:    "ERROR",
		})
		return
	}
	var wg sync.WaitGroup

	wg.Add(numberOfWorkers)
	tasks := make(chan models.Pleb)

	// Start up workers
	for i := 0; i < numberOfWorkers; i++ {
		go worker(tasks, &wg)
	}

	// Send work to workers
	for i := 0; i < len(plebs); i++ {
		tasks <- plebs[i]
	}

	close(tasks)
	wg.Wait()

	utils.Logger(utils.LogEvent{
		Datetime: time.Now(),
		Content:  fmt.Sprintf("Worker: Found a total of %d events", len(eventIndex.Items())),
		Level:    "INFO",
	})

	count := 0
	for _, event := range eventIndex.Items() {
		ok, msg := models.WriteEvent(event)
		if ok {
			count++
		} else {
			utils.Logger(utils.LogEvent{
				Datetime: time.Now(),
				Content:  fmt.Sprintf("Import failure - %s", msg),
				Level:    "ERROR",
			})
		}
	}
	utils.Logger(utils.LogEvent{
		Datetime: time.Now(),
		Content:  fmt.Sprintf("Worker: Imported %d new events", count),
		Level:    "INFO",
	})
}

// StartImporter Todo: Redo logic and structure.
func StartImporter() {

	numberOfWorkers = config.Config.Importer.Workers
	numberOfSubWorkers = config.Config.Importer.Fetchers

	go func() {
		for {
			// Let everything initialize before start
			time.Sleep(5 * time.Second)

			utils.Logger(utils.LogEvent{
				Datetime: time.Now(),
				Content:  "Worker: Running Import Jobs",
				Level:    "INFO",
			})
			plebs, err := models.GetPlebsToSync()

			if err != nil {
				utils.Logger(utils.LogEvent{
					Datetime: time.Now(),
					Content:  fmt.Sprintf("Error getting plebs: %v", err),
					Level:    "ERROR",
				})
				continue
			}

			var wg sync.WaitGroup

			wg.Add(numberOfWorkers)
			tasks := make(chan models.Pleb)

			// Start up workers
			for i := 0; i < numberOfWorkers; i++ {
				go worker(tasks, &wg)
			}

			// Send work to workers
			for i := 0; i < len(plebs); i++ {
				tasks <- plebs[i]
			}

			close(tasks)
			wg.Wait()

			utils.Logger(utils.LogEvent{
				Datetime: time.Now(),
				Content:  fmt.Sprintf("Worker: Found a total of %d events", len(eventIndex.Items())),
				Level:    "INFO",
			})

			count := 0
			for _, event := range eventIndex.Items() {
				ok, _ := models.WriteEvent(event)
				if ok {
					count++
				}

			}

			utils.Logger(utils.LogEvent{
				Datetime: time.Now(),
				Content:  fmt.Sprintf("Worker: Imported %d new events", count),
				Level:    "INFO",
			})

			time.Sleep(1 * time.Hour)
		}
	}()
}

func worker(plebs <-chan models.Pleb, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		pleb, ok := <-plebs
		if !ok {
			return
		}
		var wgg sync.WaitGroup

		wgg.Add(numberOfSubWorkers)
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Worker: Started import for PubKey %s", pleb.PubKey[:8]),
			Level:    "INFO",
		})

		subTasks := make(chan RelaySettings)

		for i := 0; i < numberOfSubWorkers; i++ {
			go subWorker(subTasks, &wgg, pleb.PubKey)
		}

		var indexRelays map[string]Relay
		if len(pleb.Relays) >= 1 {
			if err := json.Unmarshal([]byte(pleb.Relays), &indexRelays); err != nil {
				panic(err)
			}
		} else {
			indexRelays = DefaultRelays
		}

		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Worker: Starting import from %d relays for %s", len(indexRelays), pleb.PubKey[:8]),
			Level:    "INFO",
		})

		i := 0
		relay := make([]string, len(indexRelays))
		for k := range indexRelays {
			relay[i] = k
			i++
		}

		models.SetLastSync(pleb.PubKey)
		for i := 0; i < len(relay); i++ {
			relaySettings := RelaySettings{URL: relay[i], LastSynced: time.Unix(pleb.LastSynced, 0)}
			subTask := relaySettings
			subTasks <- subTask
		}
		close(subTasks)
		wgg.Wait()
		utils.Logger(utils.LogEvent{
			Datetime: time.Now(),
			Content:  fmt.Sprintf("Worker: Finished fetching events for %s", pleb.PubKey[:8]),
			Level:    "INFO",
		})
	}
}

func subWorker(relaySettings chan RelaySettings, wg *sync.WaitGroup, pubKey string) {
	defer wg.Done()
	for {
		subTask, ok := <-relaySettings
		if !ok {
			return
		}
		importFromRelay(pubKey, subTask)
	}
}

func importFromRelay(pubKey string, relayAddress RelaySettings) {

	var relayUrl string
	t := strings.Split(config.Config.Relay.RelayURL, "//")
	if len(t) > 1 {
		if t[0] == "https:" {
			relayUrl = "wss://" + t[1]
		} else {
			relayUrl = "ws://" + t[1]
		}
	}

	if relayAddress.URL == relayUrl {
		return
	}

	relay, err := nostr.RelayConnect(context.Background(), relayAddress.URL)
	if err != nil {
		// Can't connect to relay, skip it.
		return
	}

	var filters nostr.Filters
	filters = []nostr.Filter{{
		Authors: []string{pubKey},
		Since:   &relayAddress.LastSynced,
	}}

	ctx, cancel := context.WithCancel(context.Background())
	sub := relay.Subscribe(ctx, filters)

	go func() {
		select {
		case <-time.After(10 * time.Second):
			sub.Unsub()
			cancel()
			_ = relay.Close()
		case <-sub.EndOfStoredEvents:
			sub.Unsub()
			cancel()
			_ = relay.Close()
		}
	}()

	go func() {
		for ev := range sub.Events {
			eventIndex.Set(ev.ID, ev)
		}
	}()

}