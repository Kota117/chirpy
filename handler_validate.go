package main

import (
	"encoding/json"
	"net/http"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	const maxChirpLength = 140

	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Valid bool `json:"valid"`
	}

	var (
		decoder *json.Decoder
		params  parameters
		err     error
	)
	decoder = json.NewDecoder(r.Body)
	params = parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err) // HTTP Status Code 500
		return
	}

	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil) // HTTP Status Code 400
		return
	}

	respondWithJSON(w, http.StatusOK, returnVals{ // HTTP Status Code 200
		Valid: true,
	})
}
