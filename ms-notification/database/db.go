package database

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
    dsn := "postgresql://neondb_owner:npg_vbFYG9oir4mS@ep-hidden-art-a4gp3tew-pooler.us-east-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require"
    
    var err error
    DB, err = sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    err = DB.Ping()
    if err != nil {
        log.Fatal("Failed to ping database:", err)
    }

    fmt.Println("Connected to Notification Database successfully")
    
    createTableQuery := `
    CREATE TABLE IF NOT EXISTS notifications (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL,
        user_email VARCHAR(100) NOT NULL,
        event_type VARCHAR(50) NOT NULL,
        old_status VARCHAR(20),
        new_status VARCHAR(20),
        message TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        is_read BOOLEAN DEFAULT FALSE
    )`
    
    _, err = DB.Exec(createTableQuery)
    if err != nil {
        log.Fatal("Failed to create notifications table:", err)
    }
    
    fmt.Println("Notifications table ready")
}