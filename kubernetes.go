package main

var (
	apiHost                   = "http://127.0.0.1:8001"
	certificatesEndpoint      = "/apis/enterprises.upmc.com/v1/namespaces/default/customsecrets"
	certificatesWatchEndpoint = "/apis/enterprises.upmc.com/v1/namespaces/default/customsecrets?watch=true"
	secretsEndpoint           = "/api/v1/namespaces/default/secrets"
)

type Secret struct {
	Kind       string            `json:"kind"`
	ApiVersion string            `json:"apiVersion"`
	Metadata   map[string]string `json:"metadata"`
	Data       map[string]string `json:"data"`
	Type       string            `json:"type"`
}
