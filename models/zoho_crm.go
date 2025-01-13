package models

import (
	"database/sql"
	"log"
	"time"
	"tutree/student-apis/config"
)

// ZohoCRMLeads represents the structure for storing lead data to be sent to Zoho CRM.
//
// Fields:
// - Name: Full name of the user or lead.
// - PhoneNumber: Contact number of the user.
// - CreatedDate: Date and time when the lead was created.
// - EmailAddres: Email address of the user.
// - Source: The source from where the lead was generated (e.g., a campaign or referral).
// - PlanName: Subscription or plan name associated with the lead.
// - PlaneStatus: Indicates if the plan is active or inactive.
// - DemoBooked: A flag to denote if a demo session has been booked by the user.
// - PhoneNumberVerified: Indicates whether the user's phone number has been verified.
type ZohoCRMLeads struct {
	Name                string
	PhoneNumber         string
	CreatedDate         time.Time
	EmailAddres         string
	Source              string
	PlanName            string
	PlaneStatus         bool
	IsDemoBooked        bool
	PhoneNumberVerified bool
}

// GetUserDetailsForCRMLeads fetches user details from the database to be sent to Zoho CRM as lead information.
//
// Parameters:
// - userID: The ID of the user whose details need to be retrieved.
//
// Returns:
// - ZohoCRMLeads: A struct containing user lead details such as phone number, verification status, source, name, email, plan name, and plan status.
// - error: An error, if any occurs during database connection or query execution.
//
// Steps:
// 1. Connect to the database.
// 2. Execute a SQL query to fetch user-related data by joining `user`, `student`, `user2subscription`, and `subscription_plan` tables.
// 3. Scan the query result into local variables, handling nullable fields.
// 4. Map the scanned fields to the `ZohoCRMLeads` struct.
// 5. Return the populated struct or an error if the process fails.
func GetUserDetailsForCRMLeads(userID int) (ZohoCRMLeads, error) {
	var crmLeads ZohoCRMLeads
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetUserDetailsForCRMLeads: Error connecting to the database:", err)
		return crmLeads, err
	}
	defer db.Close()
	query := `
				SELECT 
					u.phone_number,
					u.phone_verified,
					u.query_param,
					st.name,
					st.parent_email,
					sp.name AS plan_name,
					us.status AS plan_status,
					CASE 
       					WHEN s.type = 'demo' THEN TRUE
        				ELSE FALSE
    				END AS demo_booked
				FROM 
					public.user AS u 
				LEFT JOIN 
					public.session AS s ON s.user_id = u.id
				LEFT JOIN 
					public.student AS st ON u.id = st.user_id
				LEFT JOIN 
					public.user2subscription AS us ON u.id = us.user_id
				LEFT JOIN 
					public.subscription_plan AS sp ON us.subscription_plan_id = sp.id
				WHERE 
					u.id = $1
				LIMIT 1`
	var (
		phone, email, source, name, planeName   sql.NullString
		phoneVerified, planStatus, isDemoBooked sql.NullBool
	)
	err = db.QueryRow(query, userID).Scan(&phone, &phoneVerified, &source, &name, &email, &planeName, &planStatus, &isDemoBooked)
	if err != nil {
		log.Println("GetUserDetailsForCRMLeads: Error while execute the query to fetch the leads from the database:", err)
		return crmLeads, err
	}
	crmLeads = ZohoCRMLeads{
		PhoneNumber:         phone.String,
		PhoneNumberVerified: phoneVerified.Bool,
		Source:              source.String,
		Name:                name.String,
		EmailAddres:         email.String,
		PlanName:            planeName.String,
		PlaneStatus:         planStatus.Bool,
		IsDemoBooked:        isDemoBooked.Bool,
	}
	return crmLeads, nil
}
