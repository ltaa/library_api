package handlers

import (
	"net/http"
	"github.com/gorilla/mux"
	//"errors"
	//"database/sql/driver"
	//"log"
	"fmt"
)


type Router struct {
	router *mux.Router
}


func (s *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		rw.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	if req.Method == "OPTIONS" {
		return
	}
	s.router.ServeHTTP(rw, req)
}



func NewMux() *Router {

	r := &Router{}

	r.router = mux.NewRouter()
	r.router.HandleFunc("/", printBook)
	r.router.Methods("GET").Path("/api/card/{id}").Handler(http.HandlerFunc(auth(cardQueryOperation)))
	r.router.Methods("POST").Path("/api/card/{id}").Handler(http.HandlerFunc(auth(cardAddHandler)))
	r.router.Methods("POST").Path("/api/card/{id}/{clientId}").Handler(http.HandlerFunc(auth(cardCheckout)))

	r.router.Methods("POST").Path("/api/card").Handler(http.HandlerFunc(auth(cardStateHandler)))


	r.router.Methods("GET").Path("/api/books").Queries("query", "{query}").HandlerFunc(booksQuerySearch)
	r.router.Methods("GET").Path("/api/books").Handler(http.HandlerFunc(auth(getBookInstances)))
	r.router.Methods("UPDATE").Path("/api/books").Handler(http.HandlerFunc(auth(updateBook)))
	r.router.Methods("DELETE").Path("/api/books").Handler(http.HandlerFunc(auth(deledeBook)))


	r.router.Methods("POST").Path("/api/books").Handler(http.HandlerFunc(auth(createBook)))


	//r.router.Methods("GET").Path("/api/clients").Queries("id", "{id}").HandlerFunc(getClientsList)
	r.router.Methods("GET").Path("/api/clients").HandlerFunc(getClientsList)
	r.router.Methods("POST").Path("/api/clients").HandlerFunc(registerClient)
	r.router.Methods("GET").Path("/api/clients/{clientId}").HandlerFunc(auth(readerDataList))
	r.router.Methods("POST").Path("/api/clients/{clientId}").HandlerFunc(auth(clientReturnBooks))

	r.router.Methods("POST").Path("/api/auth").Handler(http.HandlerFunc(getToken))
	return r
}


func printBook(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello, server"))
}

func writeOk(w http.ResponseWriter, b []byte) error {

	if (b == nil) {
		return fmt.Errorf("bytes slice is empty")
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return nil
}