package database

import (
	"database/sql"
	"unicode/utf8"
)

// Creates the table which hold the cached Homescript URLs
func createHomescriptUrlCacheTable() error {
	if _, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	homescriptUrlCache(
		Url VARCHAR(100),
		LastCheck DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(Url)
	)
	`); err != nil {
		log.Error("Failed to create homescriptUrlCache table: ", err.Error())
		return err
	}
	return nil
}

// Adds a new entry into the cache
// If the entry already exists, it is updated
func AddHomescriptUrlCacheEntry(url string) error {
	if utf8.RuneCountInString(url) > 100 {
		// URLs longer than 100 characters cannot be cached and must therefore be omitted
		return nil
	}
	query, err := db.Prepare(`
	INSERT INTO
	homescriptUrlCache(
		Url,
		LastCheck
	)
	VALUES(?, DEFAULT)
	ON DUPLICATE KEY
	UPDATE
		LastCheck=VALUES(LastCheck)
	`)
	if err != nil {
		log.Error("Failed to add or update Homescript URL cache entry: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(url); err != nil {
		log.Error("Failed to add or update Homescript URL cache entry: execution query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns all cache entries younger than 12 hours (cache invalidation time)
// => This means this function will return only valid cache-entries
func GetValidHomescriptUrlCacheEntries() ([]string, error) {
	res, err := db.Query(`
	SELECT Url
	FROM homescriptUrlCache
	WHERE LastCheck > NOW() - INTERVAL 12 HOUR
	`)
	if err != nil {
		log.Error("Failed to list valid Homescript URL cache entries: executing query failed: ", err.Error())
		return nil, err
	}
	results := make([]string, 0)
	for res.Next() {
		var currentResult string
		if err := res.Scan(&currentResult); err != nil {
			log.Error("Failed to list valid Homescript URL cache entries: scanning result failed: ", err.Error())
			return nil, err
		}
		results = append(results, currentResult)
	}
	return results, nil
}

// Returns a boolean indicating whether a specified entry exists and if it is still valid
func IsHomescriptUrlCached(urlToGet string) (bool, error) {
	query, err := db.Prepare(`
	SELECT 1
	FROM homescriptUrlCache
	WHERE LastCheck > NOW() - INTERVAL 12 HOUR
	AND Url=?
	LIMIT 1
	`)
	if err != nil {
		log.Error("Failed to check if HMS url cache entry exists: preparing query failed: ", err.Error())
		return false, err
	}
	defer query.Close()
	blackHole := 0
	if err := query.QueryRow(urlToGet).Scan(&blackHole); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Error("Failed to check if HMS url cache entry exists: executing query failed: ", err.Error())
		return false, err
	}
	return true, nil
}

// Deletes all cache entries which are older than 12 hours (cache invalidation time)
func FlushHomescriptUrlCache() error {
	if _, err := db.Exec(`
	DELETE FROM homescriptUrlCache
	WHERE LastCheck < NOW() - INTERVAL 12 HOUR
	`); err != nil {
		log.Error("Failed to flush Homescript URL cache: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes all cache entries, regardless of their age
func PurgeHomescriptUrlCache() error {
	if _, err := db.Exec(`
	DELETE FROM homescriptUrlCache
	`); err != nil {
		log.Error("Failed to purge Homescript URL cache: executing query failed: ", err.Error())
		return err
	}
	return nil
}
