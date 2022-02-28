package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Username string             `json:"username"`
	Password string             `json:"password"`
	UserID   primitive.ObjectID `json:"userID"`
}

type UserProfile struct {
	Username    string             `json:"username"`
	Password    string             `json:"password"`
	UserID      primitive.ObjectID `json:"userID"`
	CreatedTime time.Time          `json:"userCreatedTime"`
}
