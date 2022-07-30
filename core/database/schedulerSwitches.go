package database

type ScheduleSwitchJob struct {
	ScheduleId uint                  `json:"scheduleId"`
	Data       ScheduleSwitchJobData `json:"data"`
}

type ScheduleSwitchJobData struct {
	SwitchId string `json:"switchId"`
	PowerOn  bool   `json:"powerOn"`
}

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
		REFERENCES schedule(Id)
	)
	`); err != nil {
		log.Error("Failed to create schedule switches table: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns a list of all schedule switches
func ListAllScheduleSwitches() ([]ScheduleSwitchJob, error) {
	query, err := db.Prepare(`
	SELECT
		ScheduleId,
		SwitchId,
		Power
	FROM scheduleSwitches
	`)
	if err != nil {
		log.Error("Failed to list all schedule switches: preparing query failed: ", err.Error())
		return nil, err
	}
	res, err := query.Query()
	if err != nil {
		log.Error("Failed to list all schedule switches: executing query failed: ", err.Error())
		return nil, err
	}
	switches := make([]ScheduleSwitchJob, 0)
	for res.Next() {
		var switchRow ScheduleSwitchJob
		if err := res.Scan(
			&switchRow.ScheduleId,
			&switchRow.Data.SwitchId,
			&switchRow.Data.PowerOn,
		); err != nil {
			log.Error("Failed to list all schedule switches: scanning query result row failed: ", err.Error())
			return nil, err
		}
		switches = append(switches, switchRow)
	}
	return switches, nil
}

// Returns a list of all schedule switch jobs owned by a given user
func ListUserScheduleSwitches(username string) ([]ScheduleSwitchJob, error) {
	query, err := db.Prepare(`
	SELECT
		ScheduleId,
		SwitchId,
		Power
	FROM scheduleSwitches
	JOIN schedule
		ON schedule.Id = scheduleSwitches.ScheduleId
	WHERE schedule.Owner=?
	`)
	if err != nil {
		log.Error("Failed to list user schedule switches: preparing query failed: ", err.Error())
		return nil, err
	}
	res, err := query.Query(username)
	if err != nil {
		log.Error("Failed to list user schedule switches: executing query failed: ", err.Error())
		return nil, err
	}
	switches := make([]ScheduleSwitchJob, 0)
	for res.Next() {
		var switchRow ScheduleSwitchJob
		if err := res.Scan(
			&switchRow.ScheduleId,
			&switchRow.Data.SwitchId,
			&switchRow.Data.PowerOn,
		); err != nil {
			log.Error("Failed to list user schedule switches: scanning query result row failed: ", err.Error())
			return nil, err
		}
		switches = append(switches, switchRow)
	}
	return switches, nil
}

// Returns a list of schedule switch jobs belonging to an arbitrary schedule
func ListSwitchesOfSchedule(scheduleId uint) ([]ScheduleSwitchJobData, error) {
	query, err := db.Prepare(`
	SELECT
		SwitchId,
		Power
	FROM scheduleSwitches
	WHERE ScheduleId=?
	`)
	if err != nil {
		log.Error("Failed to list switches of schedule: preparing query failed: ", err.Error())
		return nil, err
	}
	res, err := query.Query(scheduleId)
	if err != nil {
		log.Error("Failed to list switches of schedule: executing query failed: ", err.Error())
		return nil, err
	}
	switches := make([]ScheduleSwitchJobData, 0)
	for res.Next() {
		var switchRow ScheduleSwitchJobData
		if err := res.Scan(
			&switchRow.SwitchId,
			&switchRow.PowerOn,
		); err != nil {
			log.Error("Failed to list switches of schedule: scanning query results failed: ", err.Error())
			return nil, err
		}
		switches = append(switches, switchRow)
	}
	return switches, nil
}

// Creates a new power job for a corresponding schedule
// All data must be validated beforehand
func CreateNewScheduleSwitch(
	scheduleId uint,
	switchId string,
	powerOn bool,
) (uint, error) {
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
		return 0, err
	}
	res, err := query.Exec(
		scheduleId,
		switchId,
		powerOn,
	)
	if err != nil {
		log.Error("Failed to create new schedule switch: executing query failed: ", err.Error())
		return 0, err
	}
	newId, err := res.LastInsertId()
	if err != nil {
		log.Error("Failed to create new schedule switch: failed to obtain inserted Id: ", err.Error())
		return 0, err
	}
	return uint(newId), nil
}

// Deletes all switch items from a given schedule
// Used when a schedule is deleted
func DeleteAllSwitchesFromSchedule(scheduleId uint) error {
	query, err := db.Prepare(`
	DELETE FROM
	scheduleSwitches
	WHERE ScheduleId=?
	`)
	if err != nil {
		log.Error("Failed to delete all switch jobs from schedule: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Exec(scheduleId); err != nil {
		log.Error("Failed to delete all switch jobs from schedule: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes an existent switch job from a given schedule
// All data has to be validated beforehand
func DeleteSwitchFromSchedule(
	switchId string,
	scheduleId uint,
) error {
	query, err := db.Prepare(`
	DELETE FROM
	scheduleSwitches
	WHERE SwitchId=?
	AND ScheduleId=?
	`)
	if err != nil {
		log.Error(`Failed to delete switch job from schedule: preparing query failed: `, err.Error())
		return err
	}
	if _, err := query.Exec(
		switchId,
		scheduleId,
	); err != nil {
		log.Error(`Failed to delete switch job from schedule: executing query failed: `, err.Error())
		return err
	}
	return nil
}
