package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	tele "gopkg.in/tucnak/telebot.v2"
)

var (
	//	uid      string
	//	db       *sql.DB
	user     string
	password string
)

const (
	server                  = "dlietnikov-aks-demo.database.windows.net"
	port                    = "1433"
	database                = "aks-demo"
	telegramTokenSecretFile = "/mnt/secrets-store/aks-demo-kv-tg-token"
	userSecretFile          = "/mnt/secrets-store/aks-demo-kv-user"
	passwordSecretFile      = "/mnt/secrets-store/aks-demo-kv-password"
)

func main() {
	// Read user and password from files
	var err error
	user, err = readSecretFromFile(userSecretFile)
	if err != nil {
		log.Fatal("Failed to read user from file:", err)
	}
	password, err = readSecretFromFile(passwordSecretFile)
	if err != nil {
		log.Fatal("Failed to read password from file:", err)
	}

	// Print debug information
	log.Printf("Server: %s", server)
	log.Printf("Port: %s", port)
	log.Printf("Database: %s", database)
	log.Printf("User: %s", user)

	// Establish connection to the database
	connString := fmt.Sprintf("server=%s;port=%s;database=%s;user id=%s;password=%s",
		server, port, database, user, password)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Failed to open database connection:", err)
	}
	defer db.Close()

	// Create table if it doesn't exist
	createTable(db)

	telegramToken, err := readSecretFromFile(telegramTokenSecretFile)
	if err != nil {
		log.Fatal("Failed to read Telegram token from file:", err)
	}

	bot, err := tele.NewBot(tele.Settings{
		Token:  telegramToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal("Failed to create Telegram bot:", err)
	}

	// Initialize chatID with an invalid value
	var chatID int64 = -1

	// Get the chatID asynchronously, and continue running the bot even if it fails
	go func() {
		var err error
		chatID, err = getChatID(bot)
		if err != nil {
			log.Println("Failed to get chat_id. The bot will wait for the first message in the chat.")
		}
	}()

	sendStatistics(db, bot, chatID)

	bot.Start()
}

func sendStatistics(db *sql.DB, bot *tele.Bot, chatID int64) {
	for {
		// Get the last processed ID from the database
		lastProcessedID, err := getLastProcessedID(db)
		if err != nil {
			log.Println("Error getting last processed ID:", err)
			time.Sleep(time.Hour)
			continue
		}

		// Query the database for new data since the last processed ID
		query := fmt.Sprintf("SELECT ID, Hour, UID, Count FROM JobTable WHERE ID > %d", lastProcessedID)
		rows, err := db.Query(query)
		if err != nil {
			log.Println("Error executing query:", err)
			time.Sleep(time.Hour)
			continue
		}

		// Check if there are new records to send
		newDataAvailable := false
		for rows.Next() {
			newDataAvailable = true
			var id int
			var hour time.Time
			var uid string
			var count int

			err := rows.Scan(&id, &hour, &uid, &count)
			if err != nil {
				log.Println("Error scanning query result:", err)
				continue
			}

			// Prepare the message to be sent
			message := fmt.Sprintf("ID: %d\nHour: %s\nUID: %s\nCount: %d", id, hour.Format(time.RFC3339), uid, count)

			// Send the message to the specified chatID
			sendTelegramMessage(bot, chatID, message)

			// Update the last processed ID in the database
			err = updateLastProcessedID(db, id)
			if err != nil {
				log.Println("Error updating last processed ID:", err)
			}
		}

		rows.Close()

		// If no new data is available, send a message indicating that
		if !newDataAvailable {
			message := "No new data available in the last hour."
			sendTelegramMessage(bot, chatID, message)
		}

		// Sleep for an hour before checking for new data again
		time.Sleep(time.Hour)
	}
}

func sendTelegramMessage(bot *tele.Bot, chatID int64, message string) {
	// Only send the message if chatID is valid
	if chatID != -1 {
		_, err := bot.Send(&tele.Chat{ID: chatID}, message)
		if err != nil {
			log.Println("Error sending message to Telegram:", err)
		}
	}
}

func getChatID(bot *tele.Bot) (int64, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", bot.Token))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool `json:"ok"`
		Result []struct {
			Message struct {
				Chat struct {
					ID int64 `json:"id"`
				} `json:"chat"`
			} `json:"message"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, err
	}

	if !result.OK || len(result.Result) == 0 {
		// Send an error message, but don't terminate the program
		log.Println("Failed to get chat_id. The bot will wait for the first message in the chat.")
		return 0, nil
	}

	chatID := result.Result[0].Message.Chat.ID
	return chatID, nil
}

func readSecretFromFile(filePath string) (string, error) {
	secretBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read secret file: %v", err)
	}

	secret := strings.TrimSpace(string(secretBytes))
	if secret == "" {
		return "", fmt.Errorf("secret is empty in file: %s", filePath)
	}

	return secret, nil
}

func getLastProcessedID(db *sql.DB) (int, error) {
	query := "SELECT ISNULL(MAX(ID), 0) FROM LastProcessedRecord"

	var lastProcessedID int
	err := db.QueryRow(query).Scan(&lastProcessedID)
	if err != nil {
		return 0, fmt.Errorf("failed to get last processed ID: %v", err)
	}

	return lastProcessedID, nil
}

func updateLastProcessedID(db *sql.DB, id int) error {
	query := "UPDATE LastProcessedRecord SET ID = @ID"

	_, err := db.Exec(query, sql.Named("ID", id))
	if err != nil {
		return fmt.Errorf("failed to update last processed ID: %v", err)
	}

	return nil
}

func createTable(db *sql.DB) {
	createTableQuery := `
		IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'LastProcessedRecord') AND type in (N'U'))
		BEGIN
			CREATE TABLE LastProcessedRecord (
				ID INT NOT NULL
			)
		END`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("LastProcessedRecord table created (if not already exists)")
}
