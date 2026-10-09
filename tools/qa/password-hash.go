// Herramienta de QA de #300. El wrapper la ejecuta dentro del único módulo Go.
package main

import (
	"encoding/json"
	"os"

	"system-barbershop/internal/modules/auth"
)

func main() {
	var passwords []string
	if json.NewDecoder(os.Stdin).Decode(&passwords) != nil {
		os.Exit(1)
	}
	hashes := make([]string, len(passwords))
	hasher := auth.NewArgon2Hasher()
	for i, password := range passwords {
		h, err := hasher.Hash(password)
		if err != nil {
			os.Exit(1)
		}
		hashes[i] = h
	}
	if json.NewEncoder(os.Stdout).Encode(hashes) != nil {
		os.Exit(1)
	}
}
