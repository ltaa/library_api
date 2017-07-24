package handlers

import (
	"log"
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"io/ioutil"
	"io"
	"fmt"
	"database/sql"
	"errors"
)


type book struct {
	Instance_id int `json:"instance_id"`
	Name string `json:"name"`
	Year int `json:"year"`
	Author bookAuthors `json:"author"`
	Publisher string `json:"publisher"`
}


func validateCreateJson(b *book) (error) {
	if b.Name == "" || b.Publisher == "" || b.Year == 0 || len(b.Author) == 0 {
		return fmt.Errorf("json is empty")
	}

	return  nil
}

func createBookResponse(msg string) []byte {
	b := createBookJson{Message: msg}
	data, err := json.Marshal(&b)
	if err != nil {
		log.Print(err)
		return nil
	}
	return data
}


func updateBookResponse(msg string) []byte {
	b := updateBookJson{Message: msg}
	data, err := json.Marshal(&b)
	if err != nil {
		log.Print(err)
		return nil
	}
	return data
}


func deledeBook(w http.ResponseWriter, r *http.Request) {
	data, err :=  ioutil.ReadAll(io.LimitReader(r.Body, maxReadLen))

	if err != nil {
		log.Print(err)
		return
	}
	b := make([]book, 0, 1)

	if err = json.Unmarshal(data, &b); err != nil {
		log.Print(err)
		return
	}

	if err = deleteBookInstance(b) ; err != nil {
		log.Print(err)

	}

	w.WriteHeader(http.StatusOK)

}

func createBook(w http.ResponseWriter, r *http.Request) {

	inputData, err := ioutil.ReadAll(io.LimitReader(r.Body, maxReadLen))

	if err != nil {
		log.Print(err)
		return
	}


	b := book{}
	json.Unmarshal(inputData, &b)
	// if error handler

	if err := validateCreateJson(&b); err != nil {
		return
	}
	createBookDb(&b)

	if err = writeOk(w, createBookResponse("book added")); err != nil {
		log.Print(err)
	}
}


func updateBook(w http.ResponseWriter, r *http.Request) {

	b, err := ioutil.ReadAll(io.LimitReader(r.Body, maxReadLen))

	if err != nil {
		log.Print(err)
	}

	if len(b) == 0 {
		log.Print("data is empty")
		return
	}

	u := book{}
	if err = json.Unmarshal(b, &u); err != nil {
		log.Print(err)
		return
	}

	if err := updateBookInstance(&u) ; err != nil {
		log.Print(err)
		return
	}

	if err := writeOk(w, updateBookResponse("updated success")); err != nil {
		log.Print(err)
	}

}


func booksQuerySearch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]

	state := library
	queryString := `select book_instances.instance_id,
		books.book_name,
		books.year,
		publishers.publisher_name,
		authors.authors,
		book_instances.state
		from
		(
		select array_agg(concat_ws('_',a.first_name,a.last_name)) as authors,
			b.book_id as book_id
			from
			authors as a join authors_books b on a.author_id=b.author_id group by b.book_id
			) authors
			join books on books.book_id = authors.book_id
			join publishers on books.publisher_id = publishers.publisher_id
			join book_instances on book_instances.book_id = books.book_id where LOWER(books.book_name) LIKE LOWER( '%' || $1 || '%')`

	var err error
	var rows *sql.Rows
	if state == library {
		queryString += "AND book_instances.state = $2"
		rows, err = db.Query(queryString, query, state)
	} else {

		rows, err = db.Query(queryString, query)
	}

	if err != nil {
		log.Print(err)
		return
	}

	bookSlice, err := getBookMap(rows, state)
	if err != nil {
		log.Print(err)
		return
	}

	if err := writeData(w, http.StatusOK, marshalJson(bookSlice)); err != nil {
		log.Print(err)
	}

	//b, err := json.Marshal(bookSlice)
	//
	//if err != nil {
	//	log.Print(err)
	//	return
	//}
	//
	//w.WriteHeader(http.StatusOK)
	//w.Write(b)
}


func getBookInstances (w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query(`select book_instances.instance_id,
        books.book_name,
    		books.year,
        publishers.publisher_name,
        authors.authors,
        book_instances.state

        from
        (
          select array_agg(concat_ws('_',a.first_name,a.last_name)) as authors,
          b.book_id as book_id
            from
              authors as a join authors_books b on a.author_id=b.author_id group by b.book_id
        ) authors
          join books on books.book_id = authors.book_id
          join publishers on books.publisher_id = publishers.publisher_id
          join book_instances on book_instances.book_id = books.book_id where book_instances.state = $1 ORDER BY instance_id`, library)


	if err != nil {
		log.Print(err)
		return
	}

	bookSlice, err := getBookMap(rows, library)
	if err != nil {
		log.Print(err)
		return
	}

	if err := writeData(w, http.StatusOK, marshalJson(bookSlice)) ; err != nil {
		log.Print(err)
	}

}



func bookChangeState(b []int, state StateChange ) ([]book, error){

	var queryInString string
	queryInString = getPrepareString(len(b), 1)
	inBookIface := getPrepareInterface(b)

	stmt, err := state.tx.Prepare(`select book_instances.instance_id,
        books.book_name,
    		books.year,
        publishers.publisher_name,
        authors.authors,
        book_instances.state
        from
        (
          select array_agg(concat_ws('_',a.first_name,a.last_name)) as authors,
          b.book_id as book_id
            from
              authors as a join authors_books b on a.author_id=b.author_id group by b.book_id
        ) authors
          join books on books.book_id = authors.book_id
          join publishers on books.publisher_id = publishers.publisher_id
          join book_instances on book_instances.book_id = books.book_id where instance_id IN ( ` + queryInString + ")")


	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(inBookIface...)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	bookSlice, err := getBookChangeState(rows, state.curent)
	if err != nil {
		return nil, err
	}

	if bookSlice == nil || len(bookSlice) != len(b) {
		return nil, errors.New("invalid result entity count")
	}

	urows, err := state.tx.Prepare("update book_instances set state = $1 where instance_id = $2")

	if err != nil {
		return nil, err
	}
	defer urows.Close()

	for _, curentBook := range bookSlice {
		if _, err := urows.Exec(state.next, curentBook.Instance_id); err != nil {
			return nil, err
		}
	}

	return bookSlice, nil
}