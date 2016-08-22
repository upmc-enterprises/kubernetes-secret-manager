package main

import (
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
	return nil
}

func processCustomSecret(c CustomSecret, db *bolt.DB) error {
	return nil
}
