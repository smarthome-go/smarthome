package database

import "database/sql"

type HmsArgInputType string

var (
	String  HmsArgInputType = "string"
	Number  HmsArgInputType = "number"
	Boolean HmsArgInputType = "boolean"
)

type HmsArgDisplay string

var (
	TypeDefault    HmsArgDisplay = "type_default"
	StringSwitches HmsArgDisplay = "string_switches"
	BooleanYesNo   HmsArgDisplay = "boolean_yes_no"
	BooleanOnOff   HmsArgDisplay = "boolean_on_off"
	NumberHour     HmsArgDisplay = "number_hour"
	NumberMinute   HmsArgDisplay = "number_minute"
)

type HomescriptArg struct {
	Id           uint              `json:"id"`
	HomescriptId string            `json:"homescriptId"`
	Data         HomescriptArgData `json:"data"`
}

type HomescriptArgData struct {
	Prompt    string          `json:"prompt"`
	InputType HmsArgInputType `json:"inputType"`
	Display   HmsArgDisplay   `json:"display"`
}

// Used for creating the table which contains the arguments of Homescripts
// The `GuiDisplay` value is just used as a hint for user-interfaces for displaying a better selection of possible predefined values
func createHomescriptArgTable() error {
	if _, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	homescriptArg(
		Id INT AUTO_INCREMENT,
		HomescriptId VARCHAR(30),
		Prompt TEXT,
		InputType ENUM(
			'string',
			'number',
			'boolean'
		),
		Display ENUM(
			'type_default',
			'string_switches',
			'boolean_yes_no',
			'boolean_on_off',
			'number_hour',
			'number_minute'
		),
		PRIMARY KEY(Id),
		FOREIGN KEY (HomescriptId)
		REFERENCES homescript(Id)
	)
	`); err != nil {
		log.Error("Could not create HomescriptArgs table: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns all HomescriptArgs of a given user as a slice
func ListAllHomescriptArgsOfUser(username string) ([]HomescriptArg, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		HomescriptId,
		Prompt,
		InputType,
		Display
	FROM homescriptArg
	JOIN homescript
		ON homescriptArgs.HomescriptId=homescript.Id
	WHERE homescript.Owner=?
	`)
	if err != nil {
		log.Error("Failed to list HomescriptArgs of user: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query(username)
	if err != nil {
		log.Error("Failed to list HomescriptArgs of user: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()
	args := make([]HomescriptArg, 0)
	for res.Next() {
		var currentArg HomescriptArg
		if err := res.Scan(
			&currentArg.Id,
			&currentArg.HomescriptId,
			&currentArg.Data.Prompt,
			&currentArg.Data.InputType,
			&currentArg.Data.Display,
		); err != nil {
			log.Error("Failed to list HomescriptArgs of user: scanning results failed: ", err.Error())
			return nil, err
		}
	}
	return args, nil
}

// Returns the arguments of a given Homescript slice
func ListArgsOfHomescript(homescriptId string) ([]HomescriptArg, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		HomescriptId,
		Prompt,
		InputType,
		Display
	FROM homescriptArg
	WHERE HomescriptId=?
	`)
	if err != nil {
		log.Error("Failed to list HomescriptArgs of script: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query(homescriptId)
	if err != nil {
		log.Error("Failed to list HomescriptArgs of script: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()
	args := make([]HomescriptArg, 0)
	for res.Next() {
		var currentArg HomescriptArg
		if err := res.Scan(
			&currentArg.Id,
			&currentArg.HomescriptId,
			&currentArg.Data.Prompt,
			&currentArg.Data.InputType,
			&currentArg.Data.Display,
		); err != nil {
			log.Error("Failed to list HomescriptArgs of script: scanning results failed: ", err.Error())
			return nil, err
		}
	}
	return args, nil
}

// Adds a new item to a Homescript's argument list
// Returns the newly created ID of the argument
func AddHomescriptArg(data HomescriptArg) (uint, error) {
	query, err := db.Prepare(`
	INSERT INTO
	homescriptArg(
		Id,
		HomescriptId,
		Prompt,
		InputType,
		Display
	)
	VALUES(DEFAULT, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to add Homescript argument: preparing query failed: ", err.Error())
		return 0, err
	}
	defer query.Close()
	res, err := query.Exec(
		data.HomescriptId,
		data.Data.Prompt,
		data.Data.InputType,
		data.Data.Display,
	)
	if err != nil {
		log.Error("Failed to add Homescript argument: executing query failed: ", err.Error())
		return 0, err
	}
	newId, err := res.LastInsertId()
	if err != nil {
		log.Error("Failed to add Homescript argument: retrieving last inserted ID failed: ", err.Error())
		return 0, err
	}
	return uint(newId), nil
}

// Modifies the data of a given Homescript argument
func ModifyHomescriptArg(id uint, newData HomescriptArgData) error {
	query, err := db.Prepare(`
	UPDATE homescriptArg
	SET
		Promt=?,
		InputType=?,
		Display=?
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to modify Homescript: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(
		newData.Prompt,
		newData.InputType,
		newData.Display,
		id,
	); err != nil {
		log.Error("Failed to modify Homescript: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns the data matching the id of an argument which is associated of a given user
func GetUserHomescriptArgById(id uint, username string) (data HomescriptArg, found bool, err error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		HomescriptId,
		Prompt,
		InputType,
		Display
	FROM homescriptArg
	JOIN homescript
		ON homescriptArg.homescriptId=homescript.Id
	WHERE Id=?
	AND
	homescript.Owner=?
	`)
	if err != nil {
		log.Error("Could not get Homescript argument by username and id: preparing query failed: ", err.Error())
		return HomescriptArg{}, false, err
	}
	defer query.Close()
	var currentArg HomescriptArg
	if err := query.QueryRow(id, username).Scan(
		&currentArg.Id,
		&currentArg.HomescriptId,
		&currentArg.Data.Prompt,
		&currentArg.Data.InputType,
		&currentArg.Data.Display,
	); err != nil {
		if err == sql.ErrNoRows {
			return HomescriptArg{}, false, nil
		}
		log.Error("Failed to get Homescript argument by username and id: scanning results failed: ", err.Error())
		return HomescriptArg{}, false, err
	}
	return currentArg, true, nil
}

// Deletes an arbitrary Homescript argument given its id
func DeleteHomescriptArg(id uint) error {
	query, err := db.Prepare(`
	DELETE FROM
	homescriptArg
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to delete Homescript argument: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(id); err != nil {
		log.Error("Failed to delete Homescript argument: executing query failed: ", err.Error())
		return err
	}
	return nil
}
