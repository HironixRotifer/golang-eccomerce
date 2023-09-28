package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/HironixRotifer/golang-eccomerce/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddAddress() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func EditHomeAddress() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func EditWorkAddress() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

// DeleteAddress delete address from user id
func DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("userID")

		if user_id == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid search index"})
			c.Abort()
			return
		}

		addresses := make([]models.Address, 0)
		userObjectID, err := primitive.ObjectIDFromHex(user_id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Internal Server error")
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()
		filter := bson.D{primitive.E{Key: "_id", Value: userObjectID}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, "Wrong command")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(http.StatusOK, "Successfully Deleted")
	}
}
