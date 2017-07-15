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


	r.router.Methods("GET").Path("/api/books").Queries("query", "{query}").HandlerFunc(auth(booksQuerySearch))
	r.router.Methods("GET").Path("/api/books").Handler(http.HandlerFunc(auth(getBookInstances)))
	r.router.Methods("UPDATE").Path("/api/books").Handler(http.HandlerFunc(auth(updateBook)))
	r.router.Methods("DELETE").Path("/api/books").Handler(http.HandlerFunc(auth(deledeBook)))


	r.router.Methods("POST").Path("/api/books").Handler(http.HandlerFunc(auth(createBook)))



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


//
//func querySearch(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	query := vars["query"]
//
//	rows, err := db.Query("SELECT * FROM authors where first_name LIKE '%' || $1 || '%'", query)
//
//
//	if err != nil {
//		log.Print(err)
//		return
//	}
//
//	a := make([]author, 0, 1)
//	for rows.Next() {
//		cur_author := author{}
//		err := rows.Scan(&cur_author.Id, &cur_author.First_name, &cur_author.Last_name)
//
//		if err != nil {
//			log.Print(err)
//			return
//		}
//
//		a = append(a, cur_author)
//
//	}
//
//
//	b, err := json.Marshal(a)
//
//	if (err != nil) {
//		log.Print(err)
//		return
//	}
//
//	w.WriteHeader(http.StatusOK)
//	w.Write(b)
//
//
//	log.Print(query)
//}
//
//func postAuthor(w http.ResponseWriter, r *http.Request) {
//
//
//	defer r.Body.Close()
//
//	w.WriteHeader(http.StatusBadRequest)
//	b, err := ioutil.ReadAll(io.LimitReader(r.Body, maxReadLen))
//
//	if err != nil {
//		log.Print(err)
//		return
//	}
//
//	arrayJson := make([]author, 0, 1)
//
//	err = json.Unmarshal(b, &arrayJson)
//
//	if (err != nil) {
//
//		log.Print(err)
//		return
//	}
//
//
//	for _, a := range arrayJson {
//		result, err := db.Exec("INSERT into author (first_name, second_name) VALUES($1, $2)", a.First_name, a.Last_name )
//
//		if err != nil {
//			log.Print(err)
//			return
//		}
//
//		log.Print(result.RowsAffected())
//	}
//
//
//	w.WriteHeader(http.StatusCreated)
//	w.Write([]byte("data created"))
//
//}
//
//
//func printAuthors(w http.ResponseWriter, r *http.Request) {
//
//	row, err := db.Query("select * from authors")
//	if err != nil {
//		fmt.Print(err)
//		w.WriteHeader(http.StatusForbidden)
//		w.Write([]byte("db query error"))
//		return
//	}
//
//	authors := make([]author, 0, 1)
//	for row.Next() {
//		a := author{}
//		row.Scan(&a.Id, &a.First_name, &a.Last_name)
//		authors = append(authors, a)
//	}
//
//	encoder := json.NewEncoder(w)
//
//	//w.Header().Set("Access-Control-Allow-Origin", "*")
//	w.WriteHeader(http.StatusOK)
//	encoder.Encode(authors)
//	//w.Write(data)
//}
