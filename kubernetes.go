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
	"time"
)

// TODO: Handle multiple namespaces!

var (
	apiHost                    = "http://127.0.0.1:8001"
	customSecretsEndpoint      = "/apis/enterprises.upmc.com/v1/namespaces/default/customsecretses"
	customSecretsWatchEndpoint = "/apis/enterprises.upmc.com/v1/namespaces/default/customsecretses?watch=true"
	secretsEndpoint            = "/api/v1/namespaces/default/secrets"
)

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
	Policy string `json:"policy"`
	Secret string `json:"secret"`
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

func syncKubernetesSecret(secretName, username, password string) error {
	metadata := make(map[string]string)
	metadata["name"] = secretName

	data := make(map[string]string)
	data["username"] = base64.StdEncoding.EncodeToString([]byte(username))
	data["password"] = base64.StdEncoding.EncodeToString([]byte(password))
	//data["ttlExpire"] = time.Now().Add(time.Minute * 2).String()

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

		if currentSecret.Data["username"] != secret.Data["username"] || currentSecret.Data["password"] != secret.Data["password"] {
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
