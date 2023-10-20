package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/adarsh-shahi/gotube-api/internals/db"
	"github.com/adarsh-shahi/gotube-api/internals/youtube"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/crypto/bcrypt"
)

func (app *appConfig) home(w http.ResponseWriter, r *http.Request) {
	log.Println("in home")
	app.writeJSON(w, http.StatusAccepted, "i did it")
}

func (app *appConfig) authYoutube(w http.ResponseWriter, r *http.Request) {
}

func (app *appConfig) signupOauth(w http.ResponseWriter, r *http.Request) {
	userType := r.URL.Query()["state"][0]
	if userType == "owner" {
		app.addCreator(w, r)
	} else if userType == "user" {
		app.addUser(w, r)
	}
}

func (app *appConfig) addCreator(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query()["code"][0]
	urlString := fmt.Sprintf(
		"https://oauth2.googleapis.com/token?code=%s&client_id=%s&client_secret=%s&redirect_uri=%s&grant_type=%s",
		code,
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
		os.Getenv("GOOGLE_REDIRECT_URI"),
		"authorization_code",
	)

	resp, err := http.Post(urlString, "application/x-www-form-urlencoding", bytes.NewBuffer([]byte("")))
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
	}
	defer resp.Body.Close()
	response := map[string]interface{}{}
	json.NewDecoder(resp.Body).Decode(&response)
	fmt.Println(response)
	access_token := response["access_token"].(string)
	id_token := response["id_token"].(string)
	refresh_token := response["refresh_token"].(string)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v1/userinfo?alt=json&access_token="+access_token, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", id_token))
	respUser, err := client.Do(req)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	defer respUser.Body.Close()
	googleUser := map[string]interface{}{}
	json.NewDecoder(respUser.Body).Decode(&googleUser)
	email := googleUser["email"].(string)

	isEmailThere, id, err := app.DB.GetEmailFromOwner(email)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	if !isEmailThere {
		ch, err := youtube.GetChannelInfo(access_token)
		addOwner := db.AddOwner{
			Channel: youtube.Channel{
				Title:           ch.Title,
				Description:     ch.Description,
				CustomUrl:       ch.CustomUrl,
				ProfileImageUrl: ch.ProfileImageUrl,
			},
			Email: email,
		}
		id, err = app.DB.AddOwner(addOwner)
		if err != nil {
			app.errorJSON(w, err, http.StatusInternalServerError)
		}
	}

	token, err := app.generateToken(email, "owner", id)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	responseData := jsonResponse{
		Error:   false,
		Message: "user created",
		Data: struct {
			JwtToken     string `json:"jwtToken"`
			RefreshToken string `json:"refreshToken"`
			UType        string `json:"uType"`
			Email        string `json:"email"`
		}{
			JwtToken:     token,
			RefreshToken: refresh_token,
			UType:        "owner",
			Email:        email,
		},
	}

	byteData, err := json.Marshal(responseData)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	params := url.Values{}
	params.Add("jsonData", string(byteData))
	http.Redirect(w, r, "http://localhost:3000/auth?"+params.Encode(), http.StatusSeeOther)
}

func (app *appConfig) addUser(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query()["code"][0]
	urlString := fmt.Sprintf(
		"https://oauth2.googleapis.com/token?code=%s&client_id=%s&client_secret=%s&redirect_uri=%s&grant_type=%s",
		code,
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
		os.Getenv("GOOGLE_REDIRECT_URI"),
		"authorization_code",
	)

	resp, err := http.Post(urlString, "application/x-www-form-urlencoding", bytes.NewBuffer([]byte("")))
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
	}
	defer resp.Body.Close()
	response := map[string]interface{}{}
	json.NewDecoder(resp.Body).Decode(&response)
	fmt.Println(response)
	access_token := response["access_token"].(string)
	id_token := response["id_token"].(string)
	refresh_token := response["refresh_token"].(string)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v1/userinfo?alt=json&access_token="+access_token, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", id_token))
	respUser, err := client.Do(req)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	defer respUser.Body.Close()
	googleUser := map[string]interface{}{}
	json.NewDecoder(respUser.Body).Decode(&googleUser)
	email := googleUser["email"].(string)
	user := db.AddUser{
		Email: email,
	}
	id, err := app.DB.AddUser(user)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	token, err := app.generateToken(email, "user", id)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	responseData := jsonResponse{
		Error:   false,
		Message: "user created",
		Data: struct {
			JwtToken     string `json:"jwtToken"`
			RefreshToken string `json:"refreshToken"`
			UType        string `json:"uType"`
			Email        string `json:"email"`
		}{
			JwtToken:     token,
			RefreshToken: refresh_token,
			UType:        "user",
			Email:        email,
		},
	}
	byteData, err := json.Marshal(responseData)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	params := url.Values{}
	params.Add("jsonData", string(byteData))
	http.Redirect(w, r, "http://localhost:3000/auth?"+params.Encode(), http.StatusSeeOther)
}

