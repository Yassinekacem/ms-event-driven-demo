package main

import (
    "database/sql"
    "log"
    
    "ms-notification/database"
    "ms-notification/events"
    "ms-notification/models"
    
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
    database.ConnectDB()
    defer database.DB.Close()
    
    err := events.ConnectRabbitMQ()
    if err != nil {
        log.Fatal("Failed to connect to RabbitMQ:", err)
    }
    defer events.CloseRabbitMQ()
    
    go events.StartConsumer()
    
    app := fiber.New()
    
    app.Use(logger.New())
    app.Use(cors.New())
    
    app.Get("/notifications", getNotifications)
    app.Get("/notifications/:id", getNotificationByID)
    app.Put("/notifications/:id/read", markAsRead)
    app.Get("/notifications/user/:userId", getNotificationsByUserID)
    
    log.Fatal(app.Listen(":3001"))
}

func getNotifications(c *fiber.Ctx) error {
    rows, err := database.DB.Query("SELECT id, user_id, user_email, event_type, old_status, new_status, message, created_at, is_read FROM notifications ORDER BY created_at DESC")
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch notifications",
        })
    }
    defer rows.Close()
    
    var notifications []models.Notification
    for rows.Next() {
        var notif models.Notification
        err := rows.Scan(&notif.ID, &notif.UserID, &notif.UserEmail, &notif.EventType, &notif.OldStatus, &notif.NewStatus, &notif.Message, &notif.CreatedAt, &notif.IsRead)
        if err != nil {
            continue
        }
        notifications = append(notifications, notif)
    }
    
    return c.JSON(notifications)
}

func getNotificationByID(c *fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid notification ID",
        })
    }
    
    var notif models.Notification
    query := "SELECT id, user_id, user_email, event_type, old_status, new_status, message, created_at, is_read FROM notifications WHERE id = $1"
    err = database.DB.QueryRow(query, id).Scan(&notif.ID, &notif.UserID, &notif.UserEmail, &notif.EventType, &notif.OldStatus, &notif.NewStatus, &notif.Message, &notif.CreatedAt, &notif.IsRead)
    
    if err == sql.ErrNoRows {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "Notification not found",
        })
    }
    
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch notification",
        })
    }
    
    return c.JSON(notif)
}

func markAsRead(c *fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid notification ID",
        })
    }
    
    result, err := database.DB.Exec("UPDATE notifications SET is_read = true WHERE id = $1", id)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to update notification",
        })
    }
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "Notification not found",
        })
    }
    
    return c.JSON(fiber.Map{
        "message": "Notification marked as read",
    })
}

func getNotificationsByUserID(c *fiber.Ctx) error {
    userID, err := c.ParamsInt("userId")
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid user ID",
        })
    }
    
    rows, err := database.DB.Query("SELECT id, user_id, user_email, event_type, old_status, new_status, message, created_at, is_read FROM notifications WHERE user_id = $1 ORDER BY created_at DESC", userID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch notifications",
        })
    }
    defer rows.Close()
    
    var notifications []models.Notification
    for rows.Next() {
        var notif models.Notification
        err := rows.Scan(&notif.ID, &notif.UserID, &notif.UserEmail, &notif.EventType, &notif.OldStatus, &notif.NewStatus, &notif.Message, &notif.CreatedAt, &notif.IsRead)
        if err != nil {
            continue
        }
        notifications = append(notifications, notif)
    }
    
    return c.JSON(notifications)
}