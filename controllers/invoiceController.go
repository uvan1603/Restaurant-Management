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

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoices")

type InvoiceViewFormat struct {
	Invoice_id       string      `json:"invoice_id"`
	Payment_method   string      `json:"payment_method"`
	Order_id         string      `json:"order_id"`
	Payment_status   *string     `json:"payment_status"`
	Payment_due      interface{} `json:"payment_due"`
	Table_number     interface{} `json:"table_number"`
	Payment_due_date time.Time   `json:"payment_due_date"`
	Order_details    interface{} `json:"order_details"`
}

/* ================= GET ALL INVOICES ================= */
func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := invoiceCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var invoices []models.Invoice
		if err = cursor.All(ctx, &invoices); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, invoices)
	}
}

/* ================= GET INVOICE BY ID ================= */
func GetInvoiceById() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		invoiceId := c.Param("invoice_id")
		var invoice models.Invoice

		if err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
			return
		}

		allOrderItems, err := ItemsByOrder(invoice.Order_id)
		if err != nil || len(allOrderItems) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "order details not found"})
			return
		}

		view := InvoiceViewFormat{
			Invoice_id:       invoice.Invoice_id,
			Order_id:         invoice.Order_id,
			Payment_due_date: invoice.Payment_due_date,
			Payment_status:   invoice.Payment_status,
			Payment_method:   "N/A",
			Payment_due:      allOrderItems[0]["payment_due"],
			Table_number:     allOrderItems[0]["table_number"],
			Order_details:    allOrderItems[0]["order_items"],
		}

		if invoice.Payment_method != nil {
			view.Payment_method = *invoice.Payment_method
		}

		c.JSON(http.StatusOK, view)
	}
}

/* ================= CREATE INVOICE ================= */
func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var invoice models.Invoice
		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// check order exists
		var order models.Order
		if err := orderCollection.FindOne(ctx, bson.M{"order_id": invoice.Order_id}).Decode(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "order not found"})
			return
		}

		if invoice.Payment_status == nil {
			status := "PENDING"
			invoice.Payment_status = &status
		}

		invoice.ID = primitive.NewObjectID()
		invoice.Invoice_id = invoice.ID.Hex()
		invoice.Created_at = time.Now()
		invoice.Updated_at = time.Now()
		invoice.Payment_due_date = time.Now().AddDate(0, 0, 1)

		if err := validate.Struct(invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := invoiceCollection.InsertOne(ctx, invoice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invoice not created"})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

/* ================= UPDATE INVOICE ================= */
func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		invoiceId := c.Param("invoice_id")
		var invoice models.Invoice
		var updateObj primitive.D

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if invoice.Payment_method != nil {
			updateObj = append(updateObj, bson.E{"payment_method", invoice.Payment_method})
		}

		if invoice.Payment_status != nil {
			updateObj = append(updateObj, bson.E{"payment_status", invoice.Payment_status})
		}

		updateObj = append(updateObj, bson.E{"updated_at", time.Now()})

		if len(updateObj) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		result, err := invoiceCollection.UpdateOne(
			ctx,
			bson.M{"invoice_id": invoiceId},
			bson.D{{"$set", updateObj}},
			options.Update().SetUpsert(false),
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invoice update failed"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
