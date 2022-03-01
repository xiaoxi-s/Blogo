package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	PostID          primitive.ObjectID `json:"postID" bson:"postID"`
	Username        string             `json:"username" bson:"username"`
	Title           string             `json:"postTitle" bson:"postTitle"`
	Tags            []string           `json:"postTags" bson:"postTags"`
	CreatedTime     time.Time          `json:"postCreatedTime" bson:"postCreatedTime"`
	LastUpdatedTime time.Time          `json:"postLastUpdatedTime" bson:"postLastUpdatedTime"`
	NumOfThumb      int64              `json:"postNumOfThumb" bson:"postNumOfThumb"`
	Content         string             `json:"postContent" bson:"postContent"`
}
