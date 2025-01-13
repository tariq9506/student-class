package timezone

import (
	"log"
	"tutree/student-apis/config"
)

type Timezone struct {
	ID           int64  `json:"id"`
	Identifier   string `json:"identifier"`
	Abbreviation string `json:"abbreviation"`
	CountryName  string `json:"country_name"`
	CountryCode  string `json:"country_code"`
}

func GetTimezones() ([]Timezone, error) {
	var timezones []Timezone

	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetTimezones: Failed when try to connect with database with error: ", err)
		return []Timezone{}, err
	}
	defer db.Close()

	query := `SELECT 
				id, 
				identifier, 
				abbreviation, 
				country_name, 
				country_code
			FROM
				timezones
			`

	rows, err := db.Query(query)
	if err != nil {
		log.Println("GetTimezones: Failed to execute the query with error: ", err)
		return []Timezone{}, err
	}

	for rows.Next() {
		var (
			id           int64
			identifier   string
			abbreviation string
			countryName  string
			countryCode  string
		)

		err = rows.Scan(&id, &identifier, &abbreviation, &countryName, &countryCode)
		if err != nil {
			log.Println("GetTimezones: Failed to scan the query with error: ", err)
			continue
		}
		timezones = append(timezones, Timezone{
			ID:           id,
			Identifier:   identifier,
			Abbreviation: abbreviation,
			CountryName:  countryName,
			CountryCode:  countryCode,
		})
	}
	return timezones, nil
}
