package main

import (
    "database/sql"
    "log"
    
    "ms-client/database"
    "ms-client/events"
    "ms-client/models"
    
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    database.ConnectDB()
    defer database.DB.Close()
    
    err := events.ConnectRabbitMQ()
    if err != nil {
        log.Println("Warning: RabbitMQ not available:", err)
    }
    defer events.CloseRabbitMQ()
    
    app := fiber.New()
    
    app.Use(logger.New())
    app.Use(cors.New())
    
    app.Post("/users", createUser)
    app.Get("/users", getUsers)
    app.Get("/users/:id", getUserByID)
    app.Put("/users/:id/status", updateUserStatus)
    
    log.Fatal(app.Listen(":3000"))
}

func createUser(c *fiber.Ctx) error {
    var user models.User
    
    if err := c.BodyParser(&user); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.MotDePasse), bcrypt.DefaultCost)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to hash password",
        })
    }
    
    if user.Statut == "" {
        user.Statut = "user"
    }
    
    query := `INSERT INTO users (nom, prenom, email, statut, mot_de_passe) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
    
    var id int
    err = database.DB.QueryRow(query, user.Nom, user.Prenom, user.Email, user.Statut, string(hashedPassword)).Scan(&id)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to create user: " + err.Error(),
        })
    }
    
    user.ID = id
    user.MotDePasse = ""
    
    return c.Status(fiber.StatusCreated).JSON(user)
}

func getUsers(c *fiber.Ctx) error {
    rows, err := database.DB.Query("SELECT id, nom, prenom, email, statut, created_at FROM users")
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch users",
        })
    }
    defer rows.Close()
    
    var users []models.User
    for rows.Next() {
        var user models.User
        err := rows.Scan(&user.ID, &user.Nom, &user.Prenom, &user.Email, &user.Statut, &user.CreatedAt)
        if err != nil {
            continue
        }
        users = append(users, user)
    }
    
    return c.JSON(users)
}

func getUserByID(c *fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid user ID",
        })
    }
    
    var user models.User
    query := "SELECT id, nom, prenom, email, statut, created_at FROM users WHERE id = $1"
    err = database.DB.QueryRow(query, id).Scan(&user.ID, &user.Nom, &user.Prenom, &user.Email, &user.Statut, &user.CreatedAt)
    
    if err == sql.ErrNoRows {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "User not found",
        })
    }
    
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch user",
        })
    }
    
    return c.JSON(user)
}

func updateUserStatus(c *fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid user ID",
        })
    }
    
    var updateData models.UserUpdate
    if err := c.BodyParser(&updateData); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    var currentUser models.User
    query := "SELECT id, email, statut FROM users WHERE id = $1"
    err = database.DB.QueryRow(query, id).Scan(&currentUser.ID, &currentUser.Email, &currentUser.Statut)
    
    if err == sql.ErrNoRows {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "User not found",
        })
    }
    
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch user",
        })
    }
    
    oldStatus := currentUser.Statut
    
    updateQuery := "UPDATE users SET statut = $1 WHERE id = $2 RETURNING id, nom, prenom, email, statut, created_at"
    var updatedUser models.User
    err = database.DB.QueryRow(updateQuery, updateData.Statut, id).Scan(
        &updatedUser.ID, &updatedUser.Nom, &updatedUser.Prenom, &updatedUser.Email, &updatedUser.Statut, &updatedUser.CreatedAt,
    )
    
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to update user status",
        })
    }
    
    if oldStatus == "user" && updateData.Statut == "admin" {
        event := models.StatusChangeEvent{
            UserID:    updatedUser.ID,
            Email:     updatedUser.Email,
            OldStatus: oldStatus,
            NewStatus: updateData.Statut,
        }
        
        err = events.PublishStatusChangeEvent(event)
        if err != nil {
            log.Printf("Failed to publish event: %v", err)
        }
    }
    
    return c.JSON(updatedUser)
}