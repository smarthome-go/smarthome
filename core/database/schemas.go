package database

// Used in user.go
type User struct {
	Username string
	Password string
}

// Permission-related

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

// Rooms and Switches
type Switch struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	RoomId string `json:"roomId"`
}

type Room struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Switches    []Switch `json:"switches"`
}
