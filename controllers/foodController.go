package controllers

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"restaurant-management/database"
	"restaurant-management/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

/* ---------- helpers ---------- */

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

/* ---------- controllers ---------- */

// GET ALL FOODS
func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, _ := strconv.Atoi(c.DefaultQuery("recordPerPage", "10"))
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

		if recordPerPage < 1 {
			recordPerPage = 10
		}
		if page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage

		pipeline := mongo.Pipeline{
			{{"$match", bson.M{}}},
			{{"$group", bson.M{
				"_id":         nil,
				"total_count": bson.M{"$sum": 1},
				"data":        bson.M{"$push": "$$ROOT"},
			}}},
			{{"$project", bson.M{
				"_id":         0,
				"total_count": 1,
				"food_items":  bson.M{"$slice": []interface{}{"$data", startIndex, recordPerPage}},
			}}},
		}

		cursor, err := foodCollection.Aggregate(ctx, pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var result []bson.M
		if err := cursor.All(ctx, &result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(result) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"total_count": 0,
				"food_items":  []interface{}{},
			})
			return
		}

		c.JSON(http.StatusOK, result[0])
	}
}

// GET FOOD BY ID
func GetFoodById() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		foodId := c.Param("food_id")
		var food models.Food

		if err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "food not found"})
			return
		}

		c.JSON(http.StatusOK, food)
	}
}

// CREATE FOOD
func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var food models.Food
		var menu models.Menu

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := validate.Struct(food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "menu not found"})
			return
		}

		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex()
		food.Created_at = time.Now()
		food.Updated_at = time.Now()

		price := toFixed(*food.Price, 2)
		food.Price = &price

		result, err := foodCollection.InsertOne(ctx, food)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "food not created"})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

// UPDATE FOOD
func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		foodId := c.Param("food_id")
		var food models.Food
		var menu models.Menu

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updateObj primitive.D

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{"name", food.Name})
		}

		if food.Price != nil {
			price := toFixed(*food.Price, 2)
			updateObj = append(updateObj, bson.E{"price", price})
		}

		if food.Food_image != nil {
			updateObj = append(updateObj, bson.E{"food_image", food.Food_image})
		}

		if food.Menu_id != nil {
			if err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "menu not found"})
				return
			}
			updateObj = append(updateObj, bson.E{"menu_id", food.Menu_id})
		}

		updateObj = append(updateObj, bson.E{"updated_at", time.Now()})

		result, err := foodCollection.UpdateOne(
			ctx,
			bson.M{"food_id": foodId},
			bson.D{{"$set", updateObj}},
			options.Update().SetUpsert(false),
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "food update failed"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
