package database

import (
	"time"
)

type WeatherMeasurement struct {
	Id                 uint      `json:"id"`
	Time               time.Time `json:"time"`
	WeatherTitle       string    `json:"weatherTitle"`
	WeatherDescription string    `json:"weatherDescription"`
	Temperature        float32   `json:"temperature"`
	FeelsLike          float32   `json:"feelsLike"`
	Humidity           uint8     `json:"humidity"`
}

func createWeatherTable() error {
	if _, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	weather(
		Id						INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		Time					DATETIME DEFAULT CURRENT_TIMESTAMP,

		WeatherTitle			TEXT,
		WeatherDescription		TEXT,
		Temperature				FLOAT(24),
		FeelsLike				FLOAT(24),
		Humidity				INT UNSIGNED
	)
	`); err != nil {
		log.Error("Failed to create weather data table: executing query failed: ", err.Error())
		return err
	}
	return nil
}

func GetWeatherDataRecords(maxAgeMinutes uint) ([]WeatherMeasurement, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		Time,
		WeatherTitle,
		WeatherDescription,
		Temperature,
		FeelsLike,
		Humidity
	From weather
	WHERE
		Time > NOW() - INTERVAL ? MINUTE
	`)
	if err != nil {
		log.Error("Failed to get weather data records: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query(maxAgeMinutes)
	if err != nil {
		log.Error("Failed to get weather data records: executing failed: ", err.Error())
		return nil, err
	}
	results := make([]WeatherMeasurement, 0)
	for res.Next() {
		var row WeatherMeasurement
		// Scan the current row
		if err := res.Scan(
			&row.Id,
			&row.Time,
			&row.WeatherTitle,
			&row.WeatherDescription,
			&row.Temperature,
			&row.FeelsLike,
			&row.Humidity,
		); err != nil {
			log.Error("Failed to get weather data records: scanning query results failed: ", err.Error())
			return nil, err
		}
		// Append the current row to the results
		results = append(results, row)
	}
	return results, nil
}

func AddWeatherDataRecord(
	weatherTitle string,
	weatherDescription string,
	temperature float32,
	feelsLike float32,
	humidity uint8,
) (uint, error) {
	query, err := db.Prepare(`
	INSERT INTO
	weather(
		Id,
		Time,
		WeatherTitle,
		WeatherDescription,
		Temperature,
		FeelsLike,
		Humidity
	)
	VALUES(DEFAULT, DEFAULT, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to add weather measurement: preparing query failed: ", err.Error())
		return 0, err
	}
	res, err := query.Exec(
		weatherTitle,
		weatherDescription,
		temperature,
		feelsLike,
		humidity,
	)
	if err != nil {
		log.Error("Failed to add weather measurement: executing query failed: ", err.Error())
		return 0, err
	}
	newId, err := res.LastInsertId()
	if err != nil {
		log.Error("Failed to add weather measurement: retrieving newly inserted id failed: ", err.Error())
		return 0, err
	}
	return uint(newId), nil
}

func PurgeWeatherData() error {
	if _, err := db.Exec(`
	DELETE FROM
	weather
	`); err != nil {
		log.Error("Failed to purge weather cache: executing query failed: ", err.Error())
		return err
	}
	return nil
}
