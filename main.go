package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/boltdb/bolt"
)

var (
	dataDir          = "/var/lib/vault-manager"
	vaultToken       = "98fe656b-2eb2-d4ec-84dc-50d872953f0d"
	vaultURL         = "http://127.0.0.1:8200"
	syncIntervalSecs = 5
	vltClient        *vaultClient
)

func main() {
	flag.StringVar(&dataDir, "data-dir", dataDir, "Data directory path.")
	flag.StringVar(&vaultToken, "vault-token", vaultToken, "Token to access vault.")
	flag.StringVar(&vaultURL, "vault-url", vaultURL, "URL to access vault.")
	flag.IntVar(&syncIntervalSecs, "sync-interval", syncIntervalSecs, "Sync interval in seconds.")
	flag.Parse()

	log.Println("Starting Kubernetes Vault Controller...")

	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:6060", nil))
	}()

	db, err := bolt.Open(path.Join(dataDir, "data.db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("Secrets"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Init vault client
	vltClient, err = newVaultClient(vaultToken, vaultURL)

	if err != nil {
		log.Println("Could not create Vault Client! ", err)
	}

	log.Println("Kubernetes Vault Controller started successfully.")

	// Process all Certificates definitions during the startup process.
	err = syncCustomSecrets(db)
	if err != nil {
		log.Println(err)
	}

	doneChan := make(chan struct{})
	var wg sync.WaitGroup

	// Watch for events that add, modify, or delete CustomSecret definitions and
	// process them asynchronously.
	log.Println("Watching for custom secret events.")
	wg.Add(1)
	watchCustomSecretsEvents(db, doneChan, &wg)

	// Start the custom secret reconciler that will ensure all Custom Secret
	// definitions are implemented with a Vault secret and a Kubernetes secret.
	log.Println("Starting reconciliation loop.")
	wg.Add(1)
	reconcileCustomSecrets(syncIntervalSecs, db, doneChan, &wg)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-signalChan:
			log.Printf("Shutdown signal received, exiting...")
			close(doneChan)
			wg.Wait()
			os.Exit(0)
		}
	}
}
