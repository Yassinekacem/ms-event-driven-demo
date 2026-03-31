package models

type Notification struct {
    ID         int    `json:"id"`
    UserID     int    `json:"user_id"`
    UserEmail  string `json:"user_email"`
    EventType  string `json:"event_type"`
    OldStatus  string `json:"old_status"`
    NewStatus  string `json:"new_status"`
    Message    string `json:"message"`
    CreatedAt  string `json:"created_at"`
    IsRead     bool   `json:"is_read"`
}

type StatusChangeEvent struct {
    UserID    int    `json:"user_id"`
    Email     string `json:"email"`
    OldStatus string `json:"old_status"`
    NewStatus string `json:"new_status"`
    Timestamp string `json:"timestamp"`
}