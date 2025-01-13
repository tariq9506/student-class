package models

import (
	"database/sql"
	"log"
	"tutree/student-apis/config"
)

// CountryDetail represents the details of a supported country.
type CountryDetail struct {
	CountryName          string         `json:"country_name"`
	CountryCode          string         `json:"country_code"`
	CountryPhoneCode     string         `json:"country_phone_code"`
	CountryFlagURL       string         `json:"country_flag_url"`
	SupportedStudentRole sql.NullString `json:"supported_student_role"`
	EmployerSupported    sql.NullBool   `json:"employer_supported"`
	DefaultPayoutPercent float64        `json:"default_cpa"`
	IsCountrySupported   bool           `json:"is_country_supported"`
	CurrencyCode         string         `json:"currency_code"`
	CountryID            int64          `json:"country_id"`
}

// GetDetailsOfSupportedCountryByCode retrieves the details of a supported country by its code.
// Parameters:
// - countryCode: The country code (ISO2 or name) for which the details are to be retrieved.
// Returns:
// - CountryDetail: The details of the supported country.
// - error: Any error encountered during the process.
func GetDetailsOfSupportedCountryByCode(countryCode string) (CountryDetail, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Printf("GetDetailsOfSupportedCountry: failed while connecting with the database: %v", err)
		return CountryDetail{}, err
	}
	defer db.Close()

	var (
		sc     CountryDetail
		exists bool
	)
	exists, err = CheckCountryIsSupported(countryCode)
	if err != nil {
		log.Printf("GetDetailsOfSupportedCountry: failed while checking if country code exists: %v", err)
		return CountryDetail{}, err
	}
	// If country code is not provided, set it to US
	if !exists || len(countryCode) == 0 {
		countryCode = "US"
	}

	query := `
		SELECT 
			id,
			name,
			iso2,
			phonecode
		FROM 
			countries
		WHERE 
			 `

	// The code block is checking the length of the `countryCode (iso2)` variable. If the length is equal to 2,
	// it means that the `countryCode(iso2)` is a two-letter country code. In this case, the query is updated to
	// search for the country using the `country_code` column in the database table. If the length is not
	// equal to 2, it means that the `countryCode` is a country name. In this case, the query is updated
	// to search for the country using the `name` column in the database table.
	if len(countryCode) == 2 {
		query += `iso2 = $1`
	} else {
		query += `iso2 = $1`
	}

	err = db.QueryRow(query, countryCode).Scan(
		&sc.CountryID,
		&sc.CountryName,
		&sc.CountryCode,
		&sc.CountryPhoneCode,
	)

	if err != nil {
		log.Printf("GetDetailsOfSupportedCountry failed while executing the query with :%v", err)
		return CountryDetail{}, err
	}
	// Return the retrieved data as a SupportedCountry struct as sc
	return sc, nil
}

// CheckCountryIsSupported checks if a country is supported based on the given country code or name.
// Parameters:
// - countryCode: The country code (ISO2 or name) to check for support.
// Returns:
// - bool: True if the country is supported, false otherwise.
// - error: Any error encountered during the process.s
func CheckCountryIsSupported(countryCode string) (bool, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("CheckCountryIsSupported: Failed while connecting with the database :", err)
		return false, err
	}
	defer db.Close()

	var isSupported sql.NullBool

	query := `SELECT EXISTS (
		SELECT 1 FROM countries 
		WHERE `

	// The above code is checking the length of the variable `countryCode`. If the length is equal to 2,
	// it appends a condition to the `query` string to search for a matching country code. If the length
	// is not equal to 2, it appends a condition to the `query` string to search for a matching country
	// name.
	if len(countryCode) == 2 {
		query += `iso2 = $1
		)`
	} else {
		query += `iso2 = $1
		)`
	}

	err = db.QueryRow(query, countryCode).Scan(&isSupported)
	if err != nil {
		log.Println("CheckCountryIsSupported: Failed while querying with:", err)
		return false, err
	}

	return isSupported.Bool, nil
}
