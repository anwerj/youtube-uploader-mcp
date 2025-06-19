package youtube

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"os/user"
// 	"path/filepath"

// 	"golang.org/x/oauth2"
// )

// const fTokenFileName = ".youtube_uploader_access_token_cache"

// // SaveToken saves the OAuth2 token to a file in the user's home directory.
// // The token is saved in JSON format, and the file is created if it does not exist.
// func SaveToken(token *oauth2.Token) error {
// 	if token == nil || token.AccessToken == "" {
// 		return fmt.Errorf("invalid token: token or access token is empty")
// 	}

// 	current, err := user.Current()
// 	if err != nil {
// 		return fmt.Errorf("failed to get current user: %s", err.Error())
// 	}
// 	fTokenPath := filepath.Join(current.HomeDir, fTokenFileName)
// 	// open file or create it
// 	file, err := os.OpenFile(fTokenPath, os.O_RDWR|os.O_CREATE, 0600)
// 	if err != nil {
// 		return fmt.Errorf("failed to open or create token file: %s", err.Error())
// 	}
// 	defer file.Close()
// 	// write the access token to the file
// 	err = json.NewEncoder(file).Encode(token)
// 	if err != nil {
// 		return fmt.Errorf("failed to write access token to file: %s", err.Error())
// 	}
// 	return nil
// }

// // ReadToken reads the OAuth2 token from a file in the user's home directory.
// // It returns the token if it exists and is valid, or an error if the file cannot be read or the token is invalid.
// // The token is expected to be in JSON format.
// func ReadToken() (*oauth2.Token, error) {
// 	current, err := user.Current()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get current user: %s", err.Error())
// 	}
// 	fTokenPath := filepath.Join(current.HomeDir, fTokenFileName)
// 	file, err := os.Open(fTokenPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open token file: %s", err.Error())
// 	}
// 	defer file.Close()

// 	var token oauth2.Token
// 	err = json.NewDecoder(file).Decode(&token)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to decode access token from file: %s", err.Error())
// 	}

// 	return &token, nil
// }
