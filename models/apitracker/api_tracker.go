package apitracker

import (
	"log"
	"tutree/student-apis/config"
)

type APITracker struct {
	ID          int32  `json:"id"`
	URL         string `json:"url"`
	IP          string `json:"ip"`
	Query       string `json:"query"`
	UserSession string `json:"user_session"`
	APIMethod   string `json:"api_method"`
	BodyParams  string `json:"body_params"`
	UserAgent   string `json:"user_agent"`
}

func StoreAPIHistory(apiTracker APITracker) {

	db, err := config.GetDB2()
	if err != nil {
		log.Println("StoreAPIHistory: Failed while connecting with the database with error:", err)
		return
	}
	defer db.Close()

	query := `INSERT INTO api_tracker (full_url, ip, query_param, user_session, api_method, body_param, user_agent, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())`
	_, err = db.Exec(query, apiTracker.URL, apiTracker.IP, apiTracker.Query, apiTracker.UserSession, apiTracker.APIMethod, apiTracker.BodyParams, apiTracker.UserAgent)

	if err != nil {
		log.Println("StoreAPIHistory: failed while executing the query with error:", err)
		return
	}
}
