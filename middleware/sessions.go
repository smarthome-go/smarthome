package middleware

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gorilla/sessions"
)

var Store *sessions.CookieStore

func StringWithCharset(length int, charset string) string {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	randomBytes := make([]byte, length)
	for i := range randomBytes {
		randomBytes[i] = charset[seededRand.Intn(len(charset))]
	}
	log.Trace(fmt.Sprintf("Generated random seed for sessions: %s", string(randomBytes)))
	return string(randomBytes)
}

func Init(useRandomSeed bool) {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" +
		"1234567890"
	if useRandomSeed {
		Store = sessions.NewCookieStore([]byte(StringWithCharset(40, charset)))
	} else {
		// By using a static sting like "", no login is required when restarting the server in non-production mode
		// The session encryption key is static, cookies stay valid
		// If a logout should be enforced during development, enable production mode temporarily
		log.Warn("\x1b[33mUsing a static string for session encryption. This is a security risk and should not be used in production.")
		Store = sessions.NewCookieStore([]byte(""))
	}
}
