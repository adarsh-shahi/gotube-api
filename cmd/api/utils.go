package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/adarsh-shahi/gotube-api/internals/db"
	"github.com/golang-jwt/jwt/v5"
)

type jsonResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func (app *appConfig) writeJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) (error, int) {
	log.Println("in write json")
	log.Println(data)
	out, err := json.Marshal(data)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(out)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, -1
}

func (app *appConfig) readJSON(w http.ResponseWriter, r *http.Request, data interface{}) (error, int) {
	format := r.Header.Get("Content-Type")
	if format != "" {
		if format != "application/json" {
			return errors.New("Content-Type header is not application/json"), http.StatusUnsupportedMediaType
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(app.jsonSizeLimit))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(data)
	if err != nil {
		switch {
		case err.Error() == "http: request body too large":
			return errors.New(
				fmt.Sprintf("Request body must not be large than %d", app.jsonSizeLimit),
			), http.StatusRequestEntityTooLarge
		case errors.Is(err, io.EOF):
			return errors.New("Request body must not be empty"), http.StatusBadRequest

		default: return errors.New("check the json data passed or contact the backend dev"), http.StatusInternalServerError
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF){
		return errors.New("Request body must contain a single JSON object"), http.StatusBadRequest
	}

	return nil, -1
}

func (app *appConfig) errorJSON(w http.ResponseWriter, err error, status ...int) (error, int) {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	resopnse := jsonResponse{
		Error:   true,
		Message: err.Error(),
	}
	return app.writeJSON(w, statusCode, resopnse)
}

func (app *appConfig) generateToken(payload db.TIdEmailPassword) (string, error){
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": payload.Id,
		"utype": payload.UType,
		"email": payload.Email,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := t.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

