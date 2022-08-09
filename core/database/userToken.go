package database

import "database/sql"

type UserToken struct {
	User  string        `json:"user"`
	Token string        `json:"token"`
	Data  UserTokenData `json:"data"`
}

type UserTokenData struct {
	Label string `json:"label"`
}

func createUserTokenTable() error {
	if _, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	userToken(
		Token CHAR(50),
		User  VARCHAR(20),
		Label VARCHAR(50),
		PRIMARY KEY(Token),
		FOREIGN KEY (User)
		REFERENCES user(Username)
	)
	`); err != nil {
		log.Error("Failed to create user token table: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Inserts a new token into the table
// Validation is required beforehand
func InsertUserToken(
	token string,
	user string,
	label string,
) error {
	query, err := db.Prepare(`
	INSERT INTO
	userToken(
		Token,
		User,
		Label
	)
	VALUES(?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to insert into user tokens: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(
		token,
		user,
		label,
	); err != nil {
		log.Error("Failed to insert into user tokens: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns the list of available authentication tokens a user has set up
func GetUserTokensOfUser(username string) ([]UserToken, error) {
	query, err := db.Prepare(`
	SELECT
		Token,
		User,
		Label
	FROM userToken
	WHERE User=?
	`)
	if err != nil {
		log.Error("Failed to get user tokens of user: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query(username)
	if err != nil {
		log.Error("Failed to get user tokens of user: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()
	tokens := make([]UserToken, 0)
	for res.Next() {
		var row UserToken
		if err := res.Scan(
			&row.Token,
			&row.User,
			&row.Data.Label,
		); err != nil {
			log.Error("Failed to get user tokens of user: scanning results failed: ", err.Error())
			return nil, err
		}
		tokens = append(tokens, row)
	}
	return tokens, nil
}

// Returns an arbitrary user token which matches the query
// Used when validating a token during authentication
func GetUserTokenByToken(token string) (data UserToken, found bool, err error) {
	query, err := db.Prepare(`
	SELECT
		Token,
		User,
		Label
	FROM userToken
	WHERE Token=?
	`)
	if err != nil {
		log.Error("Failed to get user token by token: preparing query failed: ", err.Error())
		return UserToken{}, false, err
	}
	defer query.Close()
	if err := query.QueryRow(token).Scan(
		&data.Token,
		&data.User,
		&data.Data.Label,
	); err != nil {
		if err == sql.ErrNoRows {
			return UserToken{}, false, nil
		}
		log.Error("Failed to get user token by token: scanning query results failed: ", err.Error())
		return UserToken{}, false, err
	}
	return data, true, nil
}

// Deletes an arbitrary user token
func DeleteTokenByToken(token string) error {
	query, err := db.Prepare(`
	DELETE FROM
	userToken
	WHERE Token=?
	`)
	if err != nil {
		log.Error("Failed to delete user token by token: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(token); err != nil {
		log.Error("Failed to delete user token by token: executing query failed: ", err.Error())
		return err
	}
	return nil
}
