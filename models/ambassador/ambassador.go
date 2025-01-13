package ambassador

import (
	"database/sql"
	"log"
	"tutree/student-apis/config"
)

// AmbassadorStats represents the statistics for an ambassador.
type AmbassadorStats struct {
	TotalSignup    int     `json:"total_signup"`
	TotalEarning   float64 `json:"total_earning"`
	MonthlySignup  int     `json:"monthly_signup"`
	MonthlyEarning float64 `json:"monthly_earning"`
}

// GetAmbassadorStats retrieves statistics for a given ambassador identified by userID.
// Input: UserID
// Output: AmbassadorStats{}, error
func GetAmbassadorStats(userID int) (AmbassadorStats, error) {
	// Connect to the database.
	db, err := config.GetDB2()
	var ambassadorStats AmbassadorStats
	if err != nil {
		log.Println("GetAmbassadorStats: Failed to connect to the database with error: ", err)
		return ambassadorStats, err
	}
	defer db.Close()

	// Define the SQL query to retrieve the statistics.
	query := `
		SELECT 
			COUNT(invitee) AS total_signup,
			SUM(amount_receivable) AS total_earning,
			COUNT(CASE WHEN created_at >= date_trunc('month', CURRENT_DATE) THEN invitee ELSE NULL END) AS monthly_signup,
			SUM(CASE WHEN created_at >= date_trunc('month', CURRENT_DATE) THEN amount_receivable ELSE 0 END) AS monthly_earning
		FROM
			user_invitee
		WHERE 
			inviter = $1
		GROUP BY inviter`

	// Variables to hold the query results.
	var (
		totalSignup    sql.NullInt64
		totalEarning   sql.NullFloat64
		monthlySignup  sql.NullInt64
		monthlyEarning sql.NullFloat64
	)

	// Execute the query and scan the results into the variables.
	err = db.QueryRow(query, userID).Scan(&totalSignup, &totalEarning, &monthlySignup, &monthlyEarning)
	if err != nil {
		log.Println("GetAmbassadorStats: Failed to execute the query with error: ", err)
		return ambassadorStats, err
	}

	// Populate the AmbassadorStats struct with the query results.
	ambassadorStats = AmbassadorStats{
		TotalSignup:    int(totalSignup.Int64),
		TotalEarning:   totalEarning.Float64,
		MonthlySignup:  int(monthlySignup.Int64),
		MonthlyEarning: monthlyEarning.Float64,
	}

	// Return the AmbassadorStats and nil error.
	return ambassadorStats, nil
}
