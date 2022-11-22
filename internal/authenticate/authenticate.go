package authenticate

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dghubble/sling"
	"github.com/pkg/browser"
	auth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var (
	SpotifyAuth = auth.New(auth.WithRedirectURL(redirectURI), auth.WithScopes(auth.ScopePlaylistReadPrivate, auth.ScopePlaylistModifyPublic, auth.ScopePlaylistModifyPrivate))
	ch          = make(chan *oauth2.Token)
	state       = "freshifystate"
	redirectURI = "http://localhost:8080/callback"
)

func Authenticate() error {
	// Setting handler functions for callback. This must match the callback path in redirectURI
	http.HandleFunc("/callback", completeAuth)

	// start a webserver listening on the redirectURI port
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println("Please log in to Spotify, opening browser...")
	time.Sleep(3 * time.Second)

	url := SpotifyAuth.AuthURL(state)
	browser.OpenURL(url)

	// Trigger on auth complete
	tok := <-ch

	// fmt.Printf("your access token is: %s\n", tok.AccessToken)
	fmt.Printf("your refresh token is: %s\n", tok.RefreshToken)

	return nil
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := SpotifyAuth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	ch <- tok
}

type RefreshRequest struct {
	GrantType    string `url:"grant_type,omitempty"`
	RefreshToken string `url:"refresh_token,omitempty"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
}

type RefreshError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func RefreshToken(rtoken, clientId, clientSecret string) (string, error) {
	authstring := base64.URLEncoding.EncodeToString([]byte(clientId + ":" + clientSecret))
	body := &RefreshRequest{
		GrantType:    "refresh_token",
		RefreshToken: rtoken,
	}
	resp := &RefreshResponse{}
	respErr := &RefreshError{}

	slingClient := sling.New().
		Base("https://accounts.spotify.com/api/").
		Set("Authorization", "Basic "+authstring)

	req, err := slingClient.Post("token").BodyForm(body).Receive(resp, respErr)
	if err != nil {
		return "", err
	}

	if req.StatusCode != 200 {
		return "", fmt.Errorf("token refresh error: %s\n", req.Status)
	}

	fmt.Printf("Status: %s\n", req.Status)

	// if respErr != nil {
	// 	log.Fatalf("token refresh error: %v\n", respErr)
	// }

	fmt.Printf("Got Token: %v\n", resp)

	return resp.AccessToken, nil
}
