package database

import "database/sql"

type Schedule struct {
	Id             uint   `json:"id"`
	Name           string `json:"name"`
	Owner          string `json:"owner"`
	Hour           uint   `json:"hour"`
	Minute         uint   `json:"minute"`
	HomescriptCode string `json:"homescriptCode"` // Will be executed if the scheduler runs the job
}

type ScheduleWithoudIdAndUsername struct {
	Name           string `json:"name"`
	Hour           uint   `json:"hour"`
	Minute         uint   `json:"minute"`
	HomescriptCode string `json:"homescriptCode"`
}

// Creates a new table containing the schedules for the normal scheduler jobs
func createScheduleTable() error {
	if _, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	schedule(
		Id INT AUTO_INCREMENT,
		Name VARCHAR(30),
		Owner VARCHAR(20),
		Hour INT,
		Minute INT,
		HomescriptCode TEXT,
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
func CreateNewSchedule(schedule Schedule) (uint, error) {
	query, err := db.Prepare(`
	INSERT INTO
	schedule(
		Id,
		Name,
		Owner,
		Hour,
		Minute,
		HomescriptCode
	)
	VALUES(DEFAULT, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to create new schedule: preparing query failed: ", err.Error())
		return 0, err
	}
	defer query.Close()
	res, err := query.Exec(
		schedule.Name,
		schedule.Owner,
		schedule.Hour,
		schedule.Minute,
		schedule.HomescriptCode,
	)
	if err != nil {
		log.Error("Failed to create new scheduler: executing query failed: ", err.Error())
		return 0, err
	}
	newId, err := res.LastInsertId()
	if err != nil {
		log.Error("Failed to create new scheduler: retrieving last inserted id failed: ", err.Error())
		return 0, err
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
		HomescriptCode
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
		&schedule.Name,
		&schedule.Owner,
		&schedule.Hour,
		&schedule.Minute,
		&schedule.HomescriptCode,
	); err != nil {
		if err == sql.ErrNoRows {
			return Schedule{}, false, nil
		}
		log.Error("Failed to get schedule by id: executing query failed: ", err.Error())
		return Schedule{}, false, err
	}
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
		HomescriptCode
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
	schedules := make([]Schedule, 0)
	for res.Next() {
		var schedule Schedule
		if err := res.Scan(
			&schedule.Id,
			&schedule.Name,
			&schedule.Owner,
			&schedule.Hour,
			&schedule.Minute,
			&schedule.HomescriptCode,
		); err != nil {
			log.Error("Failed to list user schedules: scanning results of query failed: ", err.Error())
			return nil, err
		}
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
		HomescriptCode
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
	defer res.Close()
	schedules := make([]Schedule, 0)
	for res.Next() {
		var schedule Schedule
		if err := res.Scan(
			&schedule.Id,
			&schedule.Name,
			&schedule.Owner,
			&schedule.Hour,
			&schedule.Minute,
			&schedule.HomescriptCode,
		); err != nil {
			log.Error("Failed to list schedules: scanning results of query failed: ", err.Error())
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}

// Modifies the metadata of a given schedule
// Does not validate the provided metadata
func ModifySchedule(id uint, newItem ScheduleWithoudIdAndUsername) error {
	query, err := db.Prepare(`
	UPDATE schedule
	SET
		Name=?,
		Hour=?,
		Minute=?,
		HomescriptCode=?
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to modify schedule: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(
		newItem.Name,
		newItem.Hour,
		newItem.Minute,
		newItem.HomescriptCode,
		id,
	); err != nil {
		log.Error("Failed to modify schedule: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes a schedule item given its Id
// Does not validate the validity of the provided Id
func DeleteScheduleById(id uint) error {
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
	query, err := db.Prepare(`
	DElETE FROM
	schedule
	WHERE Owner=?
	`)
	if err != nil {
		log.Error("Failed to delete all schedules of user: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(username); err != nil {
		log.Error("Failed to delete all schedules of user: executing query failed: ", err.Error())
		return err
	}
	return nil
}
