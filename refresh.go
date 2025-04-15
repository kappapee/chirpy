package main

import (
	"internal/auth"
	"net/http"
	"strings"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type Token struct {
		Token string `json:"token"`
	}
	var refreshToken = ""
	if val, ok := r.Header["Authorization"]; ok {
		splits := strings.Split(val[0], " ")
		refreshToken = splits[1]
	} else {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", nil)
		return
	}

	refreshTokenDetails, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Something went wrong", nil)
	}

	token, err := auth.MakeJWT(refreshTokenDetails.UserID, cfg.secretToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	respondWithJSON(w, http.StatusOK, Token{
		Token: token,
	})
}
