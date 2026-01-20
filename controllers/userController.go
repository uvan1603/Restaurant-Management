package controllers

import (
	"context"
	"restaurant-management/database"
	helper "restaurant-management/helpers"
	"restaurant-management/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, _ := strconv.Atoi(c.DefaultQuery("recordPerPage", "10"))
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		startIndex := (page - 1) * recordPerPage

		cursor, err := userCollection.Find(
			ctx,
			bson.M{},
			options.Find().
				SetSkip(int64(startIndex)).
				SetLimit(int64(recordPerPage)),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var users []models.User
		if err = cursor.All(ctx, &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

func GetUserById() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userId := c.Param("user_id")
		var user models.User

		if err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := validate.Struct(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		emailCount, _ := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		phoneCount, _ := userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})

		if emailCount > 0 || phoneCount > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "email or phone already exists"})
			return
		}

		hashed := HashPassword(*user.Password)
		user.Password = &hashed

		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		user.Created_at = time.Now()
		user.Updated_at = time.Now()

		token, refresh, _ := helper.GenerateAllTokens(
			*user.Email,
			*user.First_name,
			*user.Last_name,
			user.User_id,
		)

		user.Token = &token
		user.Refresh_Token = &refresh

		result, err := userCollection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not created"})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var input models.User
		var user models.User

		if err := c.BindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := userCollection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		ok, _ := VerifyPassword(*input.Password, *user.Password)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token, refresh, _ := helper.GenerateAllTokens(
			*user.Email,
			*user.First_name,
			*user.Last_name,
			user.User_id,
		)

		helper.UpdateAllTokens(token, refresh, user.User_id)

		user.Token = &token
		user.Refresh_Token = &refresh

		c.JSON(http.StatusOK, user)
	}
}

func HashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}

func VerifyPassword(input string, hashed string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(input))
	if err != nil {
		return false, "invalid password"
	}
	return true, ""
}
