package database

type HomescriptArg struct {
	HomescriptId string `json:"homescriptId"`
	Prompt       string `json:"prompt"`
}

// Used for creating the table which contains the arguments of Homescripts
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
		GuiDisplay ENUM(
			'type_default',
			'string_switches',
			'boolean_yes_no',
			'number_hour',
			'number_minute'
		)
	)
	`); err != nil {
		return err
	}
	return nil
}
