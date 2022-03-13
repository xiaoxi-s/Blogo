package models

type Entry struct {
	Link        string `json:"url" bson:"url"`
	Description string `json:"description" bson:"description"`
	Title       string `json:"title" bson:"title"`
}
