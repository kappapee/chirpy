package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerValidation(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	msg := checkProfanity(params.Body)
	respondWithJSON(w, http.StatusOK, returnVals{
		CleanedBody: msg,
	})
}

func checkProfanity(msg string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(msg, " ")
	cleanedWords := make([]string, len(words))
	for i, w := range words {
		for _, p := range profaneWords {
			if strings.ToLower(w) == p {
				cleanedWords[i] = "****"
				break
			} else {
				cleanedWords[i] = w
			}
		}
	}
	return strings.Join(cleanedWords, " ")
}
