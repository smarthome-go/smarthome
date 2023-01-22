package database

import "database/sql"

// Creates the table containing Homescript storage keys and values
// If the database fails, this function returns an error
func createHomescriptStorageTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	homescriptStorage(
		Owner					VARCHAR(20),
		VarKey					VARCHAR(50),
		VarValue				TEXT,

		PRIMARY KEY (Owner, VarKey),
		CONSTRAINT HomescriptOwnerStor
		FOREIGN KEY (Owner)
		REFERENCES user(Username)
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to create Homescript table: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

func InsertHmsStorageEntry(user string, key string, value string) error {
	query, err := db.Prepare(`
	INSERT INTO homescriptStorage(Owner, VarKey, VarValue)
	VALUES(?, ?, ?)
	ON DUPLICATE KEY
	UPDATE VarValue=VALUES(VarValue)
	`)

	if err != nil {
		log.Error("Could not insert HMS variable entry: Preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()

	if _, err := query.Exec(user, key, value); err != nil {
		log.Error("Could not insert HMS variable entry: Executing query failed: ", err.Error())
		return err
	}

	return nil
}

func GetHmsStorageEntry(user string, key string) (*string, error) {
	query, err := db.Prepare(`
	SELECT VarValue FROM homescriptStorage
	WHERE VarKey=? AND Owner =?
	`)

	if err != nil {
		log.Error("Could not get HMS variable entry: Preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()

	var value string
	if err := query.QueryRow(key, user).Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("Could not get HMS variable entry: Executing query failed: ", err.Error())
		return nil, err
	}

	return &value, nil
}
