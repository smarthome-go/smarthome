package database

// Creates the table containing the switch jobs of a schedule which uses the `switches` mode
func createSchedulerSwitchesTable() error {
	if _, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	scheduleSwitches(
		ScheduleId INT, 
		SwitchId VARCHAR(20),
		Power BOOLEAN,
		FOREIGN KEY (ScheduleId)
		REFERENCES schedule(Id),
		FOREIGN KEY (SwitchId)
		REFERENCES switch(Id)
	)
	`); err != nil {
		log.Error("Failed to create schedule switches table: executing query failed: ", err.Error())
		return err
	}
	return nil
}

func CreateNewScheduleSwitch(
	scheduleId uint,
	switchId string,
	powerOn bool,
) error {
	query, err := db.Prepare(`
	INSERT INTO
	scheduleSwitches(
		ScheduleId,
		SwitchId,
		Power
	)
	VALUES(?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to create new schedule switch: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Query(
		scheduleId,
		switchId,
		powerOn,
	); err != nil {
		log.Error("Failed to create new schedule switch: executing query failed: ", err.Error())
		return err
	}
	return nil
}
