package handlers

import (
	"blogo/models"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type CommentsHandler struct {
	ctx                       context.Context
	collection                *mongo.Collection
	collectionThumbupedByUser *mongo.Collection
	redisClient               *redis.Client
}

func NewCommentsHandlers(ctx context.Context, collection *mongo.Collection, collectionThumbupedByUser *mongo.Collection, redisClient *redis.Client) *CommentsHandler {
	return &CommentsHandler{
		ctx:                       ctx,
		collection:                collection,
		collectionThumbupedByUser: collectionThumbupedByUser,
		redisClient:               redisClient,
	}
}

// swagger:operation GET /comments/{postid} comment listCommentsToPosts
// List comments to a post
// ---
// produce:
// - application/json
// parameters:
//   - name: postid
//     in: path
//     description: ID of the post
//     required: true
//     type: string
// responses:
//   '200':
//     description: Success operation
//   '404':
//	   description: Invalid posts
func (handler *CommentsHandler) ListCommentsToPostHandler(c *gin.Context) {
	postIDString := c.Param("postid") // get post id

	cur, err := handler.collection.Find(handler.ctx, bson.M{
		"commentToID": postIDString,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	comments := make([]models.Comment, 0)
	for cur.Next(handler.ctx) {
		var comment models.Comment
		cur.Decode(&comment)
		comments = append(comments, comment)
	}
	c.JSON(http.StatusOK, comments)
}

// swagger:operation POST /comments/:postid comment createCommentToPost
// Create a comment to a post
// ---
// produce:
// - application/json
// responses:
//   '200':
//     description: Success operation
//   '404':
//	   description: Invalid posts
func (handler *CommentsHandler) CreateCommentToPostHandler(c *gin.Context) {
	var comment models.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	postIDString := c.Param("postid")

	// TODO: verify the postID is valid

	comment.NumOfThumb = 0
	comment.CommentID = primitive.NewObjectID()
	comment.CommentToID = postIDString
	comment.CreatedTime = time.Now()

	// TODO: use redis

	_, err := handler.collection.InsertOne(handler.ctx, comment)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comment)
}

// swagger:operation GET /comments/by/{username} comment GetListOfCommentsBy
// return a list of commentsID where the associated comments are by the given username
// ---
// produce:
// - application/json
// responses:
//   '200':
//     description: Success operation
//   '404':
//	   description: Invalid username
func (handler *CommentsHandler) GetListOfCommentsBy(c *gin.Context) {
	username := c.Param("username")

	cur, err := handler.collection.Find(handler.ctx, bson.M{
		"username": username,
	})

	if err != nil { // update error
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	commentsID := make([]string, 0)
	for cur.Next(handler.ctx) {
		var comment models.Comment
		cur.Decode(&comment)
		commentsID = append(commentsID, comment.CommentID.Hex())
	}
	c.JSON(http.StatusOK, commentsID)
}

// swagger:operation GET /comments/thumbupedby/{username} comment GetListOfCommentsBy
// return a list of commentsID where the associated comments are by the given username
// ---
// produce:
// - application/json
// responses:
//   '200':
//     description: Success operation
//   '404':
//	   description: Invalid username
func (handler *CommentsHandler) GetListOfThumbupedBy(c *gin.Context) {
	username := c.Param("username")
	fmt.Print(username)

	cur, err := handler.collectionThumbupedByUser.Find(handler.ctx, bson.M{
		"username": username,
	})

	if err != nil { // update error
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	commentsID := make([]string, 0)
	for cur.Next(handler.ctx) {
		var commentThumbed models.CommentThumbupedByUser
		cur.Decode(&commentThumbed)
		commentsID = append(commentsID, commentThumbed.CommentID)
	}
	c.JSON(http.StatusOK, commentsID)
}

// swagger:operation POST /comments/thumbup/{commentid} comment commentThumbup
// Create a comment to a post
// ---
// produce:
// - application/json
// responses:
//   '200':
//     description: Success operation
//   '404':
//	   description: Invalid posts
func (handler *CommentsHandler) CommentThumbupHandler(c *gin.Context) {
	commentIDString := c.Param("commentid")
	var user models.User
	c.ShouldBindJSON(&user) // to get access to username

	commentID, _ := primitive.ObjectIDFromHex(commentIDString)

	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": commentID,
	})
	log.Println(commentID)
	var comment models.Comment
	err := cur.Decode(&comment)
	if err != nil { // decode error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	num := comment.NumOfThumb
	_, err = handler.collection.UpdateByID(handler.ctx, commentID, bson.M{
		"$set": bson.D{{"numOfThumb", num + 1}},
	})

	if err != nil { // update error
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	var commentThumbupedByUser models.CommentThumbupedByUser
	commentThumbupedByUser.Username = user.Username
	commentThumbupedByUser.CommentID = commentIDString

	_, err = handler.collectionThumbupedByUser.InsertOne(handler.ctx, commentThumbupedByUser)

	c.JSON(http.StatusOK, gin.H{"thumbupResult": "success"})
}
