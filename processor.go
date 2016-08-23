package main

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

// processorLock ensures that reconciliation and event processing does
// not happen at the same time.
var processorLock = &sync.Mutex{}

func reconcileCustomSecrets(interval int, db *bolt.DB, done chan struct{}, wg *sync.WaitGroup) {
	go func() {
		for {
			select {
			case <-time.After(time.Duration(interval) * time.Second):
				err := syncCustomSecrets(db)
				if err != nil {
					log.Println(err)
				}
			case <-done:
				wg.Done()
				log.Println("Stopped reconciliation loop.")
				return
			}
		}
	}()
}

func watchCustomSecretsEvents(db *bolt.DB, done chan struct{}, wg *sync.WaitGroup) {
	events, watchErrs := monitorCustomSecretsEvents()
	go func() {
		for {
			select {
			case event := <-events:
				err := processCustomSecretEvent(event, db)
				if err != nil {
					log.Println(err)
				}
			case err := <-watchErrs:
				log.Println(err)
			case <-done:
				wg.Done()
				log.Println("Stopped custom secrets event watcher.")
				return
			}
		}
	}()
}

func syncCustomSecrets(db *bolt.DB) error {
	processorLock.Lock()
	defer processorLock.Unlock()

	customSecrets, err := getCustomSecrets()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, cert := range customSecrets {
		wg.Add(1)
		go func(cert CustomSecret) {
			defer wg.Done()
			err := processCustomSecret(cert, db)
			if err != nil {
				log.Println(err)
			}
		}(cert)
	}
	wg.Wait()
	return nil
}

func processCustomSecretEvent(c CustomSecretEvent, db *bolt.DB) error {
	processorLock.Lock()
	defer processorLock.Unlock()
	switch {
	case c.Type == "ADDED":
		return processCustomSecret(c.Object, db)
	case c.Type == "DELETED":
		return deleteCustomSecret(c.Object, db)
	}
	return nil
}

func deleteCustomSecret(c CustomSecret, db *bolt.DB) error {
	log.Printf("Deleting Kubernetes CustomSecret secret: %s", c.Metadata["name"])
	return deleteKubernetesSecret(c.Metadata["name"])
}

func processCustomSecret(c CustomSecret, db *bolt.DB) error {

	// accessor token is unique & safe reference to actual token
	// specific to each db user / password combo

	// 1. lookup the accessor token in local db
	// 2. check the ttl (if close to expiring then renew / revoke)

	// FAKE!!
	err := syncKubernetesSecret(c.Metadata["name"], "mycooluser", "mycoolerPa$$w0rd")

	if err != nil {
		return errors.New("Error creating Kubernetes secret: " + err.Error())
	}

	return nil
}
