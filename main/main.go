package main

//import "fmt"
import (
	//"library_api/handlers"
	"net/http"
	"fmt"
	"library_api/handlers"
)



func main() {


	router := handlers.NewMux()


	//http.HandleFunc("/books", func (w http.ResponseWriter, r *http.Request) {
	//	//w.Header().Set("Access-Control-Allow-Credentials", "true")
	//	//w.Header().Set("Access-Control-Allow-Origin", "*")
	//
	//
	//	//if origin := r.Header.Get("Origin"); origin != "" {
	//	//	w.Header().Set("Access-Control-Allow-Origin", origin)
	//	//	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	//	//	w.Header().Set("Access-Control-Allow-Headers",
	//	//		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	//	//}
	//
	//
	//	//w.WriteHeader(http.StatusOK)
	//	//
	//	//if r.Method == "OPTIONS" {
	//	//	return
	//	//}




	//	w.Write([]byte("hello, world"))
	//})

	err := http.ListenAndServe(":2020",  router)

	if err != nil {
		fmt.Print(err)
	}

	//fmt.Print("hello, golang!")
}
