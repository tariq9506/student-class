package models

import (
	"database/sql"
	"log"
	"tutree/student-apis/config"
)

// GetLocationByZipcode retrieves the city, state, and country associated with a given zipcode from the database.
// Parameters:
// - zipcode: The zipcode for which the location information is to be retrieved.
// Returns:
// - city: The name of the city associated with the given zipcode.
// - state: The name of the state associated with the given zipcode.
// - country: The name of the country associated with the given zipcode.
// - error: Any error encountered during the process.
func GetLocationByZipcode(zipcode string) (string, string, string, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetLocationByZipcode: Failed while connecting with the database :", err)
		return "", "", "", err
	}
	defer db.Close()

	var (
		city    sql.NullString
		state   sql.NullString
		country sql.NullString
	)

	query := `select city_name,state,country from zipcode where zip = $1`

	err = db.QueryRow(query, zipcode).Scan(&city, &state, &country)
	if err != nil {
		log.Println("GetLocationByZipcode: Failed to make query :", err)
		return "", "", "", err
	}

	return city.String, state.String, country.String, nil
}
