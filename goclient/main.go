package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	oidc "github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

var (
	clientID     = "myclient"
	clientSecret = "t9WKmoayUfzaZZ7rs48pPad4bTA341VL"
)

func main() {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, "http://localhost:8080/realms/myrealm")

	if err != nil {
		log.Fatal(err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "http://localhost:8081/auth/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "roles"},
	}

	state := "foobar" // TODO: generate random state

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, oauth2Config.AuthCodeURL(state), http.StatusFound)
	})

	http.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		token, err := oauth2Config.Exchange(ctx, r.URL.Query().Get("code"))

		if err != nil {
			http.Error(w, "failed to exchange token", http.StatusInternalServerError)
			return
		}

		idToken, ok := token.Extra("id_token").(string)

		if !ok {
			http.Error(w, "no id_token", http.StatusInternalServerError)
			return
		}

		userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(token))

		if err != nil {
			http.Error(w, "failed to get userinfo", http.StatusInternalServerError)
			return
		}

		resp := struct {
			AccessToken *oauth2.Token
			IDToken     string
			UserInfo    *oidc.UserInfo
		}{
			token,
			idToken,
			userInfo,
		}

		data, err := json.Marshal(resp)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(data)
	})

	http.HandleFunc("/teste", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
