package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ReuqestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload ReuqestPayload
	err := app.readJSON(w, r, &requestPayload)

	if err != nil {
		app.errorJson(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	default:
		app.errorJson(w, errors.New("unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create some  json  we'll send that to the auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call the auth service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJson(w, err)
	}

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		app.errorJson(w, err)
	}

	defer response.Body.Close()
	// make sure we get the correct status code from the auth service.

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJson(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJson(w, errors.New("error calling auth service"))
		return
	}

	// create a variable  we'll read response.Body into jsonResponse

	var jsonFromService jsonResponse

	// decode the json from the auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)

	if err != nil {
		app.errorJson(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJson(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)

}
