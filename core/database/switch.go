package database

func createSwitchTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS
	switch(
		Id VARCHAR(2) PRIMARY KEY,
		Name VARCHAR(30)
	) 
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to create switch Table: Executing query failed: ", err.Error())
		return err
	}
	return nil
}
