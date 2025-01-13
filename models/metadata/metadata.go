package metadata

import (
	"database/sql"
	"log"
	"tutree/student-apis/config"
)

type Subject struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Grade struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	SortLevel   int64  `json:"sort_level"`
	Description string `json:"description"`
}

// GetSubjects retrieves a list of all subjects from the database.
// It queries the 'subject' table and returns a slice of Subject objects.
// If an error occurs during the database connection or query execution, it returns an error.
func GetSubjects() ([]Subject, error) {
	var subjectList []Subject

	// Establish a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetSubjects: Failed to connect to the database:", err)
		return nil, err
	}
	defer db.Close()
	// SQL query to select all subjects.
	query := `
			SELECT 
				id,
				name,
				description
			FROM subject
			ORDER BY id`

	// Execute the SQL query and obtain the result set.
	rows, err := db.Query(query)
	if err != nil {
		log.Println("GetSubjects: Failed to execute the query:", err)
		return nil, err
	}
	defer rows.Close()

	// Iterate through each row in the result set.
	for rows.Next() {
		var (
			id          sql.NullInt64
			name        sql.NullString
			description sql.NullString
		)

		// Scan the current row's values into the defined variables.
		err = rows.Scan(&id, &name, &description)
		if err != nil {
			log.Println("GetSubjects: Failed to scan the row:", err)
			continue
		}

		// Append the scanned subject to the subjectList slice.
		subjectList = append(subjectList, Subject{
			ID:          id.Int64,
			Name:        name.String,
			Description: description.String,
		})
	}
	return subjectList, nil
}

// GetGrades retrieves a list of all grades from the database.
// It queries the 'grade' table and returns a slice of Grade objects.
// If an error occurs during the database connection or query execution, it returns an error.
func GetGrades() ([]Grade, error) {
	var gradeList []Grade

	// Establish a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetGrades: Failed to connect to the database:", err)
		return gradeList, err
	}
	defer db.Close()

	// SQL query to select all grades, ordered by their sort level.
	query := `
			SELECT 
				id,
				name,
				sort_level,
				description
			FROM grade 
			ORDER BY sort_level`

	rows, err := db.Query(query)
	if err != nil {
		log.Println("GetGrades: Failed to execute the query:", err)
		return gradeList, err
	}
	defer rows.Close()

	// Iterate through each row in the result set.
	for rows.Next() {
		var (
			id, sortLevel     sql.NullInt64
			name, description sql.NullString
		)

		// Scan the current row's values into the defined variables.
		err = rows.Scan(&id, &name, &sortLevel, &description)
		if err != nil {
			log.Println("GetGrades: Failed to scan the row:", err)
			continue
		}

		// Append the scanned grade to the gradeList slice.
		gradeList = append(gradeList, Grade{
			ID:          id.Int64,
			Name:        name.String,
			SortLevel:   sortLevel.Int64,
			Description: description.String,
		})
	}
	return gradeList, nil
}

// GetSubjectById is responcible to fetch subject details from data base by using
// suject id.
func GetSubjectById(subjectID int) (Subject, error) {
	var subject Subject
	// Establish a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		log.Println("[ERROR] GetSubjectById: Failed to connect to the database with error:", err)
		return subject, err
	}
	defer db.Close()
	query := `
			SELECT 
				name,
				description
			FROM
				subject
			WHERE id = $1`
	var name, description sql.NullString
	err = db.QueryRow(query, subjectID).Scan(&name, &description)
	if err != nil {
		log.Println("[ERROR] GetSubjectById: Failed to execute query to fetch the data from database.", err)
		return subject, err
	}
	subject = Subject{
		Name:        name.String,
		Description: description.String,
	}
	return subject, nil
}
