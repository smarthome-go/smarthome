package database

import (
	"database/sql"
)

type Schedule struct {
	Id    uint         `json:"id"`
	Owner string       `json:"owner"`
	Data  ScheduleData `json:"data"`
}

type ScheduleData struct {
	Name               string                  `json:"name"`
	Hour               uint                    `json:"hour"`
	Minute             uint                    `json:"minute"`
	TargetMode         ScheduleTargetMode      `json:"targetMode"`         // Specifies which actions are taken when the schedule is executed
	HomescriptCode     string                  `json:"homescriptCode"`     // Is read when using the `code` mode of the schedule
	HomescriptTargetId string                  `json:"homescriptTargetId"` // Is required when using the `hms` mode of the schedule
	SwitchJobs         []ScheduleSwitchJobData `json:"switchJobs"`
}

// Specifies which action will be performed as a target
type ScheduleTargetMode string

const (
	ScheduleTargetModeCode     ScheduleTargetMode = "code"     // Will execute Homescript code as a target
	ScheduleTargetModeSwitches ScheduleTargetMode = "switches" // Will perform a series of power actions as a target
	ScheduleTargetModeHMS      ScheduleTargetMode = "hms"      // Will execute a Homescript by its id as a target
)

// Creates a new table containing the schedules for the normal scheduler jobs
func createScheduleTable() error {
	if _, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	schedule(
		Id INT AUTO_INCREMENT,
		Owner VARCHAR(20),
		Name VARCHAR(30),
		Hour INT,
		Minute INT,
		TargetMode ENUM (
			'switches',
			'code',
			'hms'
		),
		HomescriptCode TEXT,
		HomescriptTargetId VARCHAR(30),
		PRIMARY KEY (Id),
		FOREIGN KEY (Owner)
		REFERENCES user(Username)
	)
	`); err != nil {
		log.Error("Failed to create schedule table: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Creates a new schedule which represents a job of the scheduler
func CreateNewSchedule(
	owner string,
	data ScheduleData,
) (uint, error) {
	query, err := db.Prepare(`
	INSERT INTO
	schedule(
		Id,
		Owner,
		Name,
		Hour,
		Minute,
		TargetMode,
		HomescriptCode,
		HomescriptTargetId
	)
	VALUES(DEFAULT, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to create new schedule: preparing query failed: ", err.Error())
		return 0, err
	}
	defer query.Close()
	res, err := query.Exec(
		owner,
		data.Name,
		data.Hour,
		data.Minute,
		data.TargetMode,
		data.HomescriptCode,
		data.HomescriptTargetId,
	)
	if err != nil {
		log.Error("Failed to create new schedule: executing query failed: ", err.Error())
		return 0, err
	}
	newId, err := res.LastInsertId()
	if err != nil {
		log.Error("Failed to create new schedule: retrieving last inserted id failed: ", err.Error())
		return 0, err
	}
	// Create the schedule's switch jobs
	for _, switchJob := range data.SwitchJobs {
		if _, err := CreateNewScheduleSwitch(
			uint(newId),
			switchJob.SwitchId,
			switchJob.PowerOn,
		); err != nil {
			log.Error("Failed to create new schedule: could not create switch job: ", err.Error())
			return 0, err
		}
	}
	return uint(newId), nil
}

// Returns a schedule struct which matches the given id
// If the id does not match a struct, a `false`` is returned
func GetScheduleById(id uint) (Schedule, bool, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		Name,
		Owner,
		Hour,
		Minute,
		TargetMode,
		HomescriptCode,
		HomescriptTargetId,
	FROM schedule
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to get schedule by id: preparing query failed: ", err.Error())
		return Schedule{}, false, err
	}
	defer query.Close()
	var schedule Schedule
	if err := query.QueryRow(id).Scan(
		&schedule.Id,
		&schedule.Data.Name,
		&schedule.Owner,
		&schedule.Data.Hour,
		&schedule.Data.Minute,
		&schedule.Data.TargetMode,
		&schedule.Data.HomescriptCode,
		&schedule.Data.HomescriptTargetId,
	); err != nil {
		if err == sql.ErrNoRows {
			return Schedule{}, false, nil
		}
		log.Error("Failed to get schedule by id: executing query failed: ", err.Error())
		return Schedule{}, false, err
	}

	// Obtain this schedule's switch jobs
	switches, err := ListSwitchesOfSchedule(schedule.Id)
	if err != nil {
		return Schedule{}, false, err
	}
	schedule.Data.SwitchJobs = switches

	return schedule, true, nil
}

