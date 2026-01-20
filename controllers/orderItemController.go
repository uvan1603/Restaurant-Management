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

type OrderItemPack struct {
	Table_id    *string            `json:"table_id" validate:"required"`
	Order_items []models.OrderItem `json:"order_items" validate:"required,dive"`
}

var orderItemCollection *mongo.Collection =
	database.OpenCollection(database.Client, "order_items")

// GET ALL ORDER ITEMS
func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := orderItemCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var items []models.OrderItem
		if err := cursor.All(ctx, &items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, items)
	}
}

// GET ORDER ITEM BY ID
func GetOrderItemById() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		orderItemId := c.Param("orderItem_id")

		var item models.OrderItem
		err := orderItemCollection.FindOne(
			ctx,
			bson.M{"order_item_id": orderItemId},
		).Decode(&item)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "order item not found"})
			return
		}

		c.JSON(http.StatusOK, item)
	}
}

// GET ORDER ITEMS BY ORDER ID
func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderId := c.Param("order_id")

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := orderItemCollection.Find(
			ctx,
			bson.M{"order_id": orderId},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var items []models.OrderItem
		if err := cursor.All(ctx, &items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, items)
	}
}

// UPDATE ORDER ITEM
func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		orderItemId := c.Param("orderItem_id")

		var input models.OrderItem
		if err := c.BindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		update := bson.D{}

		if input.Quantity != nil {
			update = append(update, bson.E{"quantity", input.Quantity})
		}
		if input.Unit_price != nil {
			update = append(update, bson.E{"unit_price", input.Unit_price})
		}
		if input.Food_id != nil {
			update = append(update, bson.E{"food_id", input.Food_id})
		}

		if len(update) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		update = append(update, bson.E{"updated_at", time.Now()})

		result, err := orderItemCollection.UpdateOne(
			ctx,
			bson.M{"order_item_id": orderItemId},
			bson.D{{"$set", update}},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// CREATE ORDER + ORDER ITEMS
func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var pack OrderItemPack
		if err := c.BindJSON(&pack); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(pack.Order_items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "order_items cannot be empty"})
			return
		}

		order := models.Order{
			ID:         primitive.NewObjectID(),
			Order_id:   primitive.NewObjectID().Hex(),
			Table_id:   pack.Table_id,
			Created_at: time.Now(),
			Updated_at: time.Now(),
		}

		_, err := orderCollection.InsertOne(ctx, order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "order creation failed"})
			return
		}

		var docs []interface{}
		for _, item := range pack.Order_items {
			item.ID = primitive.NewObjectID()
			item.Order_item_id = item.ID.Hex()
			item.Order_id = order.Order_id
			item.Created_at = time.Now()
			item.Updated_at = time.Now()
			docs = append(docs, item)
		}

		result, err := orderItemCollection.InsertMany(ctx, docs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert order items"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"order_id":     order.Order_id,
			"order_items":  result,
		})
	}
}


func ItemsByOrder(orderID string) ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	matchStage := bson.D{{"$match", bson.D{{"order_id", orderID}}}}

	lookupFoodStage := bson.D{{"$lookup", bson.D{
		{"from", "food"},
		{"localField", "food_id"},
		{"foreignField", "food_id"},
		{"as", "food"},
	}}}

	unwindFoodStage := bson.D{{"$unwind", bson.D{
		{"path", "$food"},
		{"preserveNullAndEmptyArrays", true},
	}}}

	lookupOrderStage := bson.D{{"$lookup", bson.D{
		{"from", "order"},
		{"localField", "order_id"},
		{"foreignField", "order_id"},
		{"as", "order"},
	}}}

	unwindOrderStage := bson.D{{"$unwind", bson.D{
		{"path", "$order"},
		{"preserveNullAndEmptyArrays", true},
	}}}

	lookupTableStage := bson.D{{"$lookup", bson.D{
		{"from", "table"},
		{"localField", "order.table_id"},
		{"foreignField", "table_id"},
		{"as", "table"},
	}}}

	unwindTableStage := bson.D{{"$unwind", bson.D{
		{"path", "$table"},
		{"preserveNullAndEmptyArrays", true},
	}}}

	// calculate amount = price * quantity
	projectStage := bson.D{{"$project", bson.D{
		{"food_name", "$food.name"},
		{"food_image", "$food.food_image"},
		{"table_number", "$table.table_number"},
		{"table_id", "$table.table_id"},
		{"order_id", "$order.order_id"},
		{"price", "$food.price"},
		{"quantity", 1},
		{"amount", bson.D{{"$multiply", bson.A{"$food.price", "$quantity"}}}},
	}}}

	groupStage := bson.D{{"$group", bson.D{
		{"_id", bson.D{
			{"order_id", "$order_id"},
			{"table_id", "$table_id"},
			{"table_number", "$table_number"},
		}},
		{"payment_due", bson.D{{"$sum", "$amount"}}},
		{"total_count", bson.D{{"$sum", "$quantity"}}},
		{"order_items", bson.D{{"$push", "$$ROOT"}}},
	}}}

	finalProject := bson.D{{"$project", bson.D{
		{"_id", 0},
		{"payment_due", 1},
		{"total_count", 1},
		{"table_number", "$_id.table_number"},
		{"order_items", 1},
	}}}

	cursor, err := orderItemCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		lookupFoodStage,
		unwindFoodStage,
		lookupOrderStage,
		unwindOrderStage,
		lookupTableStage,
		unwindTableStage,
		projectStage,
		groupStage,
		finalProject,
	})
	if err != nil {
		return nil, err
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