func (app *appConfig) login(w http.ResponseWriter, r *http.Request) {
	userToFind := db.TIdEmailPassword{}
	err, statusCode := app.readJSON(w, r, &userToFind)
	if err != nil {
		app.errorJSON(w, err, statusCode)
		return
	}
	user, err := app.DB.GetIdEmailUser(userToFind)
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

	// token, err := app.generateToken(*user)
	// if err != nil {
	// 	app.errorJSON(w, err, http.StatusInternalServerError)
	// 	return
	// }

	log.Println(user)
	jsonResponse := jsonResponse{
		Error:   false,
		Message: "loggedIn successfully",
		Data: struct {
			Token string `json:"token"`
			UType string `json:"utype"`
		}{Token: "", UType: userToFind.UType},
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
	user, err := app.DB.GetIdEmailUser(userToFind)
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

	user, err := app.DB.GetIdEmailUser(userToFind)
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
		owner, err := app.DB.GetIdEmailUser(ownerToFind)
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
		user, err := app.DB.GetIdEmailUser(userToFind)
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
	fmt.Println(parsedUserData)
	if parsedUserData.UType == "owner" {
		response.Message = "sent"
		list, err = app.DB.GetAllInvites(parsedUserData.Id, "owner")
		fmt.Println("-----------------")
		fmt.Println("-----------------")
		fmt.Println(list)
		fmt.Println("-----------------")
		fmt.Println("-----------------")
		if err != nil {
			app.errorJSON(w, err, http.StatusBadRequest)
			return
		}
	} else {
		response.Message = "received"
		list, err = app.DB.GetAllInvites(parsedUserData.Id, "user")
		if err != nil {
			app.errorJSON(w, err, http.StatusBadRequest)
			return
		}
	}
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	response.Error = false
	response.Data = list
	app.writeJSON(w, http.StatusOK, response)
}

func (app *appConfig) addContent(w http.ResponseWriter, r *http.Request) {
	content := struct {
		Name string `json:"name"`
	}{}
	err, statusCode := app.readJSON(w, r, &content)

	if err != nil {
		app.errorJSON(w, err, statusCode)
		return
	}
	log.Println(content)

	err = app.DB.AddContent(content.Name)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	app.writeJSON(w, http.StatusOK, jsonResponse{Error: false, Message: "added"})
}

func (app *appConfig) getContent(w http.ResponseWriter, r *http.Request) {
	if parsedUserData.UType == "owner" {
		contentList, err := app.DB.GetContentList(parsedUserData.Id)
		if err != nil {
			app.errorJSON(w, err, http.StatusInternalServerError)
			return
		}
		app.writeJSON(w, http.StatusAccepted, jsonResponse{Error: false, Message: "heres your list", Data: contentList})

	} else {
		ownerId := r.URL.Query().Get("id")
		fmt.Println(ownerId)
		var ownerIdInt int64
		if id, err := strconv.Atoi(ownerId); err != nil {
			app.errorJSON(w, errors.New(fmt.Sprintf("%d not accepted must provide a valid id", ownerIdInt)), http.StatusBadRequest)
			return
		} else {
			ownerIdInt = int64(id)
		}
		fmt.Println(ownerIdInt)
		if ok, err := app.DB.IsTeamMember(ownerIdInt, parsedUserData.Id); err != nil {
			app.errorJSON(w, err, http.StatusInternalServerError)
			return
		} else if !ok {
			app.errorJSON(w, errors.New("you are not a team member"), http.StatusBadRequest)
			return
		}
		contentList, err := app.DB.GetContentList(ownerIdInt)
		if err != nil {
			app.errorJSON(w, err, http.StatusInternalServerError)
			return
		}
		app.writeJSON(w, http.StatusAccepted, jsonResponse{Error: false, Message: "heres your list", Data: contentList})

	}
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
	creds := credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "")

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
	creds := credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "")

	cfg := aws.NewConfig().WithRegion(os.Getenv("AWS_REGION")).WithCredentials(creds)

	sess, err := session.NewSession(cfg)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)

	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String("gotube.adarsh"),
		Key:    aws.String("cycle/ok.jpg"),
	})

	urlStr, err := req.Presign(2 * time.Minute)
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

func (app *appConfig) getProfile(w http.ResponseWriter, r *http.Request) {
	if parsedUserData.UType == "owner" {
		user, err := app.DB.GetOwnerProfile(parsedUserData.Email)
		fmt.Println("-------------------")
		fmt.Println(user)
		if err != nil {
			app.errorJSON(w, err, http.StatusBadRequest)
			return
		}
		response := jsonResponse{
			Error:   false,
			Message: "success",
			Data:    user,
		}
		app.writeJSON(w, http.StatusOK, response)
	}
}

func (app *appConfig) getTeams(w http.ResponseWriter, r *http.Request) {
	teamsList, err := app.DB.GetTeams(parsedUserData.Id)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	response := jsonResponse{
		Error:   false,
		Message: "heres the list",
		Data:    teamsList,
	}
	app.writeJSON(w, http.StatusOK, response)
}

func (app *appConfig) createContent(w http.ResponseWriter, r *http.Request) {
	if parsedUserData.UType != "owner" {
		app.errorJSON(w, errors.New("only owners can create content"), http.StatusBadRequest)
		return
	}
	content := struct {
		Name string `json:"name"`
	}{}
	err, statusCode := app.readJSON(w, r, &content)
	if err != nil {
		app.errorJSON(w, err, statusCode)
		return
	}
	err = app.DB.CreateContent(content.Name, parsedUserData.Id)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	response := jsonResponse{
		Error:   false,
		Message: "content created",
	}
	app.writeJSON(w, http.StatusCreated, response)
}

