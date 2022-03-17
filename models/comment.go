package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	CommentID   primitive.ObjectID `json:"commentID" bson:"_id,omitempty"`
	Username    string             `json:"username" bson:"username"`
	CommentToID string             `json:"commentToID" bson:"commentToID"`
	CreatedTime time.Time          `json:"commentCreatedTime" bson:"commentCreatedTime"`
	NumOfThumb  int64              `json:"numOfThumb" bson:"numOfThumb"`
	Content     string             `json:"commentContent" bson:"commentContent"`
}

type CommentThumbupedByUser struct {
	CommentID string `json:"commentID" bson:"_id,omitempty"`
	Username  string `json:"username" bson:"username"`
}
