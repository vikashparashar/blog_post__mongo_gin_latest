package controllers

import (
	"context"
	"mongo_gin/models"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("my_secret_key")

func connectDB() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func SignupHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user.Password = string(hashedPassword)
	// Insert the user into the database
	client, err := connectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer client.Disconnect(context.Background())

	// creating a database named "jwt-auth" and a collection with name "user" in mongodb database to store information
	collection := client.Database("jwt-auth").Collection("users")

	// lets insert the username and password in mongodb database
	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Create the JWT token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	// getting userid from mongodb database and store it into claims
	claims["user_id"] = result.InsertedID.(primitive.ObjectID).Hex()

	// setting up the expiration time for jwt token
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// generating token with claims and secret key
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Return the JWT token to the client to next request so that server can know the user everytime client sent request to server
	c.JSON(http.StatusOK, gin.H{
		"token": signedToken,
	})
}

func LoginHandler(c *gin.Context) {

	// storing body payload into user
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Find the user in the database
	client, err := connectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer client.Disconnect(context.Background())
	collection := client.Database("jwt-auth").Collection("users")

	//creating a filter to check the username and password in mongodb database
	filter := bson.M{"username": user.Username}
	var result models.User

	// find the user in mongodb database which we got in request payload
	// and store it into result variable
	err = collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": "Invalid username or password"})
		return
	}
	// Check the password with store in database with password we got in request payload
	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": "Invalid username or password"})
		return
	}
	// Create the JWT token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = result.ID.Hex()
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Return the JWT token to the client
	c.JSON(http.StatusOK, gin.H{
		"token": signedToken,
	})
}
