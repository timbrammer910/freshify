package authenticate

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pkg/browser"
	"github.com/zmb3/spotify/v2"
	auth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var (
	spotifyAuth = auth.New(auth.WithRedirectURL(redirectURI), auth.WithScopes(auth.ScopePlaylistReadPrivate, auth.ScopePlaylistModifyPublic, auth.ScopePlaylistModifyPrivate))
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

	url := spotifyAuth.AuthURL(state)
	browser.OpenURL(url)

	// Trigger on auth complete
	tok := <-ch

	fmt.Printf("your access token is: %s", tok.AccessToken)
	fmt.Printf("your refresh token is: %s", tok.RefreshToken)

	return nil
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := spotifyAuth.Token(r.Context(), state, r)
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

func GetAccessToken(atoken, rtoken string) (*spotify.Client, error) {
	tok := &oauth2.Token{
		AccessToken:  atoken,
		RefreshToken: rtoken,
	}

	// build client with refresh token and test
	client := spotify.New(spotifyAuth.Client(context.Background(), tok))
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	return client, nil
}
