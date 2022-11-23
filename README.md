# Freshify
trim older songs from spotify playlists designed to be run in Github Actions or similar

## Setup
1. Create application at : https://developer.spotify.com/my-applications/
     - Set the redirect URI on the application to "http://localhost:8080/callback" (The path can be different but should be matched to the var in internal/authenticate/authenticate.go)
2. Set the SPOTIFY_ID and SPOTIFY_SECRET environment variables to the Client ID and Client Secret of the application.
3. Run freshify --auth and login to Spotify
4. Set the REFRESH_TOKEN environment variable with the returned values