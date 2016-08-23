package main

import (
	"bytes"
	"encoding/gob"

	"github.com/boltdb/bolt"
)

func getSecretLocal(name string, db *bolt.DB) (*CustomSecretSpec, error) {
	var secret *CustomSecretSpec
	err := db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket([]byte("Secrets")).Get([]byte(name))
		if data == nil {
			return nil
		}
		decoder := gob.NewDecoder(bytes.NewReader(data))
		err := decoder.Decode(&secret)
		if err != nil {
			return err
		}
		return nil
	})
	return secret, err
}

func persistSecretLocal(name string, customSecret CustomSecretSpec, db *bolt.DB) error {

	data := new(bytes.Buffer)
	enc := gob.NewEncoder(data)
	err := enc.Encode(customSecret)
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if err != nil {
			return err
		}
		bucket := tx.Bucket([]byte("Secrets"))
		err = bucket.Put([]byte(name), data.Bytes())
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func deleteSecretLocal(secretName string, db *bolt.DB) error {
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("Secrets")).Delete([]byte(secretName))
	})
	return err
}
