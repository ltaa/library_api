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
	"strings"
	//"github.com/pkg/errors"
	"errors"
)

func registerClient(w http.ResponseWriter, r *http.Request) {
	inputData, err := ioutil.ReadAll(io.LimitReader(r.Body, maxReadLen))

	if err != nil {
		log.Print(err)
		return
	}

	clientInstance := clientJson{}
	err = json.Unmarshal(inputData, &clientInstance)

	if err != nil {
		log.Print(err)
		return
	}

	if clientInstance.First_name == "" || clientInstance.Last_name == "" {
		log.Print("clients name is empty")
		return
	}
	_, err = db.Query(`insert into clients(first_name, last_name) VALUES($1, $2)`, clientInstance.First_name, clientInstance.Last_name)

	if err != nil {
		log.Print(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	//w.Write()


}

func clientsListNameQuery(nameString string) ([]clientJson, error) {
	clientInstance := make([]clientJson, 0, 8)
	queryString := `SELECT client_id, first_name, last_name from clients where LOWER(last_name) LIKE LOWER( '%' || $1 || '%')`

	rows, err := db.Query(queryString, nameString)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		newClient := clientJson{}
		err := rows.Scan(&newClient.Client_id, &newClient.First_name, &newClient.Last_name)

		if err != nil {
			//log.Print(err)
			//continue
			return nil, err
		}

		clientInstance = append(clientInstance, newClient)
	}

	if err != nil {
		return nil, err
	}

	return clientInstance, nil

}

func clientsListIdQuery(idString string) ([]clientJson, error) {
	clientInstance := make([]clientJson, 0, 8)

	if idString != "" {
		id, err := strconv.Atoi(idString)
		if err != nil {
			log.Print(err)
			return  nil, err
		}
		queryString := `SELECT client_id, first_name, last_name from clients where client_id = $1`
		newClient := clientJson{}
		err = db.QueryRow(queryString, id).Scan(&newClient.Client_id, &newClient.First_name, &newClient.Last_name)

		if err != nil {
			//log.Print(err)
			return nil, err
		}

		clientInstance = append(clientInstance, newClient)
	}

	return clientInstance, nil
}

func clientsListQuery() ([]clientJson, error) {
	clientInstance := make([]clientJson, 0, 8)

	queryString := `SELECT client_id, first_name, last_name from clients`
	rows, err := db.Query(queryString)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		newClient := clientJson{}
		err := rows.Scan(&newClient.Client_id, &newClient.First_name, &newClient.Last_name)

		if err != nil {
			//log.Print(err)
			//continue
			return nil, err
		}
		clientInstance = append(clientInstance, newClient)
	}

	if err != nil {
		return nil, err
	}
	return clientInstance, nil
}

func getClientsList(w http.ResponseWriter, r *http.Request) {

	vars := r.URL.Query()
	idString := vars.Get("id")
	idString = strings.TrimSpace(idString)

	nameString := vars.Get("name")
	nameString = strings.TrimSpace(nameString)

	var err error

	clientInstance := make([]clientJson, 0, 1)

	if idString != "" {

		clientInstance, err = clientsListIdQuery(idString)
		if err != nil {
			log.Print(err)
			return
		}

	} else if nameString != "" {
		clientInstance, err = clientsListNameQuery(nameString)
		if err != nil {
			log.Print(err)
			return
		}

	} else {
		clientInstance, err = clientsListQuery()
		if err != nil {
			log.Print(err)
			return
		}

	}

	writeData(w, http.StatusOK, marshalJson(clientInstance))

}


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

	bookSlice, err := getReaderList(clientId)

	if err != nil {
		log.Print(err)
		return
	}

	writeData(w, http.StatusOK, marshalJson(bookSlice))

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

	_, err = tx.Exec(`set transaction isolation level serializable`)
	if err != nil {
		tx.Rollback()
		return
	}

	state := StateChange{next: library, curent: client, tx: tx }
	outputSlice, err := bookChangeState(inputSlice, state)

	if err != nil {
		log.Print(err)
		return
	}

	if err := removeReader(inputSlice, tx); err != nil {
		log.Print(err)
		return
	}

	tx.Commit()

	writeData(w, http.StatusOK, marshalJson(outputSlice))

}


func removeReader(inputSlice []int, tx *sql.Tx) (error) {

	stmt, err := tx.Prepare(`DELETE from readers where instance_id = $1`)
	defer stmt.Close()
	if err != nil {
		return err
	}

	for _, instance := range inputSlice {
		if _, err := stmt.Exec(instance) ; err != nil {
			return err
		}
	}

	return nil

}


func getReaderList (clientId int) ([]book, error){

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
          join book_instances on book_instances.book_id = books.book_id
          join readers on readers.instance_id = book_instances.instance_id where readers.client_id = $1`, clientId)


	if err != nil {
		return nil, err
	}

	bookSlice, err := getBookMap(rows, client)
	if err != nil {
		return nil, err
	}
	if bookSlice == nil {
		return nil, errors.New("entities slice is empty")
	}

	return bookSlice, nil

}