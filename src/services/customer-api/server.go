package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const uri = "mongodb://localhost:27017"

type application struct {
	coll *mongo.Collection
}

func main() {
	e := echo.New()

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	// Send a ping to confirm a successful connection
	var result bson.M
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	coll := client.Database("customers_db").Collection("customers")

	app := application{
		coll: coll,
	}

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/customers", app.saveCustomer)
	e.GET("/customers", app.getCustomers)
	e.GET("/customers/:id", app.getCustomerById)
	e.PUT("/customers/:id", app.updateCustomer)
	e.DELETE("/customers/:id", app.deleteCustomer)

	e.Logger.Fatal(e.Start(":1323"))
}

type Customer struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FirstName string             `json:"firstName" bson:"firstName,omitempty"`
	LastName  string             `json:"lastName" bson:"lastname,omitempty"`
	Address   string             `json:"address"`
	Phone     string             `json:"phone"`
	Email     string             `json:"email"`
}

func (app *application) saveCustomer(c echo.Context) error {
	customer := new(Customer)
	if err := c.Bind(customer); err != nil {
		return err
	}

	result, err := app.coll.InsertOne(c.Request().Context(), customer)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, result.InsertedID)
}

func (app *application) getCustomers(c echo.Context) error {
	customerCursor, err := app.coll.Find(c.Request().Context(), bson.M{})
	if err != nil {
		return err
	}
	customers := []Customer{}
	err = customerCursor.All(c.Request().Context(), &customers)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, customers)
}

func (app *application) getCustomerById(c echo.Context) error {
	id := c.Param("id")
	p, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	var customer = Customer{}
	err = app.coll.FindOne(c.Request().Context(), bson.M{"_id": p}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("ErrNoDocuments")
		}
		return err
	}
	return c.JSON(http.StatusOK, customer)
}

func (app *application) updateCustomer(c echo.Context) error {
	// User ID from path `users/:id`
	id := c.Param("id")
	p, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.D{{Key: "_id", Value: p}}
	customer := new(Customer)
	if err := c.Bind(customer); err != nil {
		return err
	}

	_, err = app.coll.ReplaceOne(c.Request().Context(), filter, customer)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, "")
}

func (app *application) deleteCustomer(c echo.Context) error {
	id := c.Param("id")
	p, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	result, err := app.coll.DeleteOne(c.Request().Context(), bson.M{"_id": p})
	if err != nil {
		return err
	}
	resStr := "no"
	if result.DeletedCount > 0 {
		resStr = "ok"
	}
	return c.String(http.StatusNoContent, resStr)
}
