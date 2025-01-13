package services

import (
	"log"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

// Location represents geographical information.

type Location struct {
	Latitude    float64 `json:"Latitude,omitempty"`
	Longitude   float64 `json:"Longitude,omitempty"`
	CountryCode string  `json:"country_code,omitempty"`
	Country     string  `json:"country,omitempty"`
	PostalCode  string  `json:"postal_code,omitempty"`
	City        string  `json:"city,omitempty"`
	State       string  `json:"state,omitempty"`
	TimeZone    string  `json:"time_zone,omitempty"`
}

// GetLocationLocally retrieves geographical information based on the provided IP address using a local MaxMind GeoLite2 database.
// Parameters:
// - ip: The IP address for which geographical information is to be retrieved.
// Returns:
//   - Location: A struct containing country name, country ISO code, city name, postal code, latitude, and longitude.
//     If an error occurs during database access or lookup, an empty Location struct is returned.
func GetLocationLocally(ip string) Location {

	db, err := maxminddb.Open("./db/GeoLite2-City.mmdb")
	if err != nil {
		log.Println(err)
		return Location{}
	}
	defer db.Close()

	ipParsed := net.ParseIP(ip)

	var record struct {
		Country struct {
			ISOCode string            `maxminddb:"iso_code"`
			Names   map[string]string `maxminddb:"names"`
		} `maxminddb:"country"`
		City struct {
			Names map[string]string `maxminddb:"names"`
		} `maxminddb:"city"`
		Location struct {
			Latitude  float64 `maxminddb:"latitude"`
			Longitude float64 `maxminddb:"longitude"`
			TimeZone  string  `maxminddb:"time_zone"`
		} `maxminddb:"location"`
		Postal struct {
			Code string `maxminddb:"code"`
		} `maxminddb:"postal"`
	}

	err = db.Lookup(ipParsed, &record)
	if err != nil {
		log.Println(err)
		return Location{}
	}

	location := Location{
		Country:     record.Country.Names["en"],
		CountryCode: record.Country.ISOCode,
		City:        record.City.Names["en"],
		PostalCode:  record.Postal.Code,
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
		TimeZone:    record.Location.TimeZone,
	}
	return location
}
