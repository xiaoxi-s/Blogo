package handlers

import (
	"blogo/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type NewsHandler struct {
	ctx         context.Context
	collection  *mongo.Collection
	redisClient *redis.Client
}

func NewNewsHandlers(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *NewsHandler {
	return &NewsHandler{
		ctx:         ctx,
		collection:  collection,
		redisClient: redisClient,
	}
}

// swagger:operation GET /daily daily signIn
// List news
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: Success
//   '500':
//     description: Database query error
func (handler *NewsHandler) ListNewsHandler(c *gin.Context) {
	cur, err := handler.collection.Find(handler.ctx, bson.D{{}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	newsSlice := make([]models.Entry, 0)
	for cur.Next(handler.ctx) {
		var news models.Entry
		cur.Decode(&news)
		newsSlice = append(newsSlice, news)
	}
	c.JSON(http.StatusOK, newsSlice)
}
