package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

var (
	uid      string
	db       *sql.DB
	user     string
	password string
)

const (
	server   = "dlietnikov-aks-demo.database.windows.net"
	port     = "1433"
	database = "aks-demo"

	userSecretFile     = "/mnt/secrets-store/aks-demo-kv-user"
	passwordSecretFile = "/mnt/secrets-store/aks-demo-kv-password"
)

func main() {
	// Generate UID
	uid = generateUID()

	// Read user and password from files
	var err error
	user, err = readSecretFromFile(userSecretFile)
	if err != nil {
		log.Fatal(err)
	}
	password, err = readSecretFromFile(passwordSecretFile)
	if err != nil {
		log.Fatal(err)
	}

	// Print debug information
	log.Printf("Server: %s", server)
	log.Printf("Port: %s", port)
	log.Printf("Database: %s", database)
	log.Printf("User: %s", user)
	//log.Printf("Password: %s", password)

	// Establish connection to the database
	db, err = sql.Open("sqlserver", getConnectionString())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table if it doesn't exist
	createTable(db)

	// Start server
	http.HandleFunc("/", handleRequest)

	log.Println("Server started on port 8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func createTable(db *sql.DB) {
	createTableQuery := `
		IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'LogTable') AND type in (N'U'))
		BEGIN
			CREATE TABLE LogTable (
				ID INT IDENTITY(1,1) PRIMARY KEY,
				UID VARCHAR(36),
				Time DATETIME
			)
		END`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("LogTable table created (if not already exists)")
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Insert a record into LogTable
	insertLog()

	//  Send UID and current time to the browser
	currentTime := time.Now().Format(time.RFC3339)
	fmt.Fprintf(w, "UID: %s\nRequest Time: %s", uid, currentTime)
}

func insertLog() {
	insertQuery := "INSERT INTO LogTable (UID, Time) VALUES (@UID, @Time)"

	_, err := db.Exec(insertQuery, sql.Named("UID", uid), sql.Named("Time", time.Now()))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Record added to LogTable")
}

func generateUID() string {
	// Generate random numbers
	numbers := make([]int, 4)
	for i := 0; i < 4; i++ {
		numbers[i] = rand.Intn(10000)
	}

	// Generate UID with format "2342-8285-6205-6921"
	uidParts := make([]string, 4)
	for i := 0; i < 4; i++ {
		uidParts[i] = strconv.Itoa(numbers[i])
	}
	uid := strings.Join(uidParts, "-")

	return uid
}

func readSecretFromFile(filePath string) (string, error) {
	secretBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("Failed to read secret file: %v", err)
	}

	secret := strings.TrimSpace(string(secretBytes))
	if secret == "" {
		return "", fmt.Errorf("Secret is empty in file: %s", filePath)
	}

	return secret, nil
}

func getConnectionString() string {
	return fmt.Sprintf("server=%s;port=%s;database=%s;user id=%s;password=%s",
		server, port, database, user, password)
}

//
