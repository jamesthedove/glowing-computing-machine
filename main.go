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

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/secrets/{organization}", addSecretHandler).Methods("PUT")
	r.HandleFunc("/secrets/{organization}", getSecretHandler).Methods("GET")
	fmt.Println("vault client running on port ", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(":"+ os.Getenv("PORT"), r))
}

// Create new secret
func addSecretHandler(w http.ResponseWriter, r *http.Request)  {
	vars := mux.Vars(r)

	var body map[string]string

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid Request")
		return
	}

	err = createSecret(vars["organization"], body)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Could not create secret")
		return
	}


	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Secrets created")
}

// Get secret
func getSecretHandler(w http.ResponseWriter, r *http.Request)  {
	vars := mux.Vars(r)

	secrets, err := getSecret(vars["organization"])

	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(secrets)
}
