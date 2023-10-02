package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (app *appConfig) enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Acess-Control-Allow-Origin", "http://localhost:3000")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token, Authorization")
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

type ParsedUserData struct {
	Email string `json:"email"`
	Id int `json:"id"`
	UType string `json:"utype"`
}
var parsedUserData ParsedUserData

func (app *appConfig) protect(next func (w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request){

		// Extract Authorization header
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			app.errorJSON(w, errors.New("Unauthorized, login or signup again"), http.StatusUnauthorized)
			return
		}
		jwtToken := authHeader[1]

		// parse jwt token
		token ,err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header)
			}
			return []byte("secret"), nil
		})

		// extract payload data
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid{
			if float64(time.Now().Unix()) > claims["exp"].(float64){
				app.errorJSON(w, errors.New("session expired please login again"), http.StatusUnauthorized)
			}
			parsedUserData.Email = claims["email"].(string)
			parsedUserData.Id = int(claims["id"].(float64))
			parsedUserData.UType = claims["utype"].(string)
		}
		if err != nil {
			app.errorJSON(w, err, http.StatusUnauthorized)
			return 
		}

		next(w,r)
	})
}
