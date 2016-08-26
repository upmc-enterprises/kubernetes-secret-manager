/*
Copyright (c) 2016, UPMC Enterprises
All rights reserved.
Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:
    * Redistributions of source code must retain the above copyright
      notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above copyright
      notice, this list of conditions and the following disclaimer in the
      documentation and/or other materials provided with the distribution.
    * Neither the name UPMC Enterprises nor the
      names of its contributors may be used to endorse or promote products
      derived from this software without specific prior written permission.
THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL UPMC ENTERPRISES BE LIABLE FOR ANY
DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
*/

package main

import (
	"errors"
	"log"
	"math"
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

		// Lookup the duration left on the lease, if expiring soon then renew
		ttlRemaining := foundSecret.LeaseExpirationDate.Sub(time.Now())

		// If the expiration date is in the past
		if ttlRemaining.Seconds() <= 0 {
			// Refresh creds
			deleteSecretLocal(c.Spec.Secret, db)
		} else if int(math.Abs(ttlRemaining.Seconds())) <= foundSecret.LeaseDuration/2 {
			// If ttl remaining is less than 1/2 of ttl lease, renew
			log.Println("Renewing lease for id: ", foundSecret.LeaseID)
			renewedSecret, err := vltClient.renewVaultLease(foundSecret.LeaseID, foundSecret.LeaseDuration)

			if err != nil {
				return errors.New("[Processor] Error renewing lease from Vault: " + err.Error())
			}

			// If secret is hitting max ttl, refresh with new secret from Vault
			if renewedSecret.LeaseDuration < foundSecret.LeaseDuration {
				deleteSecretLocal(c.Spec.Secret, db)
			} else {

				// Update DB
				c.Spec.LeaseID = renewedSecret.LeaseID
				c.Spec.LeaseDuration = renewedSecret.LeaseDuration
				c.Spec.LeaseExpirationDate = time.Now().Add(time.Second * time.Duration(renewedSecret.LeaseDuration))
				persistSecretLocal(c.Spec.Secret, c.Spec, db)

				return nil
			}
		} else {
			log.Printf("Lease (%s) is valid, skipping renewal! TTL remaining: %f",
				foundSecret.LeaseID, math.Abs(ttlRemaining.Seconds()))

			return nil
		}
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
	c.Spec.LeaseExpirationDate = time.Now().Add(time.Second * time.Duration(secret.LeaseDuration))

	err = syncKubernetesSecret(c.Spec.Secret, username, password)

	if err != nil {
		// Delete the Vault secret since we couldn't persist to k8s
		vltClient.revokeVaultSecret(secret.LeaseID)

		return errors.New("[Processor] Error creating Kubernetes secret: " + err.Error())
	}

	// Persist to DB
	persistSecretLocal(c.Spec.Secret, c.Spec, db)

	return nil
}
