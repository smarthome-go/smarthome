package database

// Used in user.go
type User struct {
	Username string
	Password string
}

func GetPermissions() []string {
	permissions := []string{
		"authentication", // If the user is allowed to authenticate and login, if disabled, a user is `disabled`
		"foo",            // A testing purpose permission
		"bar",            // A testing purpose permission
	}
	return permissions
}
