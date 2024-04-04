package main

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Define MongoDB connection string and database/collection names
const (
	connectionString = "mongodb://localhost:27017"
	dbName           = "testdb"
	collectionName   = "items"
)

// Item struct for MongoDB documents
type Item struct {
	ID    string `json:"id" bson:"_id,omitempty"`
	Name  string `json:"name" bson:"name"`
	Price int    `json:"price" bson:"price"`
}

func main() {
	// Create MongoDB client
	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Create Echo instance
	e := echo.New()

	// Routes
	e.GET("/items", getAllItems(client))
	e.GET("/items/:id", getItem(client))
	e.POST("/items", createItem(client))
	e.PUT("/items/:id", updateItem(client))
	e.DELETE("/items/:id", deleteItem(client))

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func getItem(client *mongo.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		collection := client.Database(dbName).Collection(collectionName)
		var item Item
		err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&item)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, item)
	}
}

func createItem(client *mongo.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		item := new(Item)
		if err := c.Bind(item); err != nil {
			return err
		}
		collection := client.Database(dbName).Collection(collectionName)
		_, err := collection.InsertOne(context.Background(), item)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusCreated, item)
	}
}

func updateItem(client *mongo.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid ID")
		}
		item := new(Item)
		if err := c.Bind(item); err != nil {
			return err
		}
		collection := client.Database(dbName).Collection(collectionName)
		_, err = collection.ReplaceOne(context.Background(), bson.M{"_id": id}, item)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, item)
	}
}

func deleteItem(client *mongo.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		collection := client.Database(dbName).Collection(collectionName)
		_, err = collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.NoContent(http.StatusNoContent)
	}
}

func getAllItems(client *mongo.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		collection := client.Database(dbName).Collection(collectionName)
		cursor, err := collection.Find(context.Background(), bson.M{})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		defer cursor.Close(context.Background())

		var items []Item
		for cursor.Next(context.Background()) {
			var item Item
			if err := cursor.Decode(&item); err != nil {
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
			items = append(items, item)
		}
		if err := cursor.Err(); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, items)
	}
}
