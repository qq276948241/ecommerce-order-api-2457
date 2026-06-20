package main

import (
	"log"

	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/handler"
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	err := database.Init(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	db := database.GetDB()

	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	refundRepo := repository.NewRefundRepository(db)

	authService := service.NewAuthService(userRepo, &cfg.JWT)
	productService := service.NewProductService(productRepo)
	cartService := service.NewCartService(cartRepo, productRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, db)
	paymentService := service.NewPaymentService(paymentRepo, orderRepo, db)
	refundService := service.NewRefundService(refundRepo, orderRepo)

	authHandler := handler.NewAuthHandler(authService)
	productHandler := handler.NewProductHandler(productService)
	cartHandler := handler.NewCartHandler(cartService)
	orderHandler := handler.NewOrderHandler(orderService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	refundHandler := handler.NewRefundHandler(refundService)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/profile", middleware.JWTAuth(&cfg.JWT), authHandler.GetProfile)
		}

		products := api.Group("/products")
		{
			products.GET("", productHandler.List)
			products.GET("/:id", productHandler.GetByID)
			products.POST("", middleware.JWTAuth(&cfg.JWT), productHandler.Create)
			products.PUT("/:id", middleware.JWTAuth(&cfg.JWT), productHandler.Update)
			products.DELETE("/:id", middleware.JWTAuth(&cfg.JWT), productHandler.Delete)
		}

		cart := api.Group("/cart")
		cart.Use(middleware.JWTAuth(&cfg.JWT))
		{
			cart.GET("", cartHandler.GetCart)
			cart.POST("", cartHandler.AddItem)
			cart.PUT("/:id", cartHandler.UpdateItem)
			cart.DELETE("/:id", cartHandler.DeleteItem)
		}

		orders := api.Group("/orders")
		orders.Use(middleware.JWTAuth(&cfg.JWT))
		{
			orders.GET("", orderHandler.GetOrderList)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.POST("", orderHandler.CreateOrder)
			orders.POST("/:id/cancel", orderHandler.CancelOrder)
		}

		payments := api.Group("/payments")
		payments.Use(middleware.JWTAuth(&cfg.JWT))
		{
			payments.POST("/pay", paymentHandler.Pay)
			payments.GET("/:id", paymentHandler.GetPayment)
			payments.GET("/order/:order_id", paymentHandler.GetByOrderID)
		}

		refunds := api.Group("/refunds")
		refunds.Use(middleware.JWTAuth(&cfg.JWT))
		{
			refunds.GET("", refundHandler.GetRefundList)
			refunds.GET("/:id", refundHandler.GetRefund)
			refunds.POST("", refundHandler.CreateRefund)
		}
	}

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
