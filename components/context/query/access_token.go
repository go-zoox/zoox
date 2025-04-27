package query

import (
	"fmt"
	"strings"
)

// ConstantsQueryAccessTokenKeys is the keys that are used to identify the access token.
var ConstantsQueryAccessTokenKeys = []string{
	"access_token",
	"accessToken",
}

// AccessToken returns the access token.
func (q *query) AccessToken() (token string, err error) {
	for _, key := range ConstantsQueryAccessTokenKeys {
		if value, ok := q.request.URL.Query()[key]; ok {
			return value[0], nil
		}
	}

	return "", fmt.Errorf("access token not found(keys: %s)", strings.Join(ConstantsQueryAccessTokenKeys, ", "))
}

// MustAccessToken returns the access token.
func (q *query) MustAccessToken() string {
	token, err := q.AccessToken()
	if err != nil {
		panic(err)
	}
	return token
}
