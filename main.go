package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

type User struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password,omitempty"`
	Role        string `json:"role"`
	CompanyID   int    `json:"company_id"`
	CompanyName string `json:"company_name"`
}

type Company struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ReferralRequest struct {
	ID                 int       `json:"id"`
	Title              string    `json:"title"`
	Content            string    `json:"content"`
	Username           string    `json:"username"`
	ReferrerUserID     int       `json:"referrer_user_id"`
	ReferrerUsername   string    `json:"referrer_username"`
	CompanyID          int       `json:"company_id"`
	SentCompanyID      int       `json:"sent_company_id"`
	SentCompanyName    string    `json:"sent_company_name"`
	ReceivingCompanyID int       `json:"receiving_company_id"`
	RefereeClient      string    `json:"referee_client"`
	RefereeClientEmail string    `json:"referee_client_email"`
	CreatedAt          time.Time `json:"created_at"`
	Status             string    `json:"status"`
	CompanyName        string    `json:"company_name"`
}

func main() {
	user := "postgres"
	password := "postgres"
	dbName := "dbtest7"

	createDatabaseIfNotExists(dbName, user, password)
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbName)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}

	createTables()
	log.Println("Tables created successfully")

	// Initialize mux router
	r := mux.NewRouter()

	// Registering handlers
	r.HandleFunc("/register", RegisterHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")
	r.HandleFunc("/logout", LogoutHandler).Methods("POST")
	r.HandleFunc("/create-referral", CreateReferralRequestHandler).Methods("POST")
	r.HandleFunc("/companies", GetCompaniesHandler).Methods("GET")
	r.HandleFunc("/referrals-sent", GetReferralsSentHandler).Methods("GET")
	r.HandleFunc("/referrals-received", GetReferralsReceivedHandler).Methods("GET")
	r.HandleFunc("/referral-request-action/approve/{referralRequestID}", ApproveReferralRequestHandler).Methods("POST")
	r.HandleFunc("/referral-request-action/deny/{referralRequestID}", DenyReferralRequestHandler).Methods("POST")
	r.HandleFunc("/users", GetAllUsersHandler).Methods("GET")
	r.HandleFunc("/users/{companyID}", GetUsersByCompanyHandler).Methods("GET")
	r.HandleFunc("/create-user", CreateUserHandler).Methods("POST")
	r.HandleFunc("/delete-user", DeleteUserHandler).Methods("POST")
	r.HandleFunc("/create-company", CreateCompanyHandler).Methods("POST")
	r.HandleFunc("/delete-company", DeleteCompanyHandler).Methods("POST")
	r.HandleFunc("/get-user-info", GetUserInfoHandler).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)
	http.ListenAndServe(":8080", handler)
}

