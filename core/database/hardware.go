package database

import (
	"database/sql"
	"fmt"
)

// Hardware node
type HardwareNode struct {
	Name    string `json:"name"`
	Online  bool   `json:"online"`
	Enabled bool   `json:"enabled"` // Can be used to temporarely deactivate a node in case of maintenance
	Url     string `json:"url"`
	Token   string `json:"token"`
}

// Creates the table (unless it exists) which contains the hardware node
// If the database fails, this function returns an error
// The node's primary is its url
func createHardwareNodeTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	hardware(
		Url VARCHAR(50),
		Online BOOLEAN DEFAULT TRUE,
		Enabled BOOLEAN DEFAULT TRUE,
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
func CreateHardwareNode(node HardwareNode) error {
	query, err := db.Prepare(`
	INSERT INTO
	hardware(
		Url,
		Online,
		Enabled,
		Name,
		Token
	)
	VALUES(?, DEFAULT, DEFAULT, ?, ?)
	ON DUPLICATE KEY
	UPDATE Name=VALUES(Name)
	`)
	if err != nil {
		log.Error("Failed to create a new node: prepearing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	res, err := query.Exec(node.Url, node.Name, node.Token)
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
		log.Debug(fmt.Sprintf("Added hardware node `%s` with url `%s`", node.Name, node.Url))
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
	defer query.Close()
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
	defer query.Close()
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
		Url,
		Online,
		Enabled,
		Name,
		Token
	FROM hardware
	`
	res, err := db.Query(query)
	if err != nil {
		log.Error("Failed to list hardware nodes: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()
	nodes := make([]HardwareNode, 0)
	for res.Next() {
		var node HardwareNode
		if err := res.Scan(
			&node.Url,
			&node.Online,
			&node.Enabled,
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

// Returns a hardware node given its url
func GetHardwareNodeByUrl(url string) (HardwareNode, bool, error) {
	query, err := db.Prepare(`
	SELECT
		Url,
		Online,
		Enabled,
		Name,
		Token
	FROM hardware
	WHERE Url=?
	`)
	if err != nil {
		log.Error("Failed to get Hardware node by url: preparing query failed: ", err.Error())
		return HardwareNode{}, false, err
	}
	var node HardwareNode
	if err := query.QueryRow(url).Scan(
		&node.Url,
		&node.Online,
		&node.Enabled,
		&node.Name,
		&node.Token,
	); err != nil {
		if err == sql.ErrNoRows {
			return HardwareNode{}, false, nil
		}
		log.Error("Failed to get Hardware node by url: executing query failed: ", err.Error())
		return HardwareNode{}, false, err
	}
	return node, true, nil
}

// Changes the metadata of a given node
// Does not affect the online boolean
// For changing the online status, use `SetNodeOnline`
func ModifyHardwareNode(url string, node HardwareNode) error {
	query, err := db.Prepare(`
	UPDATE hardware
	SET
		Enabled=?,
		Name=?,
		Token=?
	WHERE Url=?
	`)
	if err != nil {
		log.Error("Failed to modify Hardware node: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(
		node.Enabled,
		node.Name,
		node.Token,
		url,
	); err != nil {
		log.Error("Failed to modify Hardware node: executing query failed: ", err.Error())
		return err
	}
	return nil
}
