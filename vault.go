package main

import (
	"log"

	vaultapi "github.com/hashicorp/vault/api"
)

type vaultClient struct {
	client *vaultapi.Client
}

func newVaultClient(token, vaultURL string) (*vaultClient, error) {
	config := &vaultapi.Config{
		Address: vaultURL,
	}

	client, err := vaultapi.NewClient(config)

	if err != nil {
		log.Println("ERROR creating Vault client! ", err)
		return nil, err
	}

	// Set token in Vault
	client.SetToken(vaultToken)

	return &vaultClient{client}, nil
}

func (vc *vaultClient) readVaultSecret(key string) (*vaultapi.Secret, error) {

	c := vc.client.Logical()

	// Read sample secret
	readSecret, err := c.Read(key)

	if err != nil {
		log.Println("ERROR getting secret: ", err)
		return nil, err
	}

	return readSecret, nil
}

func (vc *vaultClient) writeVaultSecret(key string, data map[string]interface{}) error {

	c := vc.client.Logical()
	_, err := c.Write(key, data)

	if err != nil {
		log.Println("ERROR writing secret: ", err)
		return err
	}

	return nil
}
