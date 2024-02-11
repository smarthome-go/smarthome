package database

type ScheduleDeviceJob struct {
	ScheduleId uint                  `json:"scheduleId"`
	Data       ScheduleDeviceJobData `json:"data"`
}

type ScheduleDeviceJobData struct {
	DeviceId string `json:"deviceId"`
	PowerOn  bool   `json:"powerOn"`
}

// Creates the table containing the device jobs of a schedule which uses the `device` mode
func createSchedulerDeviceJobTable() error {
	// TODO: missing foreign key for device
	if _, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	scheduleDeviceJob(
		ScheduleId INT,
		DeviceId VARCHAR(20),
		Power BOOLEAN,
		FOREIGN KEY (ScheduleId)
		REFERENCES schedule(Id)
	)
	`); err != nil {
		log.Error("Failed to create schedule device job table: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns a list of all schedule device jobs
func ListAllScheduleDeviceJobs() ([]ScheduleDeviceJob, error) {
	query, err := db.Prepare(`
	SELECT
		ScheduleId,
		DeviceId,
		Power
	FROM scheduleDeviceJob
	`)
	if err != nil {
		log.Error("Failed to list all schedule device jobs: preparing query failed: ", err.Error())
		return nil, err
	}
	res, err := query.Query()
	if err != nil {
		log.Error("Failed to list all schedule device jobs: executing query failed: ", err.Error())
		return nil, err
	}
	deviceJobs := make([]ScheduleDeviceJob, 0)
	for res.Next() {
		var jobRow ScheduleDeviceJob
		if err := res.Scan(
			&jobRow.ScheduleId,
			&jobRow.Data.DeviceId,
			&jobRow.Data.PowerOn,
		); err != nil {
			log.Error("Failed to list all schedule device jobs: scanning query result row failed: ", err.Error())
			return nil, err
		}
		deviceJobs = append(deviceJobs, jobRow)
	}
	return deviceJobs, nil
}

// Returns a list of all schedule device jobs owned by a given user
func ListUserScheduleDeviceJobs(username string) ([]ScheduleDeviceJob, error) {
	query, err := db.Prepare(`
	SELECT
		ScheduleId,
		DeviceId,
		Power
	FROM scheduleDeviceJob
	JOIN schedule
		ON schedule.Id = scheduleDeviceJob.ScheduleId
	WHERE schedule.Owner=?
	`)
	if err != nil {
		log.Error("Failed to list user schedule device jobs: preparing query failed: ", err.Error())
		return nil, err
	}
	res, err := query.Query(username)
	if err != nil {
		log.Error("Failed to list user schedule device jobs: executing query failed: ", err.Error())
		return nil, err
	}
	deviceJobs := make([]ScheduleDeviceJob, 0)
	for res.Next() {
		var deviceRow ScheduleDeviceJob
		if err := res.Scan(
			&deviceRow.ScheduleId,
			&deviceRow.Data.DeviceId,
			&deviceRow.Data.PowerOn,
		); err != nil {
			log.Error("Failed to list user schedule device jobs: scanning query result row failed: ", err.Error())
			return nil, err
		}
		deviceJobs = append(deviceJobs, deviceRow)
	}
	return deviceJobs, nil
}

// Returns a list of schedule device jobs belonging to an arbitrary schedule
func ListDeviceJobsOfSchedule(scheduleId uint) ([]ScheduleDeviceJobData, error) {
	query, err := db.Prepare(`
	SELECT
		DeviceId,
		Power
	FROM scheduleDeviceJob
	WHERE ScheduleId=?
	`)
	if err != nil {
		log.Error("Failed to list device jobs of schedule: preparing query failed: ", err.Error())
		return nil, err
	}
	res, err := query.Query(scheduleId)
	if err != nil {
		log.Error("Failed to list device jobs of schedule: executing query failed: ", err.Error())
		return nil, err
	}
	deviceJobs := make([]ScheduleDeviceJobData, 0)
	for res.Next() {
		var jobRow ScheduleDeviceJobData
		if err := res.Scan(
			&jobRow.DeviceId,
			&jobRow.PowerOn,
		); err != nil {
			log.Error("Failed to list device jobs of schedule: scanning query results failed: ", err.Error())
			return nil, err
		}
		deviceJobs = append(deviceJobs, jobRow)
	}
	return deviceJobs, nil
}

// Creates a new device job for a corresponding schedule
// All data must be validated beforehand
func CreateNewScheduleDeviceJob(
	scheduleId uint,
	deviceId string,
	powerOn bool,
) (uint, error) {
	query, err := db.Prepare(`
	INSERT INTO
	scheduleDeviceJob(
		ScheduleId,
		DeviceId,
		Power
	)
	VALUES(?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to create new schedule device job: preparing query failed: ", err.Error())
		return 0, err
	}
	res, err := query.Exec(
		scheduleId,
		deviceId,
		powerOn,
	)
	if err != nil {
		log.Error("Failed to create new schedule device job: executing query failed: ", err.Error())
		return 0, err
	}
	newId, err := res.LastInsertId()
	if err != nil {
		log.Error("Failed to create new schedule device job: failed to obtain inserted Id: ", err.Error())
		return 0, err
	}
	return uint(newId), nil
}

// Deletes all device job items from a given schedule
// Used when a schedule is deleted.
func DeleteAllDeviceJobsFromSchedule(scheduleId uint) error {
	query, err := db.Prepare(`
	DELETE FROM
	scheduleDeviceJob
	WHERE ScheduleId=?
	`)
	if err != nil {
		log.Error("Failed to delete all device jobs from schedule: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Exec(scheduleId); err != nil {
		log.Error("Failed to delete all device jobs from schedule: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes an existent device job from a given schedule
// All data has to be validated beforehand
func DeleteDeviceJobFromSchedule(
	deviceId string,
	scheduleId uint,
) error {
	query, err := db.Prepare(`
	DELETE FROM
	scheduleDeviceJob
	WHERE DeviceId=?
	AND ScheduleId=?
	`)
	if err != nil {
		log.Error(`Failed to delete device job from schedule: preparing query failed: `, err.Error())
		return err
	}
	if _, err := query.Exec(
		deviceId,
		scheduleId,
	); err != nil {
		log.Error(`Failed to delete device job from schedule: executing query failed: `, err.Error())
		return err
	}
	return nil
}
