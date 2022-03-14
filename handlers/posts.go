package handlers

import (
	"blogo/models"
	"encoding/json"
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

type PostsHandler struct {
	ctx         context.Context
	collection  *mongo.Collection
	redisClient *redis.Client
}

func NewPostsHandlers(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *PostsHandler {
	return &PostsHandler{
		ctx:         ctx,
		collection:  collection,
		redisClient: redisClient,
	}
}

// swagger:operation GET /posts post listPosts
// Return lists of posts
// ---
// produces:
// - application/json
// responses:
//  '200':
//   description: Successful operation
func (handler *PostsHandler) ListPostsHandler(c *gin.Context) {
	val, err := handler.redisClient.Get("posts").Result()
	if err == redis.Nil {
		log.Printf("Request to MongoDB")
		cur, err := handler.collection.Find(handler.ctx, bson.M{})
		if err != nil {
			log.Printf("Request to Mongo Failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cur.Close(handler.ctx)

		posts := make([]models.Post, 0)
		for cur.Next(handler.ctx) {
			var post models.Post
			cur.Decode(&post)
			posts = append(posts, post)
		}

		data, _ := json.Marshal(posts)
		handler.redisClient.Set("posts_in_redis", string(data), 0)
		c.JSON(http.StatusOK, posts)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		log.Printf("Request to redis")
		posts := make([]models.Post, 0)
		json.Unmarshal([]byte(val), &posts)
		c.JSON(http.StatusOK, posts)
	}
}

// swagger:operation POST /posts post newPost
// Create a new post
// ---
// produces:
// - application/json
// responses:
//  '200':
//   description: Successful operation
//  '500':
//   description: Decode input post error or insertion error
func (handler *PostsHandler) NewPostHandler(c *gin.Context) {
	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	post.NumOfThumb = 0
	post.PostID = primitive.NewObjectID()
	post.CreatedTime = time.Now()
	post.LastUpdatedTime = post.CreatedTime

	_, err := handler.collection.InsertOne(handler.ctx, post)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("Delete redis cache")
	handler.redisClient.Del("posts_in_redis")
	c.JSON(http.StatusOK, post)
}

// swagger:operation GET /random-post post getOneRandomPost
// Return a post sampled randomly
// ---
// produces:
// - application/json
// responses:
//  '200':
//   description: Successful operation
func (handler *PostsHandler) GetOneRandomPost(c *gin.Context) {
	// retrieve parameter id and search in database
	pipeline := []bson.D{{{"$sample", bson.D{{"size", 1}}}}}
	// TODO: use redis!

	cur, err := handler.collection.Aggregate(handler.ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	posts := make([]models.Post, 0)
	for cur.Next(handler.ctx) {
		var post models.Post
		cur.Decode(&post)
		posts = append(posts, post)
	}
	if len(posts) == 0 {
		c.JSON(http.StatusOK, "")
	} else {
		c.JSON(http.StatusOK, posts[0])
	}
}

// swagger:operation GET /posts/{id} post viewPost
// View a post given its id
// ---
// produces:
// - application/json
// responses:
//  '200':
//   description: Successful operation
//  '400':
//   description: Invalid post ID
//  '404':
//   description: post with provided ID not found
func (handler *PostsHandler) ViewPostHandler(c *gin.Context) {
	// retrieve parameter id and search in database
	postIDString := c.Param("id")
	postID, err := primitive.ObjectIDFromHex(postIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: use redis!

	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": postID,
	})
	if cur.Err() != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": cur.Err().Error()})
		return
	}

	// decode the post into go struct
	var post models.Post
	err = cur.Decode(&post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

// swagger:operation DELETE /posts/{id} post deletePost
// Delete a post given its ID
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: ID of the post
//     required: true
//     type: string
// responses:
//   '200':
//     description: Successful operation
//   '404':
//     description: Invalid post ID
func (handler *PostsHandler) DeletePostHandler(c *gin.Context) {
	id := c.Param("id")
	objectid, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectid,
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}

	handler.redisClient.Del("posts_in_redis")
	c.JSON(http.StatusOK, gin.H{"deleteResult": "success"})
}

// swagger:operation GET /post/search/{title} post searchPost
// Search a post given its title
// ---
// produces:
// - application/json
// parameters:
//   - name: title
//     in: path
//     description: post title
//     required: true
//     type: string
// responses:
//   '200':
//     description: Successful operation
//   '404':
//     description: Invalid post title
func (handler *PostsHandler) SearchPostHandler(c *gin.Context) {
	title := c.Param("title")

	// TODO: use redis!

	cur, err := handler.collection.Find(handler.ctx, bson.M{
		"postTitle": title,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	posts := make([]models.Post, 0)
	for cur.Next(handler.ctx) {
		var post models.Post
		cur.Decode(&post)
		posts = append(posts, post)
	}

	c.JSON(http.StatusOK, posts)
}

// swagger:operation POST /post/thumbup/{id} post thumbupPost
// Give a post specified by id a thumb up
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: post id
//     required: true
//     type: string
// responses:
//   '200':
//     description: Successful operation
//   '404':
//     description: Invalid post id
func (handler *PostsHandler) ThumbupPostHandler(c *gin.Context) {

	id := c.Param("id")
	objectid, _ := primitive.ObjectIDFromHex(id)

	// find the comment
	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectid,
	})
	if cur.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": cur.Err().Error()})
		return
	}

	// modify the post
	var post models.Post
	err := cur.Decode(&post)
	if err != nil { // decode error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	num := post.NumOfThumb
	_, err = handler.collection.UpdateByID(handler.ctx, objectid, bson.M{
		"$set": bson.D{{"postNumOfThumb", num + 1}},
	})

	if err != nil { // update error
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"thumbupResult": "success"})
}
