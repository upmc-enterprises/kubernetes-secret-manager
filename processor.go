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
	for _, secret := range customSecrets {
		wg.Add(1)
		go func(secret CustomSecret) {
			defer wg.Done()
			err := processCustomSecret(secret, db)
			if err != nil {
				log.Println(err)
			}
		}(secret)
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
	deleteSecretLocal(c.Spec.Secret, db)
	log.Printf("Deleting Kubernetes CustomSecret secret: %s", c.Spec.Secret)
	return deleteKubernetesSecret(c.Spec.Secret)
}

func processCustomSecret(c CustomSecret, db *bolt.DB) error {

	//See if existing already
	foundSecret, _ := getSecretLocal(c.Spec.Secret, db)

	if foundSecret != nil {
		log.Println("Cert found existing! TTL: ", "")
		return nil
	}

	// Request credentials from user
	secret, err := vltClient.readVaultSecret(c.Spec.Policy)

	if err != nil {
		return errors.New("[Processor] Error getting secret from Vault: " + err.Error())
	}

	// Pull out user/password
	username, _ := secret.Data["username"].(string)
	password, _ := secret.Data["password"].(string)
	c.Spec.LeaseDuration = secret.LeaseDuration
	c.Spec.LeaseID = secret.LeaseID

	err = syncKubernetesSecret(c.Spec.Secret, username, password)

	if err != nil {
		return errors.New("[Processor] Error creating Kubernetes secret: " + err.Error())
	}

	// Persist
	persistSecretLocal(c.Spec.Secret, c.Spec, db)

	return nil
}
