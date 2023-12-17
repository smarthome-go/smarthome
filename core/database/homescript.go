package database

import "fmt"

const HOMESCRIPT_ID_LEN = 150

type Homescript struct {
	Owner string         `json:"owner"`
	Data  HomescriptData `json:"data"`
}

type HomescriptData struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	QuickActionsEnabled bool   `json:"quickActionsEnabled"`
	IsWidget            bool   `json:"isWidget"`
	SchedulerEnabled    bool   `json:"schedulerEnabled"`
	Code                string `json:"code"`
	MDIcon              string `json:"mdIcon"`
	Workspace           string `json:"workspace"`
}

// Creates the table containing Homescript code and metadata
// If the database fails, this function returns an error
func createHomescriptTable() error {
	_, err := db.Exec(fmt.Sprintf(`
	CREATE TABLE
	IF NOT EXISTS
	homescript(
		Id					VARCHAR(%d),
		Owner				VARCHAR(20),
		Name				VARCHAR(30),
		Description			TEXT,
		QuickActionsEnabled BOOLEAN,
		SchedulerEnabled	BOOLEAN,
		isWidget			BOOLEAN,
		Code				TEXT,
		MDIcon				VARCHAR(100),
		Workspace			VARCHAR(50),

		PRIMARY KEY (Id, Owner),
		CONSTRAINT HomescriptOwner
		FOREIGN KEY (Owner)
		REFERENCES user(Username)
	)
	`, HOMESCRIPT_ID_LEN))
	if err != nil {
		log.Error("Failed to create Homescript Table: Executing query failed: ", err.Error())
		return err
	}

	// if _, err = query.Exec(HOMESCRIPT_ID_LEN); err != nil {
	// 	log.Error("Failed to create Homescript Table: Executing query failed: ", err.Error())
	// 	return err
	// }
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
		IsWidget,
		Code,
		MDIcon,
		Workspace
	)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to create new Homescript: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	// If the workspace is unset, set it to `default`
	if homescript.Data.Workspace == "" {
		homescript.Data.Workspace = "default"
	}
	// Create the Homescript
	if _, err = query.Exec(
		homescript.Data.Id,
		homescript.Owner,
		homescript.Data.Name,
		homescript.Data.Description,
		homescript.Data.QuickActionsEnabled,
		homescript.Data.SchedulerEnabled,
		homescript.Data.IsWidget,
		homescript.Data.Code,
		homescript.Data.MDIcon,
		homescript.Data.Workspace,
	); err != nil {
		log.Error("Failed to create new Homescript: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Modifies the metadata of a given homescript
// Does not check the validity of the homescript's id
func ModifyHomescriptById(id string, owner string, newData HomescriptData) error {
	// Check if the workspace is the default
	if newData.Workspace == "" {
		newData.Workspace = "default"
	}
	query, err := db.Prepare(`
	UPDATE homescript
	SET
		Name=?,
		Description=?,
		QuickActionsEnabled=?,
		SchedulerEnabled=?,
		IsWidget=?,
		Code=?,
		MDIcon=?,
		Workspace=?
	WHERE Id=? AND Owner=?
	`)
	if err != nil {
		log.Error("Failed to update Homescript: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	_, err = query.Exec(
		newData.Name,
		newData.Description,
		newData.QuickActionsEnabled,
		newData.SchedulerEnabled,
		newData.IsWidget,
		newData.Code,
		newData.MDIcon,
		newData.Workspace,
		id,
		owner,
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
		IsWidget,
		Code,
		MDIcon,
		Workspace
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
			&homescript.Data.IsWidget,
			&homescript.Data.Code,
			&homescript.Data.MDIcon,
			&homescript.Data.Workspace,
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
func ListAllHomescripts() ([]Homescript, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		Owner,
		Name,
		Description,
		QuickActionsEnabled,
		SchedulerEnabled,
		IsWidget,
		Code,
		MDIcon,
		Workspace
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
			&homescript.Data.IsWidget,
			&homescript.Data.Code,
			&homescript.Data.MDIcon,
			&homescript.Data.Workspace,
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
// Checks if a Homescript with an id exists in the database and belongs to a certain user
func DoesHomescriptExist(homescriptId string, owner string) (bool, error) {
	homescripts, err := ListAllHomescripts()
	if err != nil {
		log.Error("Failed to check existence of Homescript: ", err.Error())
		return false, err
	}
	for _, homescriptItem := range homescripts {
		if homescriptItem.Data.Id == homescriptId && homescriptItem.Owner == owner {
			return true, nil
		}
	}
	return false, nil
}

// Deletes a homescript by its Id, does not check if the user has access to the homescript
func DeleteHomescriptById(homescriptId string, owner string) error {
	if err := DeleteAllHomescriptArgsFromScript(homescriptId); err != nil {
		return err
	}
	query, err := db.Prepare(`
	DELETE FROM
	homescript
	WHERE Id=? AND Owner=?
	`)
	if err != nil {
		log.Error("Failed to delete Homescript by id: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(homescriptId, owner); err != nil {
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
		if err := DeleteHomescriptById(hms.Data.Id, username); err != nil {
			return err
		}
	}
	return nil
}
