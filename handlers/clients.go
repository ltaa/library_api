package handlers

import (
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"strconv"
	"encoding/json"
	"io/ioutil"
	"io"
	"database/sql"
)

func readerDataList(w http.ResponseWriter, r *http.Request) {
	argsMap := mux.Vars(r)

	clientString, ok := argsMap["clientId"]
	if !ok {
		log.Print("invalid arguments")
		return
	}

	clientId, err := strconv.Atoi(clientString)

	if err != nil {
		log.Print(err)
		return
	}

	bookSlice := readerDataListDb(clientId)

	if bookSlice == nil {
		log.Print("db request error")
		return
	}

	data, err := json.Marshal(bookSlice)

	if err != nil {
		log.Print(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)

}





func clientReturnBooks(w http.ResponseWriter, r *http.Request) {

	inputData, err := ioutil.ReadAll(io.LimitReader(r.Body, maxReadLen))

	defer r.Body.Close()

	if err != nil {
		log.Print(err)
		return
	}

	inputBookSlice := make([]bookId, 0, 1)

	err = json.Unmarshal(inputData, &inputBookSlice)

	if err != nil {
		log.Print(err)
		return
	}



	inputSlice := make([]int, 0, len(inputBookSlice))

	for _, instance := range inputBookSlice {
		inputSlice = append(inputSlice, instance.Instance_id)
	}


	tx, err := db.Begin()
	defer tx.Rollback()

	if err != nil {
		log.Print(err)
		return
	}


	outputSlice := bookRequestDb(inputSlice, client, tx)

	if outputSlice == nil {
		log.Print("error set state to library")
		return
	}

	err = removeReader(inputSlice, tx)
	if err != nil {
		log.Print(err)
		return
	}

	tx.Commit()

	//outputSlice := bookMapToSlice(bookMap)

	outputData, err := json.Marshal(outputSlice)

	if err != nil {
		log.Print(err)
		return
	}


	w.WriteHeader(http.StatusOK)
	w.Write(outputData)


}


func removeReader(inputSlice []int, tx *sql.Tx) (error) {

	stmt, err := tx.Prepare(`DELETE from readers where instance_id = $1`)
	defer stmt.Close()
	if err != nil {
		log.Print(err)
		return err
	}
	for _, instance := range inputSlice {
		_, err = stmt.Exec(instance)
		if err != nil {
			return err
		}
	}

	return nil

}


func readerDataListDb (clientId int) []book{

	//rows, err := db.Query(`select book_instances.instance_id,
	//books.book_name,
	//	books.year,
	//	authors.first_name,
	//	authors.last_name,
	//	publishers.publisher_name
	//
	//from
	//book_instances left join books on book_instances.book_id = books.book_id
	//join publishers on books.publisher_id = publishers.publisher_id
	//join authors_books on books.book_id = authors_books.book_id
	//join authors on authors.author_id = authors_books.author_id
	//join readers on readers.instance_id = book_instances.instance_id where readers.client_id = $1`, clientId)
	//


	rows, err := db.Query(`select book_instances.instance_id,
        books.book_name,
    		books.year,
        publishers.publisher_name,
        authors.authors
        from
        (
          select array_agg(concat_ws('_',a.first_name,a.last_name)) as authors,
          b.book_id as book_id
            from
              authors as a join authors_books b on a.author_id=b.author_id group by b.book_id
        ) authors
          join books on books.book_id = authors.book_id
          join publishers on books.publisher_id = publishers.publisher_id
          join book_instances on book_instances.book_id = books.book_id
          join readers on readers.instance_id = book_instances.instance_id where readers.client_id = $1`, clientId)








	// 		book_instances.state

	if err != nil {
		log.Print(err)
		return nil
	}

	bookSlice := getBookMap(rows)

	//bookSlice := make([]*books, 0 , 1)
	//
	//for _, instance := range bookMap {
	//	bookSlice = append(bookSlice, instance)
	//}

	return bookSlice

}