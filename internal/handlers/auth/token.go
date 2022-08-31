package auth

import "net/http"

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	// if we get here then we have teh jwt in token
	// next we want to write a custom middle ware as  a wrapper for the jwt validator
	// and validate the refresh token, and custom claims
	// at this point we can generate a new set, and return them as cookies
	// et voiala
}
