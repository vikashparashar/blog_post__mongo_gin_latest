package routes

import (
	"log"

	"mongo_gin/controllers"
	"mongo_gin/handlers"

	"github.com/gin-gonic/gin"
)

func StartApp() {
	r := gin.Default()
	// Create Signup
	r.POST("/signup", controllers.SignupHandler)

	// Create Login
	r.POST("/login", controllers.LoginHandler)
	// Create blog post
	r.POST("/posts", handlers.CreateBlogPost)

	// Get blog post by ID
	r.GET("/posts/:id", handlers.GetBlogPost)

	// Get all blog posts
	r.GET("/posts", handlers.GetBlogPosts)

	// Update blog post by ID
	r.PUT("/posts/:id", handlers.UpdateBlogPost)

	// Delete blog post by ID
	r.DELETE("/posts/:id", handlers.DeleteBlogPost)
	// check to see if we find any error during the server startup
	// if we find any error then stop the execution of program immediatly
	log.Fatal(r.Run(":8080"))
}
