package workers

import (
	"context"
	"encoding/json"
	"github.com/nbd-wtf/go-nostr"
	"github.com/orcaman/concurrent-map/v2"
	"github.com/spf13/viper"
	"log"
	"nostr-citadel/storage"
	"sync"
	"time"
)

var eventIndex = cmap.New[*nostr.Event]()

type PlebRelay struct {
	Write bool `json:"write"`
	Read  bool `json:"read"`
}

type RelaySettings struct {
	URL        string
	LastSynced time.Time
}

var (
	numberOfWorkers    = 3
	numberOfSubWorkers = 5
)

// StartImporter Todo: Redo logic and structure.
func StartImporter() {

	numberOfWorkers = viper.GetInt("importer.workers")
	numberOfSubWorkers = viper.GetInt("importer.fetchers")

	go func() {
		for {
			log.Println("Worker: Running Import Jobs")
			plebs, err := storage.GetPlebsToSync()

			if err != nil {
				log.Printf("Error getting plebs: %v", err)
				continue
			}

			var wg sync.WaitGroup

			wg.Add(numberOfWorkers)
			tasks := make(chan storage.Pleb)

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

			log.Printf("Worker: Found a total of %d events\n", len(eventIndex.Items()))
			count := 0
			for _, event := range eventIndex.Items() {
				ok, _ := storage.WriteEvent(event)
				if ok {
					count++
				}

			}
			log.Printf("Worker: Imported %d new events", count)
			time.Sleep(1 * time.Hour)
		}
	}()
}

func worker(tasks <-chan storage.Pleb, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		task, ok := <-tasks
		if !ok {
			return
		}
		var wgg sync.WaitGroup

		wgg.Add(numberOfSubWorkers)
		log.Println("Worker: Started import for PubKey", task.PubKey[:8])
		subTasks := make(chan RelaySettings)

		for i := 0; i < numberOfSubWorkers; i++ {
			go subWorker(subTasks, &wgg, task.PubKey)
		}

		var relays map[string]PlebRelay
		if err := json.Unmarshal([]byte(task.Content), &relays); err != nil {
			panic(err)
		}

		log.Printf("Worker: Starting import from %d relays for %s \n", len(relays), task.PubKey[:8])

		i := 0
		relay := make([]string, len(relays))
		for k := range relays {
			relay[i] = k
			i++
		}

		storage.SetLastSync(task.PubKey)
		for i := 0; i < len(relay); i++ {
			relay := RelaySettings{URL: relay[i], LastSynced: time.Unix(task.LastSynced, 0)}
			subTask := relay
			subTasks <- subTask
		}
		close(subTasks)
		wgg.Wait()
		log.Println("Worker: Finished fetching events for", task.PubKey[:8])
	}
}

func subWorker(subtasks chan RelaySettings, wg *sync.WaitGroup, pubKey string) {
	defer wg.Done()
	for {
		subTask, ok := <-subtasks
		if !ok {
			return
		}
		importFromRelay(pubKey, subTask)
	}
}

func importFromRelay(pubKey string, relayAddress RelaySettings) {

	if relayAddress.URL == viper.GetString("relay_config.relay_url") {
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
