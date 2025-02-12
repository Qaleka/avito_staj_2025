package main

import (
	authController "avito_staj_2025/internal/auth/controller"
	authRepository "avito_staj_2025/internal/auth/repository"
	authUsecase "avito_staj_2025/internal/auth/usecase"

	merchController "avito_staj_2025/internal/merch/controller"
	merchRepository "avito_staj_2025/internal/merch/repository"
	merchUsecase "avito_staj_2025/internal/merch/usecase"
	"avito_staj_2025/internal/service/logger"
	"avito_staj_2025/internal/service/middleware"
	"avito_staj_2025/internal/service/router"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {
	_ = godotenv.Load()
	//middleware.InitRedis()
	//redisStore := session.NewRedisSessionStore(middleware.RedisClient)
	db := middleware.DbConnect()
	jwtToken, err := middleware.NewJwtToken("secret-key")
	if err != nil {
		log.Fatalf("Failed to create JWT token: %v", err)
	}

	if err := logger.InitLoggers(); err != nil {
		log.Fatalf("Failed to initialize loggers: %v", err)
	}
	defer func() {
		err := logger.SyncLoggers()
		if err != nil {
			log.Fatalf("Failed to sync loggers: %v", err)
		}
	}()

	//sessionService := session.NewSessionService(redisStore)

	authRepository := authRepository.NewAuthRepository(db)
	authUseCase := authUsecase.NewAuthUsecase(authRepository)
	authHandler := authController.NewAuthHandler(authUseCase, jwtToken)

	merchRepository := merchRepository.NewMerchRepository(db)
	merchUseCase := merchUsecase.NewMerchUsecase(merchRepository)
	merchHandler := merchController.NewMerchHandler(merchUseCase, jwtToken)

	mainRouter := router.SetUpRoutes(authHandler, merchHandler)
	mainRouter.Use(middleware.RequestIDMiddleware)
	mainRouter.Use(middleware.RateLimitMiddleware)
	http.Handle("/", middleware.EnableCORS(mainRouter))
	fmt.Printf("Starting HTTP server on adress %s\n", os.Getenv("BACKEND_URL"))
	if err := http.ListenAndServe(os.Getenv("BACKEND_URL"), nil); err != nil {
		fmt.Printf("Error on starting server: %s", err)
	}
}
