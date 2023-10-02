package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/HironixRotifer/golang-eccomerce/database"
	"github.com/HironixRotifer/golang-eccomerce/models"
	"github.com/HironixRotifer/golang-eccomerce/tokens"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	UserCollection    *mongo.Collection = database.UserData(database.Client, "Users")
	ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
	Validate                            = validator.New()
)

// HashPassword take password string and return hash
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerefiPassword(userPassword string, givenPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPassword))
	valid := true
	msg := ""

	if err != nil {
		msg = "Login or password is incorrect"
	}
	return valid, msg
}

func SingUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		validationErr := Validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exist"})
		}

		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})

		defer cancel()

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this phone number is already use"})
			return
		}

		password := HashPassword(*user.Password)
		token, refreshtoken, _ := tokens.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, *&user.User_ID)

		user.ID = primitive.NewObjectID()
		user.User_ID = user.ID.Hex()
		user.Password = &password
		user.Token = &token
		user.Refresh_Token = &refreshtoken
		user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.User_Cart = make([]models.ProductUser, 0)
		user.Address_Details = make([]models.Address, 0)
		user.Order_Status = make([]models.Order, 0)

		_, inserterr := UserCollection.InsertOne(ctx, user)
		if inserterr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "the user didn`t get created"})
			return
		}

		defer cancel()

		c.JSON(http.StatusCreated, gin.H{"message": "Successfully signed in!"})

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or password incorrect"})
			return
		}

		PasswordIsValid, msg := VerefiPassword(*user.Password, *foundUser.Password)

		defer cancel()

		if !PasswordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			fmt.Println(msg)
			return
		}

		token, refreshToken, _ := tokens.TokenGenerator(*foundUser.Email, *foudUser.First_Name, *foudUser.Last_Name, *foudUser.User_id)

		defer cancel()

		tokens.UpdateAllTokens(token, refreshToken, foudUser.User_ID)
		c.JSON(http.StatusFound, foudUser)

	}
}

func productViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {

		var productList []models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := ProductCollection.Find(ctx, bson.D{{}})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "somthing went wrong, prlese try after some time")
			return
		}

		err = cursor.All(ctx, &productList)

		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		defer cursor.Close(ctx)

		if err := cursor.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, "invalid")
			return
		}
		defer cancel()

		c.IndentedJSON(http.StatusOK, productList)
	}
}

func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {

		var searchProducts []models.Product
		queryParam := c.Query("name")

		// you want to check if it`s empty
		if queryParam == "" {
			log.Println("query is empty")
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid search index"})
			c.Abort()
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		searchQueryDB, err := ProductCollection.Find(ctx, bson.M{"product_name": bson.M{"$regex": queryParam}})
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, "something went wrong while fetching the data")
			return
		}

		err = searchQueryDB.All(ctx, &searchProducts)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, "invalid")
			return
		}

		defer searchQueryDB.Close(ctx)

		if err := searchQueryDB.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, "invalid request")
			return
		}

		defer cancel()

		c.IndentedJSON(http.StatusOK, searchProducts)
	}
}
