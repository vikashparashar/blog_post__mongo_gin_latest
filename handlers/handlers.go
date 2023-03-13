package handlers

import (
	"context"
	"fmt"
	"log"
	"mongo_gin/databases"
	"mongo_gin/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateBlogPost(c *gin.Context) {
	var blogPost models.BlogPost
	if err := c.ShouldBindJSON(&blogPost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := databases.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("blog").Collection("posts")
	blogPost.CreatedAt = time.Now()
	blogPost.UpdatedAt = time.Now()
	result, err := collection.InsertOne(context.Background(), blogPost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert blog post"})
		return
	}

	id := result.InsertedID.(primitive.ObjectID)
	blogPost.ID = id
	c.JSON(http.StatusCreated, blogPost)
}

func GetBlogPost(c *gin.Context) {
	id := c.Param("id")

	client, err := databases.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	defer client.Disconnect(context.Background())

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var blogPost models.BlogPost
	collection := client.Database("blog").Collection("posts")
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&blogPost)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog post not found"})
		return
	}

	c.JSON(http.StatusOK, blogPost)
}

func GetBlogPosts(c *gin.Context) {
	client, err := databases.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	defer client.Disconnect(context.Background())

	var blogPosts []models.BlogPost
	collection := client.Database("blog").Collection("posts")
	cur, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve blog posts"})
		return
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var blogPost models.BlogPost
		err := cur.Decode(&blogPost)
		if err != nil {
			log.Fatal(err)
		}
		blogPosts = append(blogPosts, blogPost)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, blogPosts)
}

func UpdateBlogPost(c *gin.Context) {
	id := c.Param("id")
	var blogPost models.BlogPost
	if err := c.ShouldBindJSON(&blogPost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := databases.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	defer client.Disconnect(context.Background())

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{
		"title":   blogPost.Title,
		"content": blogPost.Content,
	}}

	collection := client.Database("blog").Collection("posts")
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update blog post"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Blog post with ID %s updated", id)})
}

func DeleteBlogPost(c *gin.Context) {
	id := c.Param("id")
	client, err := databases.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	defer client.Disconnect(context.Background())

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	filter := bson.M{"_id": objectID}
	collection := client.Database("blog").Collection("posts")
	result, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete blog post"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Blog post with ID %s deleted", id)})
}
