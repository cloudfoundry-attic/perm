package oidcx

import (
	"net/http"

	"fmt"

	"encoding/json"
)

const (
	OpenIDConfigurationEndpoint = "/.well-known/openid-configuration"
)

type OIDCProviderConfiguration struct {
	Issuer string `json:"issuer"`
}

func GetOIDCIssuer(client *http.Client, oidcProviderURL string) (string, error) {
	fullURL := fmt.Sprintf("%s%s", oidcProviderURL, OpenIDConfigurationEndpoint)

	res, err := client.Get(fullURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("HTTP bad response: %s", res.Status)
		return "", err
	}

	var configuration OIDCProviderConfiguration
	err = json.NewDecoder(res.Body).Decode(&configuration)
	if err != nil {
		return "", err
	}

	return configuration.Issuer, nil
}
