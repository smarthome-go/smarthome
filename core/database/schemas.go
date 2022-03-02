package database

// Used in user.go
type User struct {
	Username string
	Password string
}

type Permission struct {
	Permission  string
	Name        string
	Description string
}

func GetPermissions() []Permission {
	permissions := []Permission{
		{
			// If the user is allowed to authenticate and login, if disabled, a user is `disabled`
			Permission:  "authentication",
			Name:        "Authentication",
			Description: "Allows the user to authenticate",
		},
	}
	return permissions
}
