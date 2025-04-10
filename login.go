package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"internal/auth"
	"log"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUser(r.Context(), params.Email)
	if err != nil {
		log.Printf("Could not get user %s from DB: %s", params.Email, err)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	if auth.CheckPasswordHash(user.HashedPassword, params.Password) != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	} else {
		respondWithJSON(w, http.StatusOK, User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		})
	}
}
