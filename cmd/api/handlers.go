package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/adarsh-shahi/gotube-api/internals/db"
	"github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/aws/credentials"
	"golang.org/x/crypto/bcrypt"
)

func (app *appConfig) home(w http.ResponseWriter, r *http.Request) {
	log.Println("in home")
	app.writeJSON(w, http.StatusAccepted, "i did it")
}

func (app *appConfig) login(w http.ResponseWriter, r *http.Request) {
	userToFind := db.TIdEmailPassword{}
	err, statusCode := app.readJSON(w, r, &userToFind)
	if err != nil {
		app.errorJSON(w, err, statusCode)
		return
	}
	user, err := app.DB.GetIdEmailPasswordUser(userToFind)
	if err != nil {
		app.errorJSON(w, err, http.StatusNotFound)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userToFind.Password))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			app.errorJSON(w, errors.New("wrong password"), http.StatusBadRequest)
		default:
			app.errorJSON(w, err, http.StatusUnauthorized)
		}
		return
	}

	token, err := app.generateToken(*user)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	log.Println(user)
	jsonResponse := jsonResponse{
		Error:   false,
		Message: "loggedIn successfully",
		Data: struct {
			Token string `json:"token"`
		}{Token: token},
	}

	app.writeJSON(w, http.StatusAccepted, jsonResponse)
}

func (app *appConfig) signup(w http.ResponseWriter, r *http.Request) {
	user := db.AddUser{}
	err, statusCode := app.readJSON(w, r, &user)
	if err != nil {
		app.errorJSON(w, err, statusCode)
		return
	}
	log.Println(user)

	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
	}
	user.Password = string(bytes)
	app.DB.AddUser(user)
	if err != nil {
		app.errorJSON(w, err, statusCode)
	}

	response := jsonResponse{
		Error:   false,
		Message: "account created",
	}
	app.writeJSON(w, http.StatusCreated, response)
}

func (app *appConfig) sendInvite(w http.ResponseWriter, r *http.Request) {
	log.Println(parsedUserData)
	if parsedUserData.UType != "owner" {
		app.errorJSON(w, errors.New("only owners can send invites"), http.StatusBadRequest)
		return
	}
	userToFind := db.TIdEmailPassword{}
	userToFind.UType = "user"
	err, statusCode := app.readJSON(w, r, &userToFind)
	if err != nil {
		app.errorJSON(w, err, statusCode)
		return
	}
	if userToFind.Role == "" {
		userToFind.Role = "viewer"
	}
	user, err := app.DB.GetIdEmailPasswordUser(userToFind)
	if err != nil {
		app.errorJSON(w, err, http.StatusNotFound)
		return
	}
	err = app.DB.AddInvites(parsedUserData.Id, user.Id, user.Role)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	response := jsonResponse{
		Error:   false,
		Message: "invitation sent",
	}
	app.writeJSON(w, http.StatusAccepted, response)
}

func (app *appConfig) updateInviteRole(w http.ResponseWriter, r *http.Request) {
	if parsedUserData.UType != "owner" {
		app.errorJSON(w, errors.New("only owners can edit invites"), http.StatusBadRequest)
		return
	}
	editUserRole := struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}{}
	err, statusCode := app.readJSON(w, r, &editUserRole)
	if err != nil {
		app.errorJSON(w, err, statusCode)
		return
	}

	userToFind := db.TIdEmailPassword{
		Email: editUserRole.Email,
		Role:  editUserRole.Role,
	}

	user, err := app.DB.GetIdEmailPasswordUser(userToFind)
	if err != nil {
		app.errorJSON(w, err, http.StatusNotFound)
		return
	}

	err = app.DB.UpdateInviteRole(parsedUserData.Id, user.Id, editUserRole.Role)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	response := jsonResponse{
		Error:   false,
		Message: "role updated.",
	}
	app.writeJSON(w, http.StatusAccepted, response)
}

