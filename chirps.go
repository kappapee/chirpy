package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"internal/auth"
	"internal/database"
	"log"
	"net/http"
	"strings"
	"time"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get token", nil)
		log.Printf("Couldn't get token: %v", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secretToken)
	if err != nil {
		log.Printf("User not authorized: %v", err)
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", nil)
		log.Printf("Couldn't decode parameters: %v", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleaned := checkProfanity(params.Body)

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: userID,
	})
	if err != nil {
		log.Printf("Could not create chirp in DB: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", nil)
		return
	}
	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		log.Printf("Could not retrieve chirps from DB: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", nil)
		return
	}
	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Could not parse chirp's ID: %s", r.PathValue("chirpID"))
		respondWithError(w, http.StatusInternalServerError, "Error processing chirp's ID", nil)
		return
	}
	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Printf("Could not retrieve chirp from DB: %v", err)
		respondWithError(w, http.StatusNotFound, "Couldn't retrieve chirp", nil)
		return
	}
	respondWithJSON(w, http.StatusOK,
		Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			UserID:    chirp.UserID,
			Body:      chirp.Body,
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
