package middleware

import (
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var Store *sessions.CookieStore

func InitWithManualKey(randomSeed string) {
	// By using a static string,  no login is required when restarting the server
	// In this case the session encryption key is static, cookies stay valid
	// If a logout should be enforced during development, the key must be changed or ommited
	Store = sessions.NewCookieStore([]byte(randomSeed))
	log.Debug("Successfully initialized middleware session store using manual seed")
}

func InitWithRandomKey() {
	Store = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))
	Store = sessions.NewCookieStore([]byte(""))
	log.Debug("Successfully initialized middleware session store using random key")
}
