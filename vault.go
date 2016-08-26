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
		log.Println("[Vault] Error getting secret: ", err)
		return nil, err
	}

	return readSecret, nil
}

func (vc *vaultClient) writeVaultSecret(key string, data map[string]interface{}) error {

	c := vc.client.Logical()
	_, err := c.Write(key, data)

	if err != nil {
		log.Println("[Vault] Error writing secret: ", err)
		return err
	}

	return nil
}

func (vc *vaultClient) revokeVaultSecret(leaseID string) error {
	err := vc.client.Sys().Revoke(leaseID)

	if err != nil {
		log.Println("[Vault] Error revoking secret: ", err)
		return err
	}

	return nil
}

func (vc *vaultClient) renewVaultLease(leaseID string, leaseDuration int) (*vaultapi.Secret, error) {
	secret, err := vc.client.Sys().Renew(leaseID, leaseDuration)

	if err != nil {
		log.Println("[Vault] Error renewing secret: ", err)
		return nil, err
	}

	return secret, nil
}