// Assuming you have a function to get the current user from the session or token
func GetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user User
	err = db.QueryRow(`
        SELECT u.id, u.email, u.username, u.company_id 
        FROM users u 
        INNER JOIN sessions s ON u.id = s.user_id 
        WHERE s.session_id = $1
    `, sessionID.Value).Scan(&user.ID, &user.Email, &user.Username, &user.CompanyID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Respond with user information
	response := map[string]interface{}{
		"companyId": user.CompanyID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Function to create a database if it doesn't exist
func createDatabaseIfNotExists(dbName string, user string, password string) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=postgres sslmode=disable", user, password)
	tempDB, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the PostgreSQL instance: %v", err)
	}
	defer tempDB.Close()

	// Check if the database exists
	var exists bool
	err = tempDB.QueryRow("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check if database exists: %v", err)
	}

	if !exists {
		// Create the database
		_, err = tempDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		log.Printf("Database %s created successfully", dbName)
	} else {
		log.Printf("Database %s already exists", dbName)
	}
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS companies (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			company_id INTEGER,
			FOREIGN KEY (company_id) REFERENCES companies(id)
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id SERIAL PRIMARY KEY,
			session_id TEXT UNIQUE NOT NULL,
			user_id INTEGER NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS referral_requests (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			referrer_user_id INTEGER NOT NULL,
			company_id INTEGER NOT NULL,
			receiving_company_id INTEGER,
			sent_company_id INTEGER,
			referee_client TEXT NOT NULL,
			referee_client_email TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			status TEXT NOT NULL DEFAULT 'pending',
			FOREIGN KEY (referrer_user_id) REFERENCES users(id),
			FOREIGN KEY (receiving_company_id) REFERENCES companies(id),
			FOREIGN KEY (sent_company_id) REFERENCES companies(id)
		);`,
	}
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatalf("Failed to execute query: %s, error: %v", query, err)
		}
	}
}

func createSession(userID int) (string, error) {
	sessionID := generateSessionID()
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err := db.Exec("INSERT INTO sessions (session_id, user_id, expires_at) VALUES ($1, $2, $3)", sessionID, userID, expiresAt)
	if err != nil {
		return "", err
	}
	return sessionID, nil
}

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user User

	// Decode JSON request body into User struct
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if user.Email == "" || user.Username == "" || user.Password == "" || user.CompanyName == "" || user.Role == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Check if company exists or create a new one
	var companyID int
	err = db.QueryRow("SELECT id FROM companies WHERE name = $1", user.CompanyName).Scan(&companyID)
	switch {
	case err == sql.ErrNoRows:
		// Company does not exist, insert it
		log.Printf("Company %s does not exist. Creating new company.", user.CompanyName)
		err = db.QueryRow("INSERT INTO companies (name) VALUES ($1) RETURNING id", user.CompanyName).Scan(&companyID)
		if err != nil {
			log.Printf("Error inserting company: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		log.Printf("New company created with ID: %d", companyID)
	case err != nil:
		// Some other error occurred
		log.Printf("Error checking company existence: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	default:
		log.Printf("Existing company found with ID: %d", companyID)
	}

	// Insert the user with hashed password and company ID
	log.Printf("Inserting user %s into company with ID %d", user.Username, companyID)
	_, err = db.Exec("INSERT INTO users (email, username, password, role, company_id) VALUES ($1, $2, $3, $4, $5)",
		user.Email, user.Username, hashedPassword, user.Role, companyID)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "User registered successfully")
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var user User
	err := db.QueryRow("SELECT id, email, username, password, role, company_id FROM users WHERE email = $1", credentials.Email).
		Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Role, &user.CompanyID)
	if err != nil {
		log.Println("Error querying user:", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		log.Println("Error comparing password:", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	sessionID, err := createSession(user.ID)
	if err != nil {
		log.Println("Error creating session:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session_id",
		Value:   sessionID,
		Expires: time.Now().Add(24 * time.Hour),
	})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "No session found", http.StatusUnauthorized)
		return
	}
	_, err = db.Exec("DELETE FROM sessions WHERE session_id = $1", cookie.Value)
	if err != nil {
		log.Println("Error deleting session:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		MaxAge: -1,
	})
	w.WriteHeader(http.StatusOK)
}

func GetCompaniesHandler(w http.ResponseWriter, r *http.Request) {
	var userCount int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		log.Println("Error counting users:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// If no users exist, allow public access to companies
	if userCount == 0 {
		fetchCompanies(w)
		return
	}

	// Proceed with session authentication if users exist
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user User
	err = db.QueryRow("SELECT u.id, u.email, u.username, u.password, u.role, u.company_id FROM users u "+
		"INNER JOIN sessions s ON u.id = s.user_id "+
		"WHERE s.session_id = $1", sessionID.Value).
		Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Role, &user.CompanyID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fetchCompanies(w)
}

func fetchCompanies(w http.ResponseWriter) {
	var companies []Company
	rows, err := db.Query("SELECT id, name FROM companies")
	if err != nil {
		log.Println("Error fetching companies:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var company Company
		err := rows.Scan(&company.ID, &company.Name)
		if err != nil {
			log.Println("Error scanning company row:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		companies = append(companies, company)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over company rows:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(companies) == 0 {
		companies = []Company{} // Ensure an empty slice is returned if no records found
	}
	json.NewEncoder(w).Encode(companies)
}

func ApproveReferralRequestHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var user User
	err = db.QueryRow("SELECT u.id, u.email, u.username, u.password, u.role, u.company_id FROM users u "+
		"INNER JOIN sessions s ON u.id = s.user_id "+
		"WHERE s.session_id = $1", sessionID.Value).
		Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Role, &user.CompanyID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	referralRequestID, err := strconv.Atoi(vars["referralRequestID"])
	if err != nil {
		http.Error(w, "Invalid referral request ID", http.StatusBadRequest)
		return
	}
	_, err = db.Exec("UPDATE referral_requests SET status = $1 WHERE id = $2", "Approved", referralRequestID)
	if err != nil {
		log.Println("Error approving referral request:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func DenyReferralRequestHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var user User
	err = db.QueryRow("SELECT u.id, u.email, u.username, u.password, u.role, u.company_id FROM users u "+
		"INNER JOIN sessions s ON u.id = s.user_id "+
		"WHERE s.session_id = $1", sessionID.Value).
		Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Role, &user.CompanyID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	referralRequestID, err := strconv.Atoi(vars["referralRequestID"])
	if err != nil {
		http.Error(w, "Invalid referral request ID", http.StatusBadRequest)
		return
	}
	_, err = db.Exec("UPDATE referral_requests SET status = $1 WHERE id = $2", "Denied", referralRequestID)
	if err != nil {
		log.Println("Error denying referral request:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func CreateReferralRequestHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user User
	err = db.QueryRow(`
		SELECT u.id, u.email, u.username, u.password, u.role, u.company_id, c.name 
		FROM users u 
		LEFT JOIN companies c ON u.company_id = c.id 
		INNER JOIN sessions s ON u.id = s.user_id 
		WHERE s.session_id = $1
	`, sessionID.Value).Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Role, &user.CompanyID, &user.CompanyName)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var referralRequest ReferralRequest
	if err := json.NewDecoder(r.Body).Decode(&referralRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	referralRequest.Username = user.Username
	referralRequest.ReferrerUserID = user.ID
	referralRequest.CompanyID = user.CompanyID

	if referralRequest.Title == "" || referralRequest.Content == "" || referralRequest.RefereeClient == "" {
		http.Error(w, "Invalid input fields", http.StatusBadRequest)
		return
	}

	// Validate that sent_company_id exists in the companies table
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM companies WHERE id = $1", referralRequest.SentCompanyID).Scan(&count)
	if err != nil {
		log.Println("Error checking sent_company_id existence:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "Invalid sent company ID", http.StatusBadRequest)
		return
	}

	// Validate that receiving_company_id exists in the companies table
	err = db.QueryRow("SELECT COUNT(*) FROM companies WHERE id = $1", referralRequest.ReceivingCompanyID).Scan(&count)
	if err != nil {
		log.Println("Error checking receiving_company_id existence:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "Invalid receiving company ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`
		INSERT INTO referral_requests (username, title, content, referrer_user_id, company_id, sent_company_id, receiving_company_id, referee_client, referee_client_email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, referralRequest.Username, referralRequest.Title, referralRequest.Content, referralRequest.ReferrerUserID, referralRequest.CompanyID, referralRequest.SentCompanyID, referralRequest.ReceivingCompanyID, referralRequest.RefereeClient, referralRequest.RefereeClientEmail)
	if err != nil {
		log.Println("Error creating referral request:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func GetReferralsSentHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user User
	err = db.QueryRow(`
		SELECT u.id, u.email, u.username, u.password, u.role, u.company_id 
		FROM users u 
		INNER JOIN sessions s ON u.id = s.user_id 
		WHERE s.session_id = $1
	`, sessionID.Value).Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Role, &user.CompanyID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var rows *sql.Rows
	var query string

	if user.Role == "superAdmin" || user.Role == "platformAdmin" {
		query = `
			SELECT r.id, r.title, r.content, r.username, r.referrer_user_id, r.company_id, r.sent_company_id, r.referee_client, r.referee_client_email, r.created_at, r.status, c.name AS company_name, sc.name AS sent_company_name 
			FROM referral_requests r 
			LEFT JOIN companies c ON r.company_id = c.id 
			LEFT JOIN companies sc ON r.sent_company_id = sc.id 
			ORDER BY r.created_at DESC
		`
		rows, err = db.Query(query)
	} else if user.Role == "companyAdmin" {
		query = `
			SELECT r.id, r.title, r.content, r.username AS referrer_username, r.referrer_user_id, r.company_id, r.sent_company_id, r.referee_client, r.referee_client_email, r.created_at, r.status, c.name AS company_name, sc.name AS sent_company_name 
			FROM referral_requests r 
			LEFT JOIN companies c ON r.company_id = c.id 
			LEFT JOIN companies sc ON r.sent_company_id = sc.id 
			WHERE r.sent_company_id = $1 
			ORDER BY r.created_at DESC
		`
		rows, err = db.Query(query, user.CompanyID)
	} else {
		query = `
			SELECT r.id, r.title, r.content, r.username, r.referrer_user_id, r.company_id, r.sent_company_id, r.referee_client, r.referee_client_email, r.created_at, r.status, c.name AS company_name, sc.name AS sent_company_name 
			FROM referral_requests r 
			LEFT JOIN companies c ON r.company_id = c.id 
			LEFT JOIN companies sc ON r.sent_company_id = sc.id 
			WHERE r.referrer_user_id = $1 
			ORDER BY r.created_at DESC
		`
		rows, err = db.Query(query, user.ID)
	}

	if err != nil {
		log.Println("Error fetching referral requests:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	var referralRequests []ReferralRequest
	for rows.Next() {
		var referralRequest ReferralRequest
		err := rows.Scan(&referralRequest.ID, &referralRequest.Title, &referralRequest.Content, &referralRequest.ReferrerUsername,
			&referralRequest.ReferrerUserID, &referralRequest.CompanyID, &referralRequest.SentCompanyID, &referralRequest.RefereeClient, &referralRequest.RefereeClientEmail,
			&referralRequest.CreatedAt, &referralRequest.Status, &referralRequest.CompanyName, &referralRequest.SentCompanyName)
		if err != nil {
			log.Println("Error scanning referral request row:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		referralRequests = append(referralRequests, referralRequest)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over referral request rows:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(referralRequests) == 0 {
		referralRequests = []ReferralRequest{} // Ensure an empty slice is returned if no records found
	}
	json.NewEncoder(w).Encode(referralRequests)
}

func GetReferralsReceivedHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user User
	err = db.QueryRow(`
        SELECT u.id, u.email, u.username, u.password, u.role, u.company_id 
        FROM users u 
        INNER JOIN sessions s ON u.id = s.user_id 
        WHERE s.session_id = $1
    `, sessionID.Value).Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Role, &user.CompanyID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var rows *sql.Rows
	var query string

	if user.Role == "superAdmin" || user.Role == "platformAdmin" {
		query = `
            SELECT r.id, r.title, r.content, r.username AS referrer_username, r.referrer_user_id, r.company_id, r.sent_company_id, r.referee_client, r.referee_client_email, r.created_at, r.status, c.name AS company_name, sc.name AS sent_company_name 
            FROM referral_requests r 
            LEFT JOIN companies c ON r.company_id = c.id 
            LEFT JOIN companies sc ON r.sent_company_id = sc.id 
            ORDER BY r.created_at DESC
        `
		rows, err = db.Query(query)
	} else if user.Role == "companyAdmin" || user.Role == "user" {
		query = `
            SELECT r.id, r.title, r.content, r.username AS referrer_username, r.referrer_user_id, r.company_id, r.sent_company_id, r.referee_client, r.referee_client_email, r.created_at, r.status, c.name AS company_name, sc.name AS sent_company_name 
            FROM referral_requests r 
            LEFT JOIN companies c ON r.company_id = c.id 
            LEFT JOIN companies sc ON r.sent_company_id = sc.id 
            WHERE r.receiving_company_id = $1 
            ORDER BY r.created_at DESC
        `
		rows, err = db.Query(query, user.CompanyID)
	}
	// } else {
	//     query = `
	//         SELECT r.id, r.title, r.content, r.username AS referrer_username, r.referrer_user_id, r.company_id, r.sent_company_id, r.referee_client, r.referee_client_email, r.created_at, r.status, c.name AS company_name, sc.name AS sent_company_name
	//         FROM referral_requests r
	//         LEFT JOIN companies c ON r.company_id = c.id
	//         LEFT JOIN companies sc ON r.sent_company_id = sc.id
	//         WHERE r.receiving_company_id = $1 OR r.referrer_user_id = $2
	//         ORDER BY r.created_at DESC
	//     `
	//     rows, err = db.Query(query, user.CompanyID, user.ID)
	// }

	if err != nil {
		log.Println("Error fetching referral requests:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	var referralRequests []ReferralRequest
	for rows.Next() {
		var referralRequest ReferralRequest
		err := rows.Scan(&referralRequest.ID, &referralRequest.Title, &referralRequest.Content, &referralRequest.ReferrerUsername,
			&referralRequest.ReferrerUserID, &referralRequest.CompanyID, &referralRequest.SentCompanyID, &referralRequest.RefereeClient, &referralRequest.RefereeClientEmail,
			&referralRequest.CreatedAt, &referralRequest.Status, &referralRequest.CompanyName, &referralRequest.SentCompanyName)
		if err != nil {
			log.Println("Error scanning referral request row:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		referralRequests = append(referralRequests, referralRequest)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over referral request rows:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(referralRequests) == 0 {
		referralRequests = []ReferralRequest{} // Ensure an empty slice is returned if no records found
	}
	json.NewEncoder(w).Encode(referralRequests)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log the incoming user data
	log.Printf("Received user data: %+v", user)

	// Validate companyID for companyAdmin role
	if user.Role == "companyAdmin" && user.CompanyID == 0 {
		http.Error(w, "Company ID is required for companyAdmin", http.StatusBadRequest)
		return
	}

	// Check if the company exists and handle company creation if necessary
	var companyID int
	if user.CompanyID != 0 {
		err := db.QueryRow("SELECT id FROM companies WHERE id = $1", user.CompanyID).Scan(&companyID)
		if err == sql.ErrNoRows {
			// Company does not exist, insert it
			log.Printf("Company ID %d not found, inserting new company: %s", user.CompanyID, user.CompanyName)
			err = db.QueryRow("INSERT INTO companies (name) VALUES ($1) RETURNING id", user.CompanyName).Scan(&companyID)
			if err != nil {
				log.Println("Error inserting company:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			user.CompanyID = companyID
		} else if err != nil {
			log.Println("Error querying company:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		// If no company ID is provided for non-companyAdmin roles, log and return an error
		if user.Role != "companyAdmin" {
			log.Println("Company ID is required but not provided.")
			http.Error(w, "Company ID is required", http.StatusBadRequest)
			return
		}
	}

	// Now insert the user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (email, username, password, role, company_id) VALUES ($1, $2, $3, $4, $5)",
		user.Email, user.Username, hashedPassword, user.Role, user.CompanyID)
	if err != nil {
		log.Printf("Error inserting user with email %s: %v", user.Email, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Parse user details from request body
	var updatedUser User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Only hash the password if it is updated
	if updatedUser.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedUser.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}
		updatedUser.Password = string(hashedPassword)
	}

	// Perform the update in the database
	_, err := db.Exec("UPDATE users SET email = $1, username = $2, password = $3, role = $4, company_id = $5 WHERE id = $6",
		updatedUser.Email, updatedUser.Username, updatedUser.Password, updatedUser.Role, updatedUser.CompanyID, updatedUser.ID)
	if err != nil {
		log.Println("Error updating user:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Println("Error decoding JSON:", err)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println("Error beginning transaction:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Delete dependent rows from referral_requests
	_, err = tx.Exec("DELETE FROM referral_requests WHERE referrer_user_id = $1", request.UserID)
	if err != nil {
		tx.Rollback()
		log.Println("Error deleting referral requests:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Delete dependent rows from sessions
	_, err = tx.Exec("DELETE FROM sessions WHERE user_id = $1", request.UserID)
	if err != nil {
		tx.Rollback()
		log.Println("Error deleting sessions:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Delete user
	_, err = tx.Exec("DELETE FROM users WHERE id = $1", request.UserID)
	if err != nil {
		tx.Rollback()
		log.Println("Error deleting user:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Println("User deleted successfully:", request.UserID)
}

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query(`SELECT u.id, u.email, u.username, u.role, u.company_id, c.name AS company_name
                           FROM users u
                           LEFT JOIN companies c ON u.company_id = c.id`)
	if err != nil {
		log.Println("Error fetching users:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.Username, &user.Role, &user.CompanyID, &user.CompanyName)
		if err != nil {
			log.Println("Error scanning user row:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over user rows:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(users) == 0 {
		users = []User{} // Ensure an empty slice is returned if no records found
	}
	json.NewEncoder(w).Encode(users)
}

func GetUsersByCompanyHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user User
	err = db.QueryRow("SELECT u.id, u.email, u.username, u.password, u.role, u.company_id, c.name FROM users u "+
		"LEFT JOIN companies c ON u.company_id = c.id "+
		"INNER JOIN sessions s ON u.id = s.user_id "+
		"WHERE s.session_id = $1", sessionID.Value).
		Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.Role, &user.CompanyID, &user.CompanyName)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	companyID, err := strconv.Atoi(vars["companyID"])
	if err != nil {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	var companyUsers []User
	rows, err := db.Query("SELECT id, email, username, role, company_id FROM users WHERE company_id = $1", companyID)
	if err != nil {
		log.Println("Error fetching users:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var companyUser User
		err := rows.Scan(&companyUser.ID, &companyUser.Email, &companyUser.Username, &companyUser.Role, &companyUser.CompanyID)
		if err != nil {
			log.Println("Error scanning user row:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		companyUsers = append(companyUsers, companyUser)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over user rows:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(companyUsers) == 0 {
		companyUsers = []User{} // Ensure an empty slice is returned if no records found
	}
	json.NewEncoder(w).Encode(companyUsers)
}

func CreateCompanyHandler(w http.ResponseWriter, r *http.Request) {
	var company Company
	if err := json.NewDecoder(r.Body).Decode(&company); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := db.Exec("INSERT INTO companies (name) VALUES ($1)", company.Name)
	if err != nil {
		log.Println("Error inserting company:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func DeleteCompanyHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		CompanyID int `json:"company_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Println("Error decoding JSON:", err)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println("Error beginning transaction:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Delete dependent rows from referral_requests
	_, err = tx.Exec("DELETE FROM referral_requests WHERE company_id = $1", request.CompanyID)
	if err != nil {
		tx.Rollback()
		log.Println("Error deleting referral requests:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Delete dependent rows from users
	_, err = tx.Exec("DELETE FROM users WHERE company_id = $1", request.CompanyID)
	if err != nil {
		tx.Rollback()
		log.Println("Error deleting users:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Delete the company
	result, err := tx.Exec("DELETE FROM companies WHERE id = $1", request.CompanyID)
	if err != nil {
		tx.Rollback()
		log.Println("Error deleting company:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		log.Println("Error fetching rows affected:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		tx.Rollback()
		http.Error(w, "Company not found", http.StatusNotFound)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Println("Company deleted successfully:", request.CompanyID)
}
