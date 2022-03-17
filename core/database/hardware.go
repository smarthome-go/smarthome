package database

import (
	"fmt"
)

// Creates the table (unless it exists) which contains the hardware node
// If the database fails, this function returns an error
// The node's primary is its url
func CreateHardwareNodeTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	hardware(
		Url VARCHAR(50),
		Online BOOLEAN DEFAULT TRUE,
		Name VARCHAR(30),
		Token VARCHAR(100),
		PRIMARY KEY (url)
	)
	`
	if _, err := db.Exec(query); err != nil {
		log.Error("Failed to create hardware table: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Adds a new hardware node to the database, if the node already exists (same url), its name will be updated
func CreateHardwareNode(name string, url string, token string) error {
	query, err := db.Prepare(`
	INSERT INTO
	hardware(
		Url, Online, Name, Token
	)
	VALUES(?, DEFAULT, ?, ?)
	ON DUPLICATE KEY
	UPDATE Name=VALUES(Name)
	`)
	if err != nil {
		log.Error("Failed to create a new node: prepearing query failed: ", err.Error())
		return err
	}
	res, err := query.Exec(url, name, token)
	if err != nil {
		log.Error("Failed to create a new node: executing query failed: ", err.Error())
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("Could not get result of AddNode: obtaining rowsAffected failed: ", err.Error())
		return err
	}
	if rowsAffected > 0 {
		log.Debug(fmt.Sprintf("Added hardware node `%s` with url `%s`", name, url))
	}
	defer query.Close()
	return nil
}

// Updates the online / offline state of a given node (url)
func SetNodeOnline(nodeUrl string, online bool) error {
	query, err := db.Prepare(`
	UPDATE hardware
	SET Online=?
	WHERE Url=?
	`)
	if err != nil {
		log.Error("Failed to mark uptime status of node: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Exec(online, nodeUrl); err != nil {
		log.Error("Failed to mark uptime status of node: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes a node given its url
func DeleteHardwareNode(url string) error {
	query, err := db.Prepare(`
	DELETE FROM
	hardware
	WHERE Url=?
	`)
	if err != nil {
		log.Error("Failed to delete hardware node: preparing query failed: ", err.Error())
		return err
	}
	if _, err = query.Exec(url); err != nil {
		log.Error("Failed to delete hardware node: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns a list of hardware nodes
func GetHardwareNodes() ([]HardwareNode, error) {
	query := `
	SELECT
	Url, Online, Name, Token
	FROM hardware
	`
	res, err := db.Query(query)
	if err != nil {
		log.Error("Failed to list hardware nodes: executing query failed: ", err.Error())
		return nil, err
	}
	nodes := make([]HardwareNode, 0)
	for res.Next() {
		var node HardwareNode
		if err := res.Scan(
			&node.Url,
			&node.Online,
			&node.Name,
			&node.Token,
		); err != nil {
			log.Error("Failed to list hardware nodes: scanning results failed: ", err.Error())
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
