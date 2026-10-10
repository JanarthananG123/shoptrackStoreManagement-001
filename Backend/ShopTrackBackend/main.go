package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Main (+)")
	lRouter := http.NewServeMux()
	lRouter.Handle("/ShopTracker/", http.HandlerFunc(func(pWriter http.ResponseWriter, pResponse *http.Request) {
		pWriter.Write([]byte("Hello World"))
	}))
	http.ListenAndServe(":7070", lRouter)
	log.Println("Main (-)")
}