func (app *appConfig) deleteInvite(w http.ResponseWriter, r *http.Request) {
	if parsedUserData.UType == "user" {
		response := struct {
			Accept bool   `json:"accept"`
			Email  string `json:"email"`
			Role   string `json:"role"`
		}{}
		err, statusCode := app.readJSON(w, r, &response)
		if err != nil {
			app.errorJSON(w, err, statusCode)
			return
		}
		ownerToFind := db.TIdEmailPassword{
			Email: response.Email,
			UType: "owner",
		}
		log.Println(ownerToFind)
		owner, err := app.DB.GetIdEmailPasswordUser(ownerToFind)
		if err != nil {
			app.errorJSON(w, err, http.StatusBadRequest)
			return
		}
		log.Println(owner)
		if response.Accept {
			// apply database transaction
			err := app.DB.DeleteInviteAndAddToTeam(owner.Id, parsedUserData.Id, response.Role)
			if err != nil {
				app.errorJSON(w, err, http.StatusBadRequest)
				return
			}
		} else {
			err := app.DB.DeleteInvite(owner.Id, parsedUserData.Id)
			if err != nil {
				app.errorJSON(w, err, http.StatusBadRequest)
				return
			}
		}

	} else {
		response := struct {
			Email string `json:"email"`
		}{}
		err, statusCode := app.readJSON(w, r, &response)
		if err != nil {
			app.errorJSON(w, err, statusCode)
			return
		}
		userToFind := db.TIdEmailPassword{
			Email: response.Email,
		}
		user, err := app.DB.GetIdEmailPasswordUser(userToFind)
		if err != nil {
			app.errorJSON(w, err, http.StatusBadRequest)
			return
		}
		err = app.DB.DeleteInvite(parsedUserData.Id, user.Id)
		if err != nil {
			app.errorJSON(w, err, http.StatusBadRequest)
			return
		}
	}
	app.writeJSON(w, http.StatusAccepted, "done.")
}

func (app *appConfig) getInvites(w http.ResponseWriter, r *http.Request) {
	list := []db.Invites{}
	response := jsonResponse{}
	var err error
	if parsedUserData.UType == "owner" {
		response.Message = "sent"
		list, err = app.DB.GetAllInvites(parsedUserData.Id, "owner")
	} else {
		response.Message = "received"
		list, err = app.DB.GetAllInvites(parsedUserData.Id, "user")
	}
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	response.Error = false
	response.Data = list
	app.writeJSON(w, http.StatusOK, response)
}

func (app *appConfig) uploadVideo(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("image")
	if err != nil {
		app.errorJSON(w, errors.New("error retrieving the file"), http.StatusBadRequest)
		return
	}
	defer file.Close()

	extension := filepath.Ext(handler.Filename)
	newFileName := "uploaded_image" + extension

	outputFile, err := os.Create(newFileName)
	if err != nil {
		app.errorJSON(w, errors.New("error creating the file"), http.StatusBadRequest)
		return
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, file)
	if err != nil {
		app.errorJSON(w, errors.New("error copying the file"), http.StatusBadRequest)
		return
	}
	w.Write([]byte("hello there"))
}

func (app *appConfig) getSignedUrl(w http.ResponseWriter, r *http.Request) {

	creds := credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"),os.Getenv("AWS_SECRET_ACCESS_KEY") , "")

	cfg := aws.NewConfig().WithRegion(os.Getenv("AWS_REGION")).WithCredentials(creds)

	sess, err := session.NewSession(cfg)

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String("gotube.adarsh"),
		Key:    aws.String("memes/setup.jpg"),
	})

	urlStr, err := req.Presign(1 * time.Minute)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	response := jsonResponse{
		Error: false,
		Data:  urlStr,
	}

	app.writeJSON(w, http.StatusAccepted, response)
}

func (app *appConfig) putSignedUrl(w http.ResponseWriter, r *http.Request) {
}
