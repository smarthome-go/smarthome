package database

import (
	"database/sql"
	"fmt"
)

//
// Homescript owner management
//

func createHomescriptOwnerTable() error {
	_, err := db.Query(fmt.Sprintf(`
	CREATE TABLE
	IF NOT EXISTS
	homescriptPermission(
		Username		VARCHAR(20),
		HomescriptId    VARCHAR(%d),
		FOREIGN KEY (Username)
		REFERENCES user(Username),
		FOREIGN KEY (HomescriptId)
		REFERENCES homescript(Id)
	)`, HOMESCRIPT_ID_LEN))
	if err != nil {
		log.Error("Failed to create homescript permissions table: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

func addHomescriptOwner(username, homescriptId string) (aRowWasInserted bool, err error) {
	query, err := db.Prepare(`
	INSERT INTO homescriptPermission(
		Username, HomescriptId
	) VALUES(?, ?)
	`)

	if err != nil {
		log.Errorf("Failed to insert HMS permission: preparing query failed: %s", err.Error())
		return false, err
	}

	defer query.Close()

	res, err := query.Exec(username, homescriptId)
	if err != nil {
		log.Errorf("Failed to insert HMS permission: executing query failed: %s", err.Error())
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Errorf("Failed to insert HMS permission: getting rows affected failed: %s", err.Error())
		return false, err
	}

	return rows != 0, nil
}

func removeHomescriptOwner(username, homescriptId string) (aRowWasRemoved bool, err error) {
	query, err := db.Prepare(`
	DELETE FROM homescriptPermission
	WHERE homescriptPermission.Username=?
	AND homescriptPermission.HomescriptId=?
	`)
	if err != nil {
		log.Errorf("Failed to remove HMS permission: preparing query failed: %s", err.Error())
		return false, err
	}

	defer query.Close()

	res, err := query.Exec(username, homescriptId)
	if err != nil {
		log.Errorf("Failed to remove HMS permission: executing query failed: %s", err.Error())
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Errorf("Failed to remove HMS permission: getting rows affected failed", err.Error())
		return false, err
	}

	return rows != 0, nil
}

func usernameOwnsHomescript(username, homescriptId string) (bool, error) {
	query, err := db.Prepare(`
	SELECT 1 FROM homescriptPermission
	WHERE homescriptPermission.Username=?
	AND homescriptPermission.HomescriptId=?
	`)

	if err != nil {
		log.Errorf("Failed to check if user has HMS permission: preparing query failed: %s", err.Error())
		return false, err
	}

	defer query.Close()

	row := query.QueryRow(username, homescriptId)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return false, nil
		}

		log.Errorf("Failed to check if user has HMS permission: executing query failed: %s", row.Err().Error())
		return false, err
	}

	return true, nil
}

func listHomescriptIdsOfUser(username string) ([]string, error) {
	query, err := db.Prepare(`
	SELECT
		homescriptPermission.HomescriptId
	FROM homescriptPermission
	WHERE homescriptPermission.Username=?
	`)

	if err != nil {
		log.Errorf("Failed to list homescript IDs of user: preparing query failed", err.Error())
		return nil, err
	}

	defer query.Close()

	res, err := query.Query(username)
	if err != nil {
		log.Errorf("Failed to list homescript IDs of user: executing query failed", err.Error())
		return nil, err
	}

	output := make([]string, 0)
	for res.Next() {
		var buf string
		if err := res.Scan(&buf); err != nil {
			log.Errorf("Failed to list homescript IDs of user: scanning query results failed", err.Error())
			return nil, err
		}
		output = append(output, buf)
	}

	return output, nil
}
