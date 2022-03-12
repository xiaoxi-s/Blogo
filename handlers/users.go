package handlers

import (
	"blogo/models"
	"context"
	"crypto/sha256"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	ctx        context.Context
	collection *mongo.Collection
}

func NewAuthHandler(ctx context.Context, collection *mongo.Collection) *AuthHandler {
	return &AuthHandler{
		ctx:        ctx,
		collection: collection,
	}
}

// swagger:operation POST /signin auth signIn
// Login with user name and password
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: Successful sign in
//   '401':
//     description: Invalid credentials
func (handler *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h := sha256.New()

	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"username": user.Username,
		"password": string(h.Sum([]byte(user.Password))),
	})

	if cur.Err() != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	sessionToken := xid.New().String()
	session := sessions.Default(c)
	session.Set("username", user.Username)
	session.Set("token", sessionToken)
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "sign in succeed", "cookie": sessionToken})
}

// swagger:operation POST /signout auth signOut
// Sign out
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: Successful sign out
func (handler *AuthHandler) SignOutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "signed out"})
}

// swagger:operation POST /signup auth signUp
// Sign up as a new user
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: Successful sign up
//   '400':
//     description: Username is used
//   '500':
//     description: Server databaes error
func (handler *AuthHandler) SignUpHandler(c *gin.Context) {
	var newUser, user models.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := handler.collection.FindOne(handler.ctx, bson.M{
		"username": user.Username,
	}).Decode(&user)

	if err != mongo.ErrNoDocuments { // username already exists
		c.JSON(http.StatusBadRequest, gin.H{"error": "username already exists"})
		return
	} else if err != mongo.ErrNoDocuments && err != nil { // unknonw database error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// do not insert plaintext
	h := sha256.New()
	newUser.Password = string(h.Sum([]byte(newUser.Password)))

	// insert the new user into database
	_, err = handler.collection.InsertOne(handler.ctx, newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sessionToken := xid.New().String()
	session := sessions.Default(c)
	session.Set("username", user.Username)
	session.Set("token", sessionToken)
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "sign up successful"})
}

func (handler *AuthHandler) AuthMiddileware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		sessionToken := session.Get("token")
		if sessionToken == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "not signed in"})
			c.Abort()
		}
		c.Next()
	}
}
