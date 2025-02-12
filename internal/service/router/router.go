package router

import (
	auth "avito_staj_2025/internal/auth/controller"
	merch "avito_staj_2025/internal/merch/controller"
	"github.com/gorilla/mux"
)

func SetUpRoutes(authHandler *auth.AuthHandler, merchHandler *merch.MerchHandler) *mux.Router {
	router := mux.NewRouter()
	api := "/api"

	router.HandleFunc(api+"/auth", authHandler.LoginUser).Methods("POST")               // Auth or register user
	router.HandleFunc(api+"/info", merchHandler.GetUserMerchInformation).Methods("GET") // Get user inventory and transactions info
	router.HandleFunc(api+"/buy/{item}", merchHandler.BuyItem).Methods("GET")           // Buy item by user
	router.HandleFunc(api+"/sendCoin", merchHandler.SendCoins).Methods("POST")          // Send coins to other user
	return router
}
