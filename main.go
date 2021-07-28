package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)
import _ "github.com/joho/godotenv/autoload"
import vault "github.com/hashicorp/vault/api"

var vaultClient *vault.Client

func main() {
	initVault()

	r := mux.NewRouter()
	r.HandleFunc("/initialize/{organization}", initializeOrganization).Methods("POST")
	r.HandleFunc("/configure/{organization}", configureAWS).Methods("POST")
	r.HandleFunc("/generate-credentials/{organization}", generateCredentials).Methods("POST")

	fmt.Println("vault client running on port ", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), r))
}

func initVault() {
	conf := &vault.Config{
		Address: os.Getenv("VAULT_API_ADDR"),
	}

	client, err := vault.NewClient(conf)
	if err != nil {
		panic(err)
	}
	vaultClient = client
	vaultClient.SetToken(os.Getenv("VAULT_TOKEN"))

	fmt.Println("connected to vault")
}

func getOrganizationPath(id string) string {
	return id + "_aws"
}

func generateCredentials(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	secret, err := vaultClient.Logical().Read(getOrganizationPath(vars["organization"]) + "/creds/backend-role")

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(secret.Data)
}

func configureAWS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var body map[string]string

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid Request")
		return
	}

	_, err = vaultClient.Logical().Write(getOrganizationPath(vars["organization"])+"/config/root", map[string]interface{}{
		"access_key": body["aws_secret_key"],
		"secret_key": body["secret_key"],
		"region":     body["region"],
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}

	//create backend role
	_, err = vaultClient.Logical().Write(getOrganizationPath(vars["organization"])+"/roles/backend-role", map[string]interface{}{
		"credential_type": "iam_user",
		"policy_document": `
				{
				  "Version": "2012-10-17",
				  "Statement": [
					{
					  "Sid": "Stmt1426528957000",
					  "Effect": "Allow",
					  "Action": [
						"ec2:*"
					  ],
					  "Resource": [
						"*"
					  ]
					}
				  ]
				}
			`,
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "AWS configured")
}

func initializeOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// mounts organization at /{organization}_aws
	err := vaultClient.Sys().Mount(getOrganizationPath(vars["organization"]), &vault.MountInput{
		Type: "aws",
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Organization mounted")
}
