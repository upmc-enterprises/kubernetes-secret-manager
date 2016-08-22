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
	dataDir      = "/var/lib/vault-manager"
	vaultToken   = "41b108e8-dcac-3e6d-2980-d0ddae5e3696"
	vaultURL     = "http://127.0.0.1:8200"
	syncInterval = 120
)

func main() {
	flag.StringVar(&dataDir, "data-dir", dataDir, "Data directory path.")
	flag.StringVar(&vaultToken, "vault-token", vaultToken, "Token to access vault.")
	flag.StringVar(&vaultURL, "vault-url", vaultURL, "URL to access vault.")
	flag.IntVar(&syncInterval, "sync-interval", syncInterval, "Sync interval in seconds.")
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
		_, err = tx.CreateBucketIfNotExists([]byte("Accounts"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	client, _ := newVaultClient(vaultToken, vaultURL)
	c := client.Logical()

	// Write test secret
	_, err = c.Write("secret/foo",
		map[string]interface{}{
			"name":  "steve",
			"level": "kubernaut",
		})

	if err != nil {
		log.Println("ERROR writing secret: ", err)
	} else {
		log.Println("WROTE Secret!")
	}

	// Read sample secret
	readSecret, err := c.Read("secret/foo")

	if err != nil {
		log.Println("ERROR getting secret: ", err)
	} else {
		log.Println("Got secret: ", readSecret)
	}

	log.Println("Kubernetes Vault Controller started successfully.")

	// Process all Certificates definitions during the startup process.
	// err = syncCertificates(db)
	// if err != nil {
	// 	log.Println(err)
	// }

	doneChan := make(chan struct{})
	var wg sync.WaitGroup

	// Watch for events that add, modify, or delete Certificate definitions and
	// process them asynchronously.
	// log.Println("Watching for certificate events.")
	// wg.Add(1)
	// watchCertificateEvents(db, doneChan, &wg)
	//
	// // Start the certificate reconciler that will ensure all Certificate
	// // definitions are backed by a LetsEncrypt certificate and a Kubernetes
	// // TLS secret.
	// log.Println("Starting reconciliation loop.")
	// wg.Add(1)
	// reconcileCertificates(syncInterval, db, doneChan, &wg)

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
