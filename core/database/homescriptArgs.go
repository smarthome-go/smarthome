package database

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
	HomescriptId string `json:"homescriptId"`
	Prompt       string `json:"prompt"`
}

// Used for creating the table which contains the arguments of Homescripts
// The `GuiDisplay` value is just used as a hint for user-interfaces for displaying a better selection of possible predefined values
func createHomescriptArgsTable() error {
	if _, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	homescriptArgs(
		HomescriptId VARCHAR(30) NOT NULL,
		Prompt TEXT NOT NULL,
		InputType ENUM(
			'string',
			'number',
			'boolean'
		),
		Display ENUM(
			'type_default',
			'string_switches',
			'boolean_yes_no',
			'boolean_on_off'
			'number_hour',
			'number_minute'
		)
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
		HomescriptId,
		Prompt,
		InputType,
		Display
	FROM homescriptArgs
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
		currentArg.Prompt = "TODO"
	}
	return args, nil
}
