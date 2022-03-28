package database

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
		Id, Name, Owner, Hour, Minute, HomescriptCode
	)
	VALUES(DEFAULT, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to create new schedule: preparing query failed: ", err.Error())
		return 0, err
	}
	res, err := query.Exec(
		schedule.Id,
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
