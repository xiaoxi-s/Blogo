package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Username string             `json:"username" bson:"username"`
	Password string             `json:"password" bson:"password"`
	UserID   primitive.ObjectID `json:"userID" bson:"_id"`
}

type UserProfile struct {
	Username    string             `json:"username" bson:"username"`
	Password    string             `json:"password" bson:"password"`
	UserID      primitive.ObjectID `json:"userID" bson:"_id"`
	CreatedTime time.Time          `json:"userCreatedTime" bson:"userCreatedTime"`
}
