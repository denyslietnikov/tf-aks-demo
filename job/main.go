package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

var (
	uid      string
	db       *sql.DB
	user     string
	password string
	server   string
	port     int
	database string
)

const (
	//	server   = "dlietnikov-aks-demo.database.windows.net"
	//	port     = "1433"
	//	database = "aks-demo"
	serverSecretFile   = "/mnt/secrets-store/aks-demo-kv-server"
	portSecretFile     = "/mnt/secrets-store/aks-demo-kv-port"
	databaseSecretFile = "/mnt/secrets-store/aks-demo-kv-database"
	userSecretFile     = "/mnt/secrets-store/aks-demo-kv-user"
	passwordSecretFile = "/mnt/secrets-store/aks-demo-kv-password"
)

func main() {

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

	server, err = readSecretFromFile(serverSecretFile)
	if err != nil {
		log.Fatal(err)
	}
	portStr, err := readSecretFromFile(portSecretFile)
	if err != nil {
		log.Fatal(err)
	}
	port, err = strconv.Atoi(portStr)
	if err != nil {
		log.Fatal(err)
	}

	database, err = readSecretFromFile(databaseSecretFile)
	if err != nil {
		log.Fatal(err)
	}

	// Establish connection to the database
	connString := getConnectionString()
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the JobTable if it doesn't exist
	createTable(db)

	// Start the job every hour
	startJob(db)

	// Wait indefinitely
	select {}
}

func createTable(db *sql.DB) {
	createTableQuery := `
		IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'JobTable') AND type in (N'U'))
		BEGIN
			CREATE TABLE JobTable (
				ID INT IDENTITY(1,1) PRIMARY KEY,
				Hour DATETIME,
				UID VARCHAR(36),
				Count INT
			)
		END`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("JobTable table created (if not already exists)")
}

func startJob(db *sql.DB) {
	// Determine the next hour
	now := time.Now()
	nextHour := now.Truncate(time.Hour).Add(time.Hour)

	// Calculate the duration to the next hour
	duration := nextHour.Sub(now)

	log.Printf("Starting job in %s", duration)

	// Wait until the next hour
	time.Sleep(duration)

	// Start the job and repeat every hour
	for {
		go runJob(db)

		// Wait until the next hour
		time.Sleep(time.Hour)
	}
}

func runJob(db *sql.DB) {
	// Get the current hour and the previous hour
	now := time.Now()
	previousHour := now.Truncate(time.Hour).Add(-time.Hour)

	// Execute the query to count the records for each unique UID in the previous hour
	query := `
		INSERT INTO JobTable (Hour, UID, Count)
		SELECT @PreviousHour, UID, COUNT(*)
		FROM LogTable
		WHERE Time >= @PreviousHour AND Time < @CurrentHour
		GROUP BY UID`

	_, err := db.Exec(query, sql.Named("PreviousHour", previousHour), sql.Named("CurrentHour", now))
	if err != nil {
		log.Println("Error executing the job:", err)
		return
	}

	log.Println("Job executed successfully")
}

func readSecretFromFile(filePath string) (string, error) {
	secretBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("Failed to read secret file: %v", err)
	}

	secret := string(secretBytes)
	return secret, nil
}

func getConnectionString() string {
	return fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s",
		server, user, password, port, database)
}

//
