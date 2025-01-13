package utility

import (
	"fmt"
	"log"
	"strings"
	"time"
)

func ConvertToUTCTime(timeStr string) (time.Time, error) {

	// Example time string in ISO 8601 forma
	// Parse the time string to a time.Time object
	// parsedTime, err := time.Parse(time.RFC3339, time)
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		log.Println("Error parsing time:", err)
		return time.Time{}, err
	}

	// Ensure the time is in UTC
	utcTime := parsedTime.UTC()

	// Format the time in ISO 8601 format
	formattedUTC := utcTime.Format(time.RFC3339)

	log.Println("Original time string:", timeStr)
	log.Println("Parsed UTC time:", formattedUTC)
	return utcTime, nil

}

const (
	DateLayout = "2006-01-02"
	TimeLayout = "15:04:05"
)

func GetCurrenDate() string {
	return time.Now().Format(DateLayout)
}
func GetCurrentTime() string {
	return time.Now().Format(TimeLayout)
}

// GetParsedOrCurrentDate and formats a given date string. If the date string is empty or only contains whitespace,
// it returns the current date in the "2006-01-02" format. If the date string is not empty, it attempts to parse
// the date string using the "2006-01-02" format and returns the formatted date. If parsing fails, an error is logged
// and returned.
//
// Parameters:
//   - dateParam: a string representing a date in the "2006-01-02" format, or an empty string.
//
// Returns:
//   - string: the formatted date string.
//   - error: an error if the date parsing fails.
func GetFormattedDate(dateParam string) (string, error) {
	dateParam = strings.TrimSpace(dateParam)

	if dateParam == "" || len(dateParam) == 0 {
		// currentDate := GetCurrenDate()
		return "", nil
	} else {
		dateToParsed, err := time.Parse(DateLayout, dateParam)
		if err != nil {
			log.Println("[ERROR] GetParsedOrCurrentDate: Failed to parse the date-to value with ", err)
			return "", err
		}
		return dateToParsed.Format(DateLayout), nil
	}
}

// GetDate extracts the date in "YYYY-MM-DD" format from a datetime string
func GetDate(datetime string) (string, error) {
	parsedTime, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		return "", err
	}
	return parsedTime.Format(DateLayout), nil
}

// GetTime extracts the time in "HH:MM:SS" format from a datetime string
func GetTime(datetime string) (string, error) {
	parsedTime, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		return "", err
	}
	return parsedTime.Format(TimeLayout), nil
}

func DayOfWeek(date string) string {
	// Parse the date string to a time.Time object
	layout := "2006-01-02"
	t, err := time.Parse(layout, date)
	if err != nil {
		return ""
	}

	// Get the day of the week
	dayOfWeek := t.Weekday().String()
	return dayOfWeek
}
func FormatToReadableDateTime(dataTime time.Time, timeZone string) (string, string) {
	// ConvertTimeZone formats a time.Time object to a given time zone and format
	convertedTime, err := ConvertTimeZone(dataTime, timeZone)
	if err != nil {
		log.Println("FormatToReadableDateTime : Failed to convert time according to given time zone with error :", err)
		return "", ""
	}
	// Convert to desired format "Thursday, May 30, at 4:45 PM"
	timePart := convertedTime.Format("3:04 PM")
	datePart := convertedTime.Format("Monday, January 2")

	// Print the formatted time
	log.Println(datePart, " at ", timePart)
	return datePart, timePart
}

// ConvertTimeZone formats a time.Time object to a given time zone and format
func ConvertTimeZone(dateTime time.Time, timeZone string) (time.Time, error) {
	// Load the location for the desired time zone
	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return time.Time{}, fmt.Errorf("ConvertTimeZone :error loading location for time zone %s: %w", timeZone, err)
	}

	// Convert the time to the desired time zone
	localTime := dateTime.In(location)

	// Format the time according to the provided format
	// formattedTime := localTime.Format(format)

	return localTime, nil
}
func SessionEndDateTime(sessionStart time.Time) time.Time {
	sessionDuration := SessionDuration
	return sessionStart.Add(time.Duration(sessionDuration) * time.Minute)
}
func ParseDateWithEndOfDayTime(endDateStr string) (time.Time, error) {
	var endDate time.Time
	if len(endDateStr) != 0 {
		layout := time.RFC3339
		parsedTime, err := time.Parse(layout, endDateStr)
		if err != nil {
			log.Println("[ERROR] ParseDateWithEndOfDayTime: Failed while try to parse string into time.Time with error: ", err)
			return time.Time{}, err
		}

		// Extract only the date portion (Year, Month, Day)
		//endDate := parsedTime.Format("2006-01-02")
		endDate = time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 23, 59, 59, 0, parsedTime.Location())
	}
	return endDate, nil
}