// Returns a list containing schedules of a given user
func GetUserSchedules(username string) ([]Schedule, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		Name,
		Owner,
		Hour,
		Minute,
		TargetMode,
		HomescriptCode,
		HomescriptTargetId
	FROM schedule
	WHERE Owner=?
	`)
	if err != nil {
		log.Error("Failed to list user schedules: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query(username)
	if err != nil {
		log.Error("Failed to list user schedules: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()

	// Obtain schedule switches
	switches, err := ListUserScheduleSwitches(username)
	if err != nil {
		return nil, err
	}

	schedules := make([]Schedule, 0)
	for res.Next() {
		var schedule Schedule
		if err := res.Scan(
			&schedule.Id,
			&schedule.Data.Name,
			&schedule.Owner,
			&schedule.Data.Hour,
			&schedule.Data.Minute,
			&schedule.Data.TargetMode,
			&schedule.Data.HomescriptCode,
			&schedule.Data.HomescriptTargetId,
		); err != nil {
			log.Error("Failed to list user schedules: scanning results of query failed: ", err.Error())
			return nil, err
		}

		// Append the schedule's switches to the data
		schedule.Data.SwitchJobs = make([]ScheduleSwitchJobData, 0)
		for _, swItem := range switches {
			if swItem.ScheduleId == schedule.Id {
				schedule.Data.SwitchJobs = append(schedule.Data.SwitchJobs, swItem.Data)
			}
		}

		// Append the row to the list
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}

// Returns a list of schedules of all users, used for activating schedules at the start of the server
func GetSchedules() ([]Schedule, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		Name,
		Owner,
		Hour,
		Minute,
		TargetMode,
		HomescriptCode,
		HomescriptTargetId
	FROM schedule
	`)
	if err != nil {
		log.Error("Failed to list schedules: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query()
	if err != nil {
		log.Error("Failed to list schedules: executing query failed: ", err.Error())
		return nil, err
	}

	// Obtain all schedule switch jobs
	switches, err := ListAllScheduleSwitches()
	if err != nil {
		return nil, err
	}

	defer res.Close()
	schedules := make([]Schedule, 0)
	for res.Next() {
		var schedule Schedule
		if err := res.Scan(
			&schedule.Id,
			&schedule.Data.Name,
			&schedule.Owner,
			&schedule.Data.Hour,
			&schedule.Data.Minute,
			&schedule.Data.TargetMode,
			&schedule.Data.HomescriptCode,
			&schedule.Data.HomescriptTargetId,
		); err != nil {
			log.Error("Failed to list schedules: scanning results of query failed: ", err.Error())
			return nil, err
		}
		// Append all needed switch jobs to this row
		for _, switchJob := range switches {
			schedule.Data.SwitchJobs = append(schedule.Data.SwitchJobs, switchJob.Data)
		}

		// Apppend the row to the results
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}

// Modifies the metadata of a given schedule
// Does not validate the provided metadata
func ModifySchedule(id uint, newData ScheduleData) error {
	query, err := db.Prepare(`
	UPDATE schedule
	SET
		Name=?,
		Hour=?,
		Minute=?,
		TargetMode=?,
		HomescriptCode=?,
		HomescriptTargetId=?
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to modify schedule: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(
		newData.Name,
		newData.Hour,
		newData.Minute,
		newData.TargetMode,
		newData.HomescriptCode,
		newData.HomescriptTargetId,
		id,
	); err != nil {
		log.Error("Failed to modify schedule: executing query failed: ", err.Error())
		return err
	}
	// Perform switch diff operations
	oldSwitches, err := ListSwitchesOfSchedule(id)
	if err != nil {
		return err
	}
	add, del := getSwitchDiff(
		oldSwitches,
		newData.SwitchJobs,
	)

	// Remove all unused switches
	for _, swDel := range del {
		if err := DeleteSwitchFromSchedule(
			swDel.SwitchId,
			id,
		); err != nil {
			return err
		}
	}
	// Add all missing switches
	for _, swAdd := range add {
		if _, err := CreateNewScheduleSwitch(
			id,
			swAdd.SwitchId,
			swAdd.PowerOn,
		); err != nil {
			return err
		}
	}
	return nil
}

// Compares two slices which contain schedule switch jobs
// Outputs two slices which determine which actions have to be taken to transform the old state into the new state
// This function is used in schedule modification
func getSwitchDiff(
	oldSwitches []ScheduleSwitchJobData,
	newSwitches []ScheduleSwitchJobData,
) (
	add []ScheduleSwitchJobData,
	del []ScheduleSwitchJobData,
) {
	// Determine deletions
	for _, swOld := range oldSwitches {
		exists := false
		for _, swNew := range newSwitches {
			if swNew.SwitchId == swOld.SwitchId && swNew.PowerOn == swOld.PowerOn {
				exists = true
				break
			}
		}
		if !exists {
			del = append(del, swOld)
		}
	}
	// Determine addition
	for _, swNew := range newSwitches {
		exists := false
		for _, swOld := range oldSwitches {
			if swOld.SwitchId == swNew.SwitchId && swOld.PowerOn == swNew.PowerOn {
				exists = true
				break
			}
		}
		if !exists {
			add = append(add, swNew)
		}
	}

	return add, del
}

// Deletes a schedule item given its Id
// Deletes all switch jobs first
// Does not validate the validity of the provided Id
func DeleteScheduleById(id uint) error {
	// Delete all switch jobs first
	if err := DeleteAllSwitchesFromSchedule(id); err != nil {
		return err
	}
	// Delete the actual schedule
	query, err := db.Prepare(`
	DELETE FROM
	schedule
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to delete schedule by id: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(id); err != nil {
		log.Error("Failed to delete schedule by id: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes all schedules from a given user
func DeleteAllSchedulesFromUser(username string) error {
	schedules, err := GetUserSchedules(username)
	if err != nil {
		return err
	}
	for _, schedule := range schedules {
		if err := DeleteScheduleById(schedule.Id); err != nil {
			return err
		}
	}
	return nil
}
