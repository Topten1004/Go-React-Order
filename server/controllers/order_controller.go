package controllers

import (
	"context"
	"net/http"
	"server/configs"
	"server/models"
	"server/responses"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var orderCollection *mongo.Collection = configs.GetCollection(configs.DB, "orders")
var validate = validator.New()

func AddOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var order models.Order
		defer cancel()

		// validate the request body
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, responses.OrderResponse{
				Status:  http.StatusBadRequest,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		// use the validate library to validate fields
		if validationErr := validate.Struct(&order); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.OrderResponse{
				Status:  http.StatusBadRequest,
				Message: "error",
				Data: map[string]interface{}{
					"data": validationErr.Error(),
				},
			})
			return
		}

		newOrder := models.Order{
			ID:     primitive.NewObjectID(),
			Dish:   order.Dish,
			Price:  order.Price,
			Server: order.Server,
			Table:  order.Table,
		}

		result, err := orderCollection.InsertOne(ctx, newOrder)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.OrderResponse{
				Status:  http.StatusBadRequest,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusOK, responses.OrderResponse{
			Status:  http.StatusOK,
			Message: "success",
			Data: map[string]interface{}{
				"data": result,
			},
		})
	}
}

func GetAllOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var orders []models.Order
		defer cancel()

		results, err := orderCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.OrderResponse{
				Status:  http.StatusInternalServerError,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		// reading from the db in an optimal way
		defer results.Close(ctx)
		for results.Next(ctx) {
			var singleOrder models.Order
			if err := results.Decode(&singleOrder); err != nil {
				c.JSON(http.StatusInternalServerError, responses.OrderResponse{
					Status:  http.StatusInternalServerError,
					Message: "error",
					Data: map[string]interface{}{
						"data": err.Error(),
					},
				})
			}
			orders = append(orders, singleOrder)
		}

		c.JSON(http.StatusOK, orders)
	}
}

func GetOrderById() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		orderId := c.Param("orderId")
		var order models.Order
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(orderId)

		err := orderCollection.FindOne(ctx, bson.M{"_id": objId}).Decode(&order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.OrderResponse{
				Status:  http.StatusInternalServerError,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusOK, responses.OrderResponse{
			Status:  http.StatusOK,
			Message: "success",
			Data: map[string]interface{}{
				"data": order,
			},
		})
	}
}

func GetOrderByWaiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		waiter := c.Param("waiter")
		var orders []models.Order
		defer cancel()

		result, err := orderCollection.Find(ctx, bson.M{"server": waiter})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.OrderResponse{
				Status:  http.StatusInternalServerError,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		defer result.Close(ctx)
		for result.Next(ctx) {
			var singleOrder models.Order
			if err := result.Decode(&singleOrder); err != nil {
				c.JSON(http.StatusInternalServerError, responses.OrderResponse{
					Status:  http.StatusInternalServerError,
					Message: "error",
					Data: map[string]interface{}{
						"data": err.Error(),
					},
				})
			}
			orders = append(orders, singleOrder)
		}

		c.JSON(http.StatusOK, responses.OrderResponse{
			Status:  http.StatusOK,
			Message: "success",
			Data: map[string]interface{}{
				"data": orders,
			},
		})
	}
}

func UpdateWaiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		orderId := c.Param("orderId")
		defer cancel()

		type Waiter struct {
			Server *string `json:"server"`
		}
		var waiter Waiter

		objId, _ := primitive.ObjectIDFromHex(orderId)

		// validate the request body
		if err := c.BindJSON(&waiter); err != nil {
			c.JSON(http.StatusBadRequest, responses.OrderResponse{
				Status:  http.StatusBadRequest,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		// user the validator library to validate fields
		if validationErr := validate.Struct(&waiter); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.OrderResponse{
				Status:  http.StatusBadRequest,
				Message: "error",
				Data: map[string]interface{}{
					"data": validationErr.Error(),
				},
			})
			return
		}

		result, err := orderCollection.UpdateOne(ctx,
			bson.M{"_id": objId},
			bson.D{
				{"$set", bson.D{{"server", waiter.Server}}},
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.OrderResponse{
				Status:  http.StatusInternalServerError,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		// get updated order details
		var updatedOrder models.Order
		if result.MatchedCount == 1 {
			err := orderCollection.FindOne(ctx, bson.M{"_id": objId}).Decode(&updatedOrder)
			if err != nil {
				c.JSON(http.StatusInternalServerError, responses.OrderResponse{
					Status:  http.StatusInternalServerError,
					Message: "error",
					Data: map[string]interface{}{
						"data": err.Error(),
					},
				})
				return
			}
		}

		c.JSON(http.StatusOK, responses.OrderResponse{
			Status:  http.StatusOK,
			Message: "success",
			Data: map[string]interface{}{
				"data": updatedOrder,
			},
		})
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		orderId := c.Param("orderId")
		var order models.Order
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(orderId)

		// validate the request body
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, responses.OrderResponse{
				Status:  http.StatusBadRequest,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		// user the validator library to validate fields
		if validationErr := validate.Struct(&order); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.OrderResponse{
				Status:  http.StatusBadRequest,
				Message: "error",
				Data: map[string]interface{}{
					"data": validationErr.Error(),
				},
			})
			return
		}

		result, err := orderCollection.ReplaceOne(ctx,
			bson.M{"_id": objId},
			bson.M{
				"dish":   order.Dish,
				"price":  order.Price,
				"server": order.Server,
				"table":  order.Table,
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.OrderResponse{
				Status:  http.StatusInternalServerError,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		// get updated order details
		var updatedOrder models.Order
		if result.MatchedCount == 1 {
			err := orderCollection.FindOne(ctx, bson.M{"_id": objId}).Decode(&updatedOrder)
			if err != nil {
				c.JSON(http.StatusInternalServerError, responses.OrderResponse{
					Status:  http.StatusInternalServerError,
					Message: "error",
					Data: map[string]interface{}{
						"data": err.Error(),
					},
				})
				return
			}
		}

		c.JSON(http.StatusOK, responses.OrderResponse{
			Status:  http.StatusOK,
			Message: "success",
			Data: map[string]interface{}{
				"data": updatedOrder,
			},
		})
	}
}

func DeleteOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		orderId := c.Param("orderId")
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(orderId)

		result, err := orderCollection.DeleteOne(ctx, bson.M{"_id": objId})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.OrderResponse{
				Status:  http.StatusInternalServerError,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		if result.DeletedCount < 1 {
			c.JSON(http.StatusNotFound, responses.OrderResponse{
				Status:  http.StatusNotFound,
				Message: "error",
				Data: map[string]interface{}{
					"data": "Order with specified ID not found.",
				},
			})
			return
		}

		c.JSON(http.StatusOK, responses.OrderResponse{
			Status:  http.StatusOK,
			Message: "success",
			Data: map[string]interface{}{
				"data": "Order successfully deleted.",
			},
		})
	}
}
