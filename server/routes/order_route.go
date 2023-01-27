package routes

import (
	"server/controllers"

	"github.com/gin-gonic/gin"
)

func OrderRoute(router *gin.Engine) {
	// All order routes come here.
	router.POST("/order/create", controllers.AddOrder())
	router.GET("/orders", controllers.GetAllOrders())
	router.GET("/order/:orderId", controllers.GetOrderById())
	router.GET("/waiter/:waiter", controllers.GetOrderByWaiter())

	router.PUT("/waiter/update/:orderId", controllers.UpdateWaiter())
	router.PUT("/order/update/:orderId", controllers.UpdateOrder())

	router.DELETE("/order/delete/:orderId", controllers.DeleteOrder())
}
