package middleware

import (
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var Store *sessions.CookieStore

func Init(useRandomSeed bool) {
	if useRandomSeed {
		Store = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))
	} else {
		// By using a static sting like "", no login is required when restarting the server in non-production mode
		// The session encryption key is static, cookies stay valid
		// If a logout should be enforced during development, enable production mode temporarily
		log.Warn("\x1b[33mUsing static session encryption. This is a security risk.")
		Store = sessions.NewCookieStore([]byte(""))
	}
}
