package database

import "database/sql"

type HmsArgInputType string

// Datatypes which a Homescript argument can use
// Type conversion is handled by the target Homescript
// These types act as a hint for the user and
var (
	String  HmsArgInputType = "string"
	Number  HmsArgInputType = "number"
	Boolean HmsArgInputType = "boolean"
)

type HmsArgDisplay string

var (
	TypeDefault    HmsArgDisplay = "type_default"    // Uses a normal input field matching the specified data type
	StringSwitches HmsArgDisplay = "string_switches" // Shows a list of switches from which the user can select one as a string
	BooleanYesNo   HmsArgDisplay = "boolean_yes_no"  // Uses `yes` and `no` as substitutes for true and false
	BooleanOnOff   HmsArgDisplay = "boolean_on_off"  // Uses `on` and `off` as substitutes for true and false
	NumberHour     HmsArgDisplay = "number_hour"     // Displays a hour picker (0 <= h <= 24)
	NumberMinute   HmsArgDisplay = "number_minute"   // Displays a minute picker (0 <= m <= 60)
)

type HomescriptArg struct {
	Id   uint              `json:"id"`   // The Id is automatically generated
	Data HomescriptArgData `json:"data"` // The main data of the argument
}

type HomescriptArgData struct {
	ArgKey       string          `json:"argKey"`       // The unique key of the argument
	HomescriptId string          `json:"homescriptId"` // The Homescript to which the argument belongs to
	Prompt       string          `json:"prompt"`       // What the user will be prompted
	InputType    HmsArgInputType `json:"inputType"`    // Which data type is expected
	Display      HmsArgDisplay   `json:"display"`      // How the prompt will look like
}

// Used for creating the table which contains the arguments of Homescripts
// The `GUIDisplay` value is just used as a hint for user-interfaces for displaying a better selection of possible predefined values
func createHomescriptArgTable() error {
	if _, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	homescriptArg(
		Id INT AUTO_INCREMENT,
		ArgKey VARCHAR(100),
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

// Returns the data matching the id of an argument which is associated of a given user
func GetUserHomescriptArgById(id uint, username string) (data HomescriptArg, found bool, err error) {
	query, err := db.Prepare(`
	SELECT
		homescriptArg.Id,
		ArgKey,
		HomescriptId,
		Prompt,
		InputType,
		Display
	FROM homescriptArg
	JOIN homescript
		ON homescriptArg.homescriptId=homescript.Id
	WHERE homescriptArg.Id=?
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
		&currentArg.Data.ArgKey,
		&currentArg.Data.HomescriptId,
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

// Returns all HomescriptArgs of a given user as a slice
func ListAllHomescriptArgsOfUser(username string) ([]HomescriptArg, error) {
	query, err := db.Prepare(`
	SELECT
		homescriptArg.Id,
		ArgKey,
		HomescriptId,
		Prompt,
		InputType,
		Display
	FROM homescriptArg
	JOIN homescript
		ON homescriptArg.HomescriptId=homescript.Id
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
			&currentArg.Data.ArgKey,
			&currentArg.Data.HomescriptId,
			&currentArg.Data.Prompt,
			&currentArg.Data.InputType,
			&currentArg.Data.Display,
		); err != nil {
			log.Error("Failed to list HomescriptArgs of user: scanning results failed: ", err.Error())
			return nil, err
		}
		args = append(args, currentArg)
	}
	return args, nil
}

// Returns the arguments of a given Homescript as a slice
func ListArgsOfHomescript(homescriptId string) ([]HomescriptArg, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		ArgKey,
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
			&currentArg.Data.ArgKey,
			&currentArg.Data.HomescriptId,
			&currentArg.Data.Prompt,
			&currentArg.Data.InputType,
			&currentArg.Data.Display,
		); err != nil {
			log.Error("Failed to list HomescriptArgs of script: scanning results failed: ", err.Error())
			return nil, err
		}
		args = append(args, currentArg)
	}
	return args, nil
}

// Adds a new item to a Homescript's argument list
// Returns the newly created ID of the argument
func AddHomescriptArg(data HomescriptArgData) (uint, error) {
	query, err := db.Prepare(`
	INSERT INTO
	homescriptArg(
		Id,
		ArgKey,
		HomescriptId,
		Prompt,
		InputType,
		Display
	)
	VALUES(DEFAULT, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to add Homescript argument: preparing query failed: ", err.Error())
		return 0, err
	}
	defer query.Close()
	res, err := query.Exec(
		data.ArgKey,
		data.HomescriptId,
		data.Prompt,
		data.InputType,
		data.Display,
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
		ArgKey=?,
		Prompt=?,
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
		newData.ArgKey,
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

// Deletes all arguments referencing a given Homescript
// Used before deleting a Homescript from the database
func DeleteAllHomescriptArgsFromScript(homescriptId string) error {
	query, err := db.Prepare(`
	DELETE FROM
	homescriptArg
	WHERE homescriptId=?
	`)
	if err != nil {
		log.Error("Failed to delete all arguments from Homescript: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(homescriptId); err != nil {
		log.Error("Failed to delete all arguments from Homescript: executing query failed: ", err.Error())
		return err
	}
	return nil
}
