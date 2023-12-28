package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoInstance

const dbName = "fiber-hrms"
const mongoURI = "mongodb://localhost:27017/" + dbName

type Employee struct {
	Id     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name   string             `json:"name" bson:"name"`
	Salary float64            `json:"salary" bson:"salary"`
	Age    float64            `json:"age" bson:"age"`
}

func Connect() error {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	db := client.Database(dbName)
	if err != nil {
		return err
	}
	mg = MongoInstance{
		client,
		db,
	}
	return nil
}

func main() {
	if err := Connect(); err != nil {
		log.Fatal(err)
	}
	app := fiber.New()
	app.Get("/employee", func(c *fiber.Ctx) error {
		query := bson.D{{}}
		cursor, err := mg.Db.Collection("employees").Find(c.Context(), query)
		if err != nil {
			return c.Status(500).SendString(err.Error())
			defer cursor.Close(c.Context())

		}
		var employees []Employee = make([]Employee, 0)
		if err := cursor.All(c.Context(), &employees); err != nil {
			return c.Status(500).SendString(err.Error())

		}
		return c.JSON(employees)
	})
	app.Post("/employee", func(c *fiber.Ctx) error {
		var employee Employee
		if err := c.BodyParser(&employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}
		insertResult, err := mg.Db.Collection("employees").InsertOne(c.Context(), employee)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.JSON(insertResult)
	})

	app.Put("/employee/:id", func(c *fiber.Ctx) error {
		var employee Employee
		if err := c.BodyParser(&employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}
		id, err := primitive.ObjectIDFromHex(c.Params("id"))
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
		update := bson.D{
			{"$set", bson.D{
				{"name", employee.Name},
				{"salary", employee.Salary},
				{"age", employee.Age},
			}},
		}
		updateResult, err := mg.Db.Collection("employees").UpdateOne(c.Context(),
			bson.M{"_id": id}, update)
		if err != nil {
			return c.Status(500).SendString(err.Error())

		}
		return c.JSON(updateResult)
	})

	app.Delete("/employee/:id", func(c *fiber.Ctx) error {
		id, err := primitive.ObjectIDFromHex(c.Params("id"))
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
		deleteResult, err := mg.Db.Collection("employees").DeleteOne(c.Context(), bson.M{"_id": id})
		if err != nil {
			return c.Status(500).SendString(err.Error())

		}
		return c.JSON(deleteResult)

	})

	log.Fatal(app.Listen(":3000"))

}
