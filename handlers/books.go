package handlers

import (
	"log"
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"strconv"
	"database/sql"
	"io/ioutil"
	"io"
	"fmt"
	//"strings"
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
	data,err := json.Marshal(&b)
	if err != nil {
		log.Print(err)
		return nil
	}
	return data
}


func updateBookResponse(msg string) []byte {
	b := updateBookJson{Message: msg}
	data,err := json.Marshal(&b)
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

	err = json.Unmarshal(data, &b)

	if err != nil {
		log.Print(err)
		return
	}

	err = deleteBookInstance(b)

	if err != nil {
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


	log.Print(b)

	if err := validateCreateJson(&b); err != nil {
		return
	}
	createBookDb(&b)

	err = writeOk(w, createBookResponse("book added"))

	if err != nil {
		log.Print(err)
	}
}

type updateJson struct {
	Book_id int	`json:"book_id"`
	book
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
	err = json.Unmarshal(b, &u)
	if err != nil {
		log.Print(err)
		return
	}


	err = updateBookInstance(&u)

	if err != nil {
		log.Print(err)
		return
	}

	err = writeOk(w, updateBookResponse("updated success"))

	if err != nil {
		log.Print(err)
	}


}






func booksQuerySearch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]

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
			join book_instances on book_instances.book_id = books.book_id where LOWER(books.book_name) LIKE LOWER( '%' || $1 || '%')`, query)


	if err != nil {
		log.Print(err)
		return
	}

bookSlice := make([]book, 0, 1)

	for rows.Next() {
		b := book{}
		var state BookState
		rows.Scan(&b.Instance_id, &b.Name, &b.Year, &b.Publisher, &b.Author, &state)

		if state != "library" {
			continue
		}


		bookSlice = append(bookSlice, b)

	}

	b, err := json.Marshal(bookSlice)

	if err != nil {
		log.Print(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
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
          join book_instances on book_instances.book_id = books.book_id`)




	if err != nil {
		log.Print(err)
		return
	}

	bookSlice := make([]book, 0, 1)

	for rows.Next() {
		b := book{}
		var state BookState
		rows.Scan(&b.Instance_id, &b.Name, &b.Year, &b.Publisher, &b.Author, &state)

		if state != "library" {
			continue
		}

		bookSlice = append(bookSlice, b)

	}

	data, err := json.Marshal(bookSlice)
	log.Print(data)

	if err != nil {
		log.Print(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)

}



func bookRequestDb(b []int, state BookState, tx *sql.Tx ) ([]book){

	var next_state BookState
	if state == library {
		next_state = processing
	} else if state == processing {
		next_state = client
	} else if state == client {
		next_state = library
	}

	var queryInString string
	queryInString = getPrepareString(len(b), 1)
	inBookIface := getPrepareInterface(b)

	stmt, err := tx.Prepare(`select book_instances.instance_id,
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
		log.Print(err)
		return nil
	}

	log.Printf("len of inBookIface ", len(inBookIface))
	rows, err := stmt.Query(inBookIface...)

	if err != nil {
		log.Print(err)
		return nil
	}

	defer stmt.Close()

	bookSlice := make([]book, 0, 1)

	for rows.Next() {
		b := book{}
		var row_state BookState
		rows.Scan(&b.Instance_id, &b.Name, &b.Year, &b.Publisher, &b.Author, &row_state)

			if row_state != state {
				log.Print("error state")
				return nil
			}


		bookSlice = append(bookSlice, b)

	}

	lenSlice := len(b) + 1
	queryInString += ", $" + strconv.FormatInt(int64(lenSlice), 10)
	log.Print(queryInString)
	updateIfaceSlice := make([]interface{}, 0, len(inBookIface) + 1)
	updateIfaceSlice = append(updateIfaceSlice, next_state)
	updateIfaceSlice = append(updateIfaceSlice, inBookIface...)

	urows, err := tx.Query("update book_instances set state = $1 where instance_id IN ( " + queryInString[3:] + ")" ,  updateIfaceSlice...)

	defer urows.Close()

	if err != nil {
		log.Print(err)
		return nil
	}


	return bookSlice
}