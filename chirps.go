package main

import (
	"encoding/json"
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
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get token", nil)
		log.Printf("Couldn't get token: %s", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.platform)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization error HERE", nil)
		log.Printf("User not authorized HERE: %s", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	// TODO: check why is this thing throwing a W/E if I use err here as well?
	error := decoder.Decode(&params)
	if error != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", nil)
		log.Printf("Couldn't decode parameters: %s", error)
		return
	}

	if userID.String() != params.UserID.String() {
		respondWithError(w, http.StatusUnauthorized, "Authorization error", nil)
		log.Printf("User not authorized: %s", err)
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
		UserID: params.UserID,
	})
	if err != nil {
		log.Printf("Could not create chirp in DB: %s", err)
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
		log.Printf("Could not retrieve chirps from DB: %s", err)
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
		log.Printf("Could not retrieve chirp from DB: %s", err)
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
