package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform-exec/tfexec"
	"log"
	"net/http"
	"os"
	"time"
)
import _ "github.com/joho/godotenv/autoload"
import vault "github.com/hashicorp/vault/api"

var vaultClient *vault.Client
var tf *tfexec.Terraform

func main() {
	initVault()
	initTerraform()

	r := mux.NewRouter()
	r.HandleFunc("/initialize/{organization}", initializeOrganization).Methods("POST")
	r.HandleFunc("/configure/{organization}", configureAWS).Methods("POST")
	r.HandleFunc("/generate-credentials/{organization}", generateCredentials).Methods("POST")
	r.HandleFunc("/run-tf/{organization}", runTF).Methods("POST")

	log.Println("vault client running on port ", os.Getenv("PORT"))
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

	log.Println("connected to vault")
}

func initTerraform() {
	workingDir, _ := os.Getwd()
	//TODO install terraform in docker
	terraform, err := tfexec.NewTerraform(workingDir, "/usr/local/bin/terraform")

	tf = terraform

	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	log.Println("terraform initialized")

}

func getOrganizationPath(id string) string {
	return id + "_aws"
}

func generateCredentials(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	secret, err := vaultClient.Logical().Read(getOrganizationPath(vars["organization"]) + "/creds/backend-role")

	if err != nil {
		log.Println(err)
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
		log.Println(err)
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
					  "Effect": "Allow",
					  "Action": [
						"iam:*", "ec2:*"
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
		log.Println(err)
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
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Organization mounted")
}

func runTF(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	err := tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	secret, err := vaultClient.Logical().Read(getOrganizationPath(vars["organization"]) + "/creds/backend-role")

	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	err = os.Setenv("TF_VAR_aws_access_key", secret.Data["access_key"].(string))
	err = os.Setenv("TF_VAR_aws_secret_key", secret.Data["secret_key"].(string))

	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	//wait for credentials to be active
	time.Sleep(5 * time.Second)

	err = tf.Apply(context.Background())

	if err != nil {
		log.Fatalf("error running apply %s", err)
	}

	fmt.Fprintf(w, "terraform applied")
}
