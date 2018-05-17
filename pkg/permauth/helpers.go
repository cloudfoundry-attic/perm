package permauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	UAATokenKeysEndpoint = "token_keys"
	UAAIssuerEndpoint    = ".well-known/openid-configuration"
)

func GetUAAPubKey(UAAURL string) (string, error) {
	contents, err := getPageContents(UAAURL, UAATokenKeysEndpoint)
	if err != nil {
		return "", err
	}

	var tokenKeys struct {
		Keys []struct {
			Value string `json:"value"`
		} `json:"keys"`
	}

	err = json.Unmarshal(contents, &tokenKeys)
	if err != nil {
		return "", err
	}

	if len(tokenKeys.Keys) == 0 {
		return "", fmt.Errorf("No public key found on the UAA /%s endpoint", UAATokenKeysEndpoint)
	}

	return tokenKeys.Keys[0].Value, nil
}

func GetUAAIssuer(UAAURL string) (string, error) {
	contents, err := getPageContents(UAAURL, UAAIssuerEndpoint)
	if err != nil {
		return "", err
	}

	var uaaIssuer struct {
		Issuer string `json:"issuer"`
	}

	err = json.Unmarshal(contents, &uaaIssuer)
	if err != nil {
		return "", err
	}

	if uaaIssuer.Issuer == "" {
		return "", fmt.Errorf("No issuer found on the UAA /%s endpoint", UAAIssuerEndpoint)
	}

	return uaaIssuer.Issuer, nil
}

func getPageContents(url string, path string) ([]byte, error) {
	response, err := http.Get(fmt.Sprintf("%s/%s", url, path))
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return contents, nil
}
