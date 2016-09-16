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

/*
   Changes
   2016-09-12: Lachlan Evenson - Add namespace support
   2016-09-12: Lachlan Evenson - Add TPR creation PR 11
   2016-09-16: Steve Sloka - Enable custom secrets PR 4
*/

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	apiHost = "http://127.0.0.1:8001"
	// Add namespace support - namespace variable provided by Kubernetes downwards API.
	namespace                  = os.Getenv("NAMESPACE")
	customSecretsEndpoint      = fmt.Sprintf("/apis/enterprises.upmc.com/v1/namespaces/%s/customsecretses", namespace)
	customSecretsWatchEndpoint = fmt.Sprintf("/apis/enterprises.upmc.com/v1/namespaces/%s/customsecretses?watch=true", namespace)
	secretsEndpoint            = fmt.Sprintf("/api/v1/namespaces/%s/secrets", namespace)
	tprEndpoint                = "/apis/extensions/v1beta1/thirdpartyresources"
)

//ThirdPartResource in Kubernetes
type ThirdPartyResource struct {
	ApiVersion  string               `json:"apiVersion"`
	Kind        string               `json:"kind"`
	Description string               `json:"description"`
	Metadata    map[string]string    `json:"metadata"`
	Versions    [1]map[string]string `json:"versions,omitempty"`
}

// CustomSecretEvent stores when a secret needs created
type CustomSecretEvent struct {
	Type   string       `json:"type"`
	Object CustomSecret `json:"object"`
}

// CustomSecret represents a custom secret object
type CustomSecret struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   map[string]string `json:"metadata"`
	Spec       CustomSecretSpec  `json:"spec"`
}

// CustomSecretSpec represents the custom data of the object
type CustomSecretSpec struct {
	Policy              string    `json:"policy"`
	Secret              string    `json:"secret"`
	LeaseDuration       int       `json:"leaseDuration"`
	LeaseID             string    `json:"leastId"`
	LeaseExpirationDate time.Time `json:"leaseExpirationDate"`
}

// CustomSecretList represents a list of CustomSecrets
type CustomSecretList struct {
	ApiVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   map[string]string `json:"metadata"`
	Items      []CustomSecret    `json:"items"`
}

// Secret represents a Kubernetes secret type
type Secret struct {
	Kind       string            `json:"kind"`
	ApiVersion string            `json:"apiVersion"`
	Metadata   map[string]string `json:"metadata"`
	Data       map[string]string `json:"data"`
	Type       string            `json:"type"`
}

func getCustomSecrets() ([]CustomSecret, error) {
	var resp *http.Response
	var err error
	for {
		resp, err = http.Get(apiHost + customSecretsEndpoint)
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	var customSecretList CustomSecretList
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&customSecretList)
	if err != nil {
		return nil, err
	}

	return customSecretList.Items, nil
}

func monitorCustomSecretsEvents() (<-chan CustomSecretEvent, <-chan error) {
	events := make(chan CustomSecretEvent)
	errc := make(chan error, 1)
	go func() {
		for {
			resp, err := http.Get(apiHost + customSecretsWatchEndpoint)
			if err != nil {
				errc <- err
				time.Sleep(5 * time.Second)
				continue
			}
			if resp.StatusCode != 200 {
				errc <- errors.New("Invalid status code: " + resp.Status)
				time.Sleep(5 * time.Second)
				continue
			}

			decoder := json.NewDecoder(resp.Body)
			for {
				var event CustomSecretEvent
				err = decoder.Decode(&event)
				if err != nil {
					errc <- err
					break
				}
				events <- event
			}
		}
	}()

	return events, errc
}

func checkSecret(name string) (bool, error) {
	resp, err := http.Get(apiHost + secretsEndpoint + "/" + name)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, nil
	}
	return true, nil
}

func deleteKubernetesSecret(domain string) error {
	req, err := http.NewRequest("DELETE", apiHost+secretsEndpoint+"/"+domain, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Deleting %s secret failed: %s", domain, resp.Status)
	}
	return nil
}

func syncKubernetesSecret(secretName string, secretData map[string]interface{}) error {
	metadata := make(map[string]string)
	metadata["name"] = secretName

	// Map the Vault Secret Map to a String Map
	// NOTE: `secretData` is from VaultSecret struct
	data := make(map[string]string)
	for k, v := range secretData {
		data[k] = base64.StdEncoding.EncodeToString([]byte(v.(string)))
	}

	secret := &Secret{
		ApiVersion: "v1",
		Data:       data,
		Kind:       "Secret",
		Metadata:   metadata,
		Type:       "Opaque",
	}

	resp, err := http.Get(apiHost + secretsEndpoint + "/" + secretName)
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		// compare current cert
		var currentSecret Secret
		d, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		err = json.Unmarshal(d, &currentSecret)
		if err != nil {
			return err
		}

		if currentSecret.Data["username"] != secret.Data["username"] ||
			currentSecret.Data["password"] != secret.Data["password"] {

			log.Printf("%s secret out of sync.", secretName)

			currentSecret.Data = secret.Data
			var b []byte
			body := bytes.NewBuffer(b)
			err := json.NewEncoder(body).Encode(currentSecret)
			if err != nil {
				return err
			}
			req, err := http.NewRequest("PUT", apiHost+secretsEndpoint+"/"+secretName, body)
			if err != nil {
				return err
			}
			req.Header.Add("Content-Type", "application/json")
			respSecret, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			if resp.StatusCode != 200 {
				return errors.New("Updating secret failed:" + respSecret.Status)
			}
			log.Printf("Syncing %s secret complete.", secretName)
		}
		return nil
	}

	if resp.StatusCode == 404 {
		log.Printf("%s secret missing.", secretName)
		var b []byte
		body := bytes.NewBuffer(b)
		err := json.NewEncoder(body).Encode(secret)
		if err != nil {
			return err
		}

		resp, err := http.Post(apiHost+secretsEndpoint, "application/json", body)
		if err != nil {
			return err
		}
		if resp.StatusCode != 201 {
			return errors.New("Secrets: Unexpected HTTP status code" + resp.Status)
		}
		log.Printf("%s secret created.", secretName)
		return nil
	}
	return nil
}

// Check if CustomSecrets TPR exists. If not, create
func createKubernetesThirdPartyResource(tpr_name string, tpr_desc string, tpr_version string) error {
	metadata := make(map[string]string)
	metadata["name"] = tpr_name

	data := [1]map[string]string{}
	aom1 := map[string]string{"name": tpr_version}
	data[0] = aom1

	tpr := &ThirdPartyResource{
		ApiVersion:  "extensions/v1beta1",
		Kind:        "ThirdPartyResource",
		Description: tpr_desc,
		Metadata:    metadata,
		Versions:    data,
	}

	resp, _ := http.Get(apiHost + customSecretsEndpoint)

	if resp.StatusCode == 200 {
		// ThirdPartyResource already exists. Move on
		log.Printf("ThirdPartyResource %s exists.", tpr_name)
		return nil
	}

	if resp.StatusCode == 404 {
		log.Printf("creating ThirdPartyResource %s", tpr_name)
		var b []byte
		body := bytes.NewBuffer(b)
		err := json.NewEncoder(body).Encode(tpr)
		if err != nil {
			return err
		}

		resp, err := http.Post(apiHost+tprEndpoint, "application/json", body)
		if err != nil {
			return err
		}
		if resp.StatusCode != 201 {
			return errors.New("ThirdPartyResource: Unexpected HTTP status code" + resp.Status)
		}
		log.Printf("ThirdPartyResource %s created.", tpr_name)
		return nil
	}
	return nil
}
