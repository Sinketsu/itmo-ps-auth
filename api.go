package main

import (
	"itmo-ps-auth/handlers"
	"itmo-ps-auth/middleware"
	"net/http"
)

func GetAPI() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/", middleware.AuthRequired(http.HandlerFunc(handlers.Index)))
	mux.HandleFunc("/signup", handlers.SignUp)
	mux.HandleFunc("/signin", handlers.SignIn)
	mux.Handle("/logout", middleware.AuthRequired(http.HandlerFunc(handlers.Logout)))

	return mux
}
