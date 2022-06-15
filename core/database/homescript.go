package database

type Homescript struct {
	Owner string         `json:"owner"`
	Data  HomescriptData `json:"data"`
}

type HomescriptData struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	QuickActionsEnabled bool   `json:"quickActionsEnabled"`
	SchedulerEnabled    bool   `json:"schedulerEnabled"`
	Code                string `json:"code"`
	MDIcon              string `json:"mdIcon"`
}

// Creates the table containing Homescript code and metadata
// If the database fails, this function returns an error
func createHomescriptTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	homescript(
		Id VARCHAR(30) PRIMARY KEY,
		Owner VARCHAR(20),
		Name VARCHAR(30),
		Description TEXT,
		QuickActionsEnabled BOOLEAN,
		SchedulerEnabled BOOLEAN,
		Code TEXT,
		MDIcon VARCHAR(100),
		CONSTRAINT HomescriptOwner
		FOREIGN KEY (Owner)
		REFERENCES user(Username)
	) 
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to create Homescript Table: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Creates a new homescript entry
func CreateNewHomescript(homescript Homescript) error {
	query, err := db.Prepare(`
	INSERT INTO
	homescript(
		Id, 
		Owner,
		Name,
		Description,
		QuickActionsEnabled,
		SchedulerEnabled,
		Code,
		MDIcon
	)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to create new Homescript: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err = query.Exec(
		homescript.Data.Id,
		homescript.Owner,
		homescript.Data.Name,
		homescript.Data.Description,
		homescript.Data.QuickActionsEnabled,
		homescript.Data.SchedulerEnabled,
		homescript.Data.Code,
		homescript.Data.MDIcon,
	); err != nil {
		log.Error("Failed to create new Homescript: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Modifies the metadata of a given homescript
// Does not check the validity of the homescript's id
func ModifyHomescriptById(id string, homescript HomescriptData) error {
	query, err := db.Prepare(`
	UPDATE homescript
	SET 
		Name=?,
		Description=?,
		QuickActionsEnabled=?,
		SchedulerEnabled=?,
		Code=?,
		MDIcon=?
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to update Homescript: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	_, err = query.Exec(
		homescript.Name,
		homescript.Description,
		homescript.QuickActionsEnabled,
		homescript.SchedulerEnabled,
		homescript.Code,
		homescript.MDIcon,
		id,
	)
	if err != nil {
		log.Error("Failed to update Homescript: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns a list of homescripts owned by a given user
func ListHomescriptOfUser(username string) ([]Homescript, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		Owner,
		Name,
		Description,
		QuickActionsEnabled,
		SchedulerEnabled,
		Code,
		MDIcon
	FROM homescript
	WHERE Owner=?
	`)
	if err != nil {
		log.Error("Failed to list Homescript of user: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query(username)
	if err != nil {
		log.Error("Failed to list homescript of user: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()
	var homescriptList []Homescript = make([]Homescript, 0)
	for res.Next() {
		var homescript Homescript
		err := res.Scan(
			&homescript.Data.Id,
			&homescript.Owner,
			&homescript.Data.Name,
			&homescript.Data.Description,
			&homescript.Data.QuickActionsEnabled,
			&homescript.Data.SchedulerEnabled,
			&homescript.Data.Code,
			&homescript.Data.MDIcon,
		)
		if err != nil {
			log.Error("Failed to list Homescript of user: scanning results failed: ", err.Error())
			return nil, err
		}
		homescriptList = append(homescriptList, homescript)
	}
	return homescriptList, nil
}

// Lists all Homescript files in the database
func ListHomescriptFiles() ([]Homescript, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		Owner,
		Name,
		Description,
		QuickActionsEnabled,
		SchedulerEnabled,
		Code,
		MDIcon
	FROM homescript
	`)
	if err != nil {
		log.Error("Failed to list Homescript: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query()
	if err != nil {
		log.Error("Failed to list Homescript: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()
	var homescriptList []Homescript = make([]Homescript, 0)
	for res.Next() {
		var homescript Homescript
		err := res.Scan(
			&homescript.Data.Id,
			&homescript.Owner,
			&homescript.Data.Name,
			&homescript.Data.Description,
			&homescript.Data.QuickActionsEnabled,
			&homescript.Data.SchedulerEnabled,
			&homescript.Data.Code,
			&homescript.Data.MDIcon,
		)
		if err != nil {
			log.Error("Failed to list Homescript: scanning results failed: ", err.Error())
			return nil, err
		}
		homescriptList = append(homescriptList, homescript)
	}
	return homescriptList, nil
}

// Returns a Homescript given its id
// Returns Homescript, has been found, error
// TODO: replace with query row
func GetUserHomescriptById(homescriptId string, username string) (Homescript, bool, error) {
	homescripts, err := ListHomescriptOfUser(username)
	if err != nil {
		log.Error("Failed to get Homescript by id: ", err.Error())
		return Homescript{}, false, err
	}
	for _, homescriptItem := range homescripts {
		if homescriptItem.Data.Id == homescriptId {
			return homescriptItem, true, nil
		}
	}
	return Homescript{}, false, nil
}

// TODO: remove and intigrate into get user homescript
// Checks if a Homescript with an id exists in the database
func DoesHomescriptExist(homescriptId string) (bool, error) {
	homescripts, err := ListHomescriptFiles()
	if err != nil {
		log.Error("Failed to check existence of Homescript: ", err.Error())
		return false, err
	}
	for _, homescriptItem := range homescripts {
		if homescriptItem.Data.Id == homescriptId {
			return true, nil
		}
	}
	return false, nil
}

// Deletes a homescript by its Id, does not check if the user has access to the homescript
func DeleteHomescriptById(homescriptId string) error {
	if err := DeleteAllHomescriptArgsFromScript(homescriptId); err != nil {
		return err
	}
	query, err := db.Prepare(`
	DELETE FROM
	homescript
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to delete Homescript by id: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(homescriptId); err != nil {
		log.Error("Failed to delete Homescript by id: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes all Homescripts of a given user
// Uses the `DeleteHomescriptById` function
func DeleteAllHomescriptsOfUser(username string) error {
	homescripts, err := ListHomescriptOfUser(username)
	if err != nil {
		return err
	}
	for _, hms := range homescripts {
		if err := DeleteAllHomescriptArgsFromScript(hms.Data.Id); err != nil {
			return err
		}
		if err := DeleteHomescriptById(hms.Data.Id); err != nil {
			return err
		}
	}
	return nil
}
