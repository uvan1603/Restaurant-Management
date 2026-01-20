package controllers

import (
	"context"
	"net/http"
	"time"

	"restaurant-management/database"
	"restaurant-management/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var orderCollection *mongo.Collection =
	database.OpenCollection(database.Client, "orders")

// GET ALL ORDERS
func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := orderCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var orders []models.Order
		if err := cursor.All(ctx, &orders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, orders)
	}
}

// GET ORDER BY ID
func GetOrderById() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		orderId := c.Param("order_id")
		var order models.Order

		err := orderCollection.FindOne(
			ctx,
			bson.M{"order_id": orderId},
		).Decode(&order)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}

		c.JSON(http.StatusOK, order)
	}
}

// CREATE ORDER
func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var order models.Order
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Only validate fields provided by client
		if order.Table_id == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "table_id is required"})
			return
		}

		order.ID = primitive.NewObjectID()
		order.Order_id = order.ID.Hex()
		order.Order_Date = time.Now()
		order.Created_at = time.Now()
		order.Updated_at = time.Now()

		result, err := orderCollection.InsertOne(ctx, order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "order not created"})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

// UPDATE ORDER
func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		orderId := c.Param("order_id")
		var input models.Order
		var updateObj primitive.D

		if err := c.BindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.Table_id != nil {
			updateObj = append(updateObj, bson.E{"table_id", input.Table_id})
		}

		if len(updateObj) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		updateObj = append(updateObj, bson.E{"updated_at", time.Now()})

		result, err := orderCollection.UpdateOne(
			ctx,
			bson.M{"order_id": orderId},
			bson.D{{"$set", updateObj}},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "order update failed"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
