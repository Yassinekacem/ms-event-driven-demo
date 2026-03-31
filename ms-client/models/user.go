package models

type User struct {
    ID         int    `json:"id"`
    Nom        string `json:"nom"`
    Prenom     string `json:"prenom"`
    Email      string `json:"email"`
    Statut     string `json:"statut"`
    MotDePasse string `json:"mot_de_passe,omitempty"`
    CreatedAt  string `json:"created_at,omitempty"`
}

type UserUpdate struct {
    ID     int    `json:"id"`
    Statut string `json:"statut"`
}

type StatusChangeEvent struct {
    UserID    int    `json:"user_id"`
    Email     string `json:"email"`
    OldStatus string `json:"old_status"`
    NewStatus string `json:"new_status"`
    Timestamp string `json:"timestamp"`
}