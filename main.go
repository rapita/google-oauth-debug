package main

import (
	"context"
	"encoding/json"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
var (
	k         = koanf.New(".")
	oauthConf = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "",
		Scopes:       []string{},
		Endpoint:     google.Endpoint,
	}
)

type TokenFormatted struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Expiry       int64  `json:"expires_in,omitempty"`
}

func initOauth() {
	oauthConf.ClientID = k.String("client_id")
	oauthConf.ClientSecret = k.String("client_secret")
	oauthConf.RedirectURL = "http://localhost:" + k.String("port") + k.String("callback_route")
	oauthConf.Scopes = k.Strings("scopes")
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<html>
	<head>
		<title>Google OAuth-2 Debug</title>
	</head>
	<body>
		<h2>Google OAuth-2 Debug</h2>
		<a href="/login">Google</a>
	</body>
</html>
`))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("Request: " + r.URL.String())
	URL, err := url.Parse(oauthConf.Endpoint.AuthURL)
	if err != nil {
		log.Fatal("parse: " + err.Error())
	}

	parameters := url.Values{}
	parameters.Add("client_id", oauthConf.ClientID)
	parameters.Add("scope", strings.Join(oauthConf.Scopes, " "))
	parameters.Add("redirect_uri", oauthConf.RedirectURL)
	parameters.Add("response_type", "code")
	URL.RawQuery = parameters.Encode()
	redirectUrl := URL.String()

	log.Println("Redirect: " + redirectUrl)

	http.Redirect(w, r, redirectUrl, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("Request: " + r.URL.String())

	code := r.FormValue("code")
	log.Println("Code: " + code)

	if code == "" {
		log.Println("Code not found..")
		w.Write([]byte("Code Not Found to provide AccessToken..\n"))
		reason := r.FormValue("error_reason")
		if reason == "user_denied" {
			w.Write([]byte("User has denied Permission.."))
		}
		return
	}

	token, err := oauthConf.Exchange(context.Background(), code)
	if err != nil {
		log.Println("oauthConf.Exchange() failed with " + err.Error() + "\n")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	jsonToken, err := json.Marshal(TokenFormatted{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry.Unix(),
	})
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	log.Println("JSON Token: " + string(token.AccessToken))

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonToken)
	return
}

func main() {
	// Load JSON config.
	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	initOauth()

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc(k.String("callback_route"), handleCallback)

	log.Println("Started on http://localhost:" + k.String("port"))
	log.Fatal(http.ListenAndServe(":"+k.String("port"), nil))
}
