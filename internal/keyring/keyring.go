package keyring

import keyring "github.com/zalando/go-keyring"

const service = "i3-barista"

// Get secret for user on the i3-barista service in the keyring.
func Get(user string) (string, error) {
	return keyring.Get(service, user)
}

// MustGet behaves like Get but panics if reading the secret fails for some
// reason.
func MustGet(user string) string {
	secret, err := Get(user)
	if err != nil {
		panic(err)
	}

	return secret
}

// Set secret for user on the i3-barista service in the keyring.
func Set(user, secret string) error {
	return keyring.Set(service, user, secret)
}