// GetTimeWithAMOrPM takes a time string in "HH:mm" (24-hour format) as input,
// parses it, and returns the time formatted in 12-hour format with AM or PM.
func GetTimeWithAMOrPM(timeStr string) string {
	parsedTime, err := time.Parse(TimeLayout, timeStr) // Updated format to "15:04" for proper parsing of hours and minutes
	if err != nil {
		log.Println("[ERROR] GetTimeWithAMOrPM: Failed to parse time, with error:", err)
		return ""
	}

	timeWithPeriod := parsedTime.Format("03:04 PM")

	return timeWithPeriod
}

// GetTimeZoneAbbreviation takes a full timezone string and returns its abbreviation.
func GetTimeZoneAbbreviation(timezone string) string {
	// Load the location based on the timezone string
	location, err := time.LoadLocation(timezone)
	if err != nil {
		log.Println("[ERROR] GetTimeZoneAbbreviation: Failed to load location:", err)
		return ""
	}

	// Get the current time and convert it to the specified location
	currentTime := time.Now().In(location)

	// Return the time zone abbreviation (e.g., IST, PST)
	abbreviation := currentTime.Format("MST") // MST gives the time zone abbreviation
	return abbreviation
}

// ConvertTimeInUTC converts a specified day and time from a given timezone to a UTC time string.
//
// Parameters:
// - day: The day of the week (e.g., "Monday").
// - timeStr: The time in "HH:MM:SS" format to be converted.
// - timeZone: The timezone of the provided day and time.
//
// Returns:
// - A string representing the converted time in UTC in "HH:MM:SS" format.
// - An error if loading the timezone or parsing the time fails.
func ConvertTimeInUTC(day, timeStr, timeZone string) (string, string, error) {
	// Load the timezone location
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return "", "", fmt.Errorf("[ERROR] ConvertTimeInUTC: Failed to load location: %w", err)
	}

	// Get today's date and the weekday (e.g., "Monday", "Tuesday", etc.)
	currentTime := time.Now()

	// Calculate how many days ahead or behind we are from the provided day
	targetDay := getDayOfWeek(day)
	daysToAdd := targetDay - int(currentTime.Weekday())
	if daysToAdd <= 0 {
		daysToAdd += 7 // Move to the next week
	}

	// Add the days to get the next occurrence of the target day
	adjustedDate := currentTime.AddDate(0, 0, daysToAdd)

	// Combine the adjusted date with the time string
	combinedDateTime := fmt.Sprintf("%s %s %02d-%02d-%d", day, timeStr, adjustedDate.Month(), adjustedDate.Day(), adjustedDate.Year())

	// Parse the combined date and time
	localParsedTime, err := time.ParseInLocation("Monday 15:04:05 01-02-2006", combinedDateTime, loc)
	if err != nil {
		return "", "", fmt.Errorf("[ERROR] ConvertTimeInUTC : Failed to parse time: %w", err)
	}

	// Convert the local time to UTC
	utcTime := localParsedTime.UTC()

	// Return only the time part in UTC
	return utcTime.Format("Monday"), utcTime.Format("15:04:05"), nil
}

// getDayOfWeek converts a weekday name (e.g., "friday") into the corresponding int value (0-6)
func getDayOfWeek(day string) int {
	dayMap := map[string]int{
		"sunday":    0,
		"monday":    1,
		"tuesday":   2,
		"wednesday": 3,
		"thursday":  4,
		"friday":    5,
		"saturday":  6,
	}

	// Return the value from the map, or -1 if the day is not found
	if val, ok := dayMap[day]; ok {
		return val
	}
	return -1
}
