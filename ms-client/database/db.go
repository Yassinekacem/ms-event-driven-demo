package database

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
    dsn := "postgresql://neondb_owner:npg_HOyFYgMQ1ni9@ep-ancient-mud-an410t48-pooler.c-6.us-east-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require"
    
    var err error
    DB, err = sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    err = DB.Ping()
    if err != nil {
        log.Fatal("Failed to ping database:", err)
    }

    fmt.Println("Connected to Client Database successfully")
    
    createTableQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        nom VARCHAR(100) NOT NULL,
        prenom VARCHAR(100) NOT NULL,
        email VARCHAR(100) UNIQUE NOT NULL,
        statut VARCHAR(20) DEFAULT 'user',
        mot_de_passe VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`
    
    _, err = DB.Exec(createTableQuery)
    if err != nil {
        log.Fatal("Failed to create users table:", err)
    }
    
    fmt.Println("Users table ready")
}