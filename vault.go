package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)


type SecretResponse struct {
	Data struct{
		Data map[string]string
	} `json:"data"`
}

func getSecret(key string) (map[string]string, error)  {
	client := http.Client{}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/secret/data/%s", os.Getenv("VAULT_API_ADDR"), key), nil)

	if err != nil {
		return nil, err
	}

	request.Header.Set("X-Vault-Token", os.Getenv("VAULT_TOKEN"))

	response, err := client.Do(request)

	if err != nil {
		return nil, err
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var body SecretResponse
	err = json.Unmarshal(responseBody, &body)
	if err != nil {
		return nil, err
	}

	return body.Data.Data, nil
}

func createSecret(key string, values map[string]string) error  {
	client := http.Client{}

	body, err := json.Marshal(map[string]interface{}{
		"data": values,
	})
	if err != nil {
		return err
	}

	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/v1/secret/data/%s", os.Getenv("VAULT_API_ADDR"), key), bytes.NewBuffer(body))

	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Vault-Token", os.Getenv("VAULT_TOKEN"))

	response, err := client.Do(request)

	if err != nil {
		return err
	}

	if response.StatusCode == 200{
		return nil
	} else {
		b, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(b))
		return errors.New("UNABLE TO CREATE SECRET")
	}
}
