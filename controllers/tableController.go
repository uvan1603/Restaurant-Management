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
	"go.mongodb.org/mongo-driver/mongo/options"
)

var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "tables")

// GET ALL TABLES
func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := tableCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var tables []models.Table
		if err = cursor.All(ctx, &tables); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, tables)
	}
}

// GET TABLE BY ID
func GetTableById() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		tableId := c.Param("table_id")
		var table models.Table

		err := tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&table)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "table not found"})
			return
		}

		c.JSON(http.StatusOK, table)
	}
}

// CREATE TABLE
func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var table models.Table
		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := validate.Struct(table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		table.ID = primitive.NewObjectID()
		table.Table_id = table.ID.Hex()
		table.Created_at = time.Now()
		table.Updated_at = time.Now()

		result, err := tableCollection.InsertOne(ctx, table)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "table not created"})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

// UPDATE TABLE
func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		tableId := c.Param("table_id")
		var table models.Table
		var updateObj primitive.D

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if table.Number_of_guests != nil {
			updateObj = append(updateObj, bson.E{"number_of_guests", table.Number_of_guests})
		}

		if table.Table_number != nil {
			updateObj = append(updateObj, bson.E{"table_number", table.Table_number})
		}

		table.Updated_at = time.Now()
		updateObj = append(updateObj, bson.E{"updated_at", table.Updated_at})

		if len(updateObj) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		result, err := tableCollection.UpdateOne(
			ctx,
			bson.M{"table_id": tableId},
			bson.D{{"$set", updateObj}},
			options.Update().SetUpsert(false),
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "table update failed"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
