package handlers

import (
	"blogo/models"
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
	ctx         context.Context
	collection  *mongo.Collection
	redisClient *redis.Client
}

func NewCommentsHandlers(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *CommentsHandler {
	return &CommentsHandler{
		ctx:         ctx,
		collection:  collection,
		redisClient: redisClient,
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
	postID, err := primitive.ObjectIDFromHex(postIDString)
	log.Println(postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	cur, err := handler.collection.Find(handler.ctx, bson.M{
		"commentToID": postID,
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
	postID, err := primitive.ObjectIDFromHex(postIDString)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}

	comment.NumOfThumb = 0
	comment.CommentID = primitive.NewObjectID()
	comment.CommentToID = postID
	comment.CreatedTime = time.Now()

	// TODO: use redis

	_, err = handler.collection.InsertOne(handler.ctx, comment)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comment)
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
		"$set": bson.D{{"postNumOfThumb", num + 1}},
	})

	if err != nil { // update error
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"thumbupResult": "success"})
}
