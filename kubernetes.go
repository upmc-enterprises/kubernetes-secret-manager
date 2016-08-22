package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

var (
	apiHost                    = "http://127.0.0.1:8001"
	customSecretsEndpoint      = "/apis/enterprises.upmc.com/v1/namespaces/default/customsecrets"
	customSecretsWatchEndpoint = "/apis/enterprises.upmc.com/v1/namespaces/default/customsecrets?watch=true"
	secretsEndpoint            = "/api/v1/namespaces/default/secrets"
)

type CustomSecretEvent struct {
	Type   string       `json:"type"`
	Object CustomSecret `json:"object"`
}

type CustomSecret struct {
	ApiVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   map[string]string `json:"metadata"`
	Spec       CustomSecretSpec  `json:"spec"`
}

type CustomSecretSpec struct {
	Policy string `json:"policy"`
}

type CustomSecretList struct {
	ApiVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   map[string]string `json:"metadata"`
	Items      []CustomSecret    `json:"items"`
}

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
