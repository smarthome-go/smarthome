package database

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
		Code TEXT,
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
