package main

import (
	"log"

	vaultapi "github.com/hashicorp/vault/api"
)

func newVaultClient(token, vaultURL string) (*vaultapi.Client, error) {
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

	return client, nil
}
