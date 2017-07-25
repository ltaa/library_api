package handlers

import (
	"net/http"
	"github.com/gorilla/mux"
	"strconv"
	"encoding/json"
	"io/ioutil"
	"io"
	"database/sql"
	"time"
	//"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	//"log"
	"log"
	//"github.com/pkg/errors"
	"errors"
	//"strings"
)

func cardQueryOperation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	//vars := r.URL.Query()

	idString:= vars["id"]

	//idString := strings.TrimSpace(vars.Get("id"))

	if idString == "" {
		level.Info(Config.Logger).Log("message", "idString is empty")
		return
	}

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		level.Error(Config.Logger).Log("error parsing id value")
		return
	}

	bookSlice, err := getCardProcessing(id, nil)
	if err != nil {
		level.Error(Config.Logger).Log(err)
		return
	}

	if err := writeData(w, http.StatusOK, marshalJson(bookSlice)); err != nil {
		log.Print(err)
	}
}

//func cardAddQueryValidate(idString string)  {
//
//}


func cardAddHandler(w http.ResponseWriter, r *http.Request) {
	muxVars := mux.Vars(r)

	idString := muxVars["id"]
	//id, err := strconv.Atoi(idString)

	//level.Error(Config.Logger).Log(r.URL.String())
	//vars := r.URL.Query()
	//idString := vars.Get("id")
	//idString = strings.TrimSpace(idString)
	if idString == "" {
		level.Error(Config.Logger).Log("message:", "id argument is empty")
		writeData(w, http.StatusBadRequest, marshalJson(emptyJson{}))
		return
	}
	id, err := strconv.Atoi(idString)


	if err != nil {
		level.Error(Config.Logger).Log("message:", "invalid request param value")
		writeData(w, http.StatusBadRequest, marshalJson(emptyJson{}))
		return
	}

	data, err := ioutil.ReadAll(io.LimitReader(r.Body, maxReadLen))
	defer r.Body.Close()

	if err != nil {
		level.Error(Config.Logger).Log("message:",err)
		writeData(w, http.StatusBadRequest, marshalJson(emptyJson{}))
		return
	}

	inBookSlice  := make([]bookId, 0, 1)

	if err := json.Unmarshal(data, &inBookSlice); err != nil {
		level.Error(Config.Logger).Log("message:",err)
		writeData(w, http.StatusBadRequest, marshalJson(emptyJson{}))
		return
	}

	if len(inBookSlice) == 0 {
		level.Error(Config.Logger).Log("message:", "input slice is empty")
		writeData(w, http.StatusBadRequest, marshalJson(emptyJson{}))
		return
	}

	bookSlice := make([]int, 0 , 1)
	for _,cardInstance := range inBookSlice {
		bookSlice = append(bookSlice, cardInstance.Instance_id)
	}

	tx, err := db.Begin()

	defer tx.Rollback()

	if err != nil {
		level.Error(Config.Logger).Log("message:",err)
		writeData(w, http.StatusInternalServerError, marshalJson(errorJson{Message: "invalid transaction"}))
		return
	}

	if _, err := tx.Exec(`set transaction isolation level serializable`); err != nil {
		level.Error(Config.Logger).Log("message:",err)
		writeData(w, http.StatusInternalServerError, marshalJson(errorJson{Message: "invalid transaction"}))
		return
	}

	state := StateChange{next: processing, curent: library, tx: tx}
	outBookSlice, err := bookChangeState(bookSlice, state)
	if err != nil {
		level.Error(Config.Logger).Log("message:",err)
		writeData(w, http.StatusInternalServerError, marshalJson(errorJson{Message: err.Error()}))

	}

	stmt, err := tx.Prepare("INSERT  INTO operations(instance_id, worker_id) VALUES ($1, $2)")

	for _, bookInstance := range outBookSlice {
		_, err := stmt.Exec(bookInstance.Instance_id, id)
		if err != nil {
			level.Error(Config.Logger).Log("message:",err)
			writeData(w, http.StatusInternalServerError, marshalJson(errorJson{Message: "invalid transaction"}))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		level.Error(Config.Logger).Log("message:",err)
		writeData(w, http.StatusInternalServerError, marshalJson(errorJson{Message: "invalid transaction"}))
		return
	}

	if err := writeData(w, http.StatusOK, marshalJson(messageJson{Message: "data added"})); err != nil {
		level.Error(Config.Logger).Log("message:",err)
	}
}


func cardCheckout(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	idString:= vars["id"]
	clientIdString := vars["clientId"]


	if idString == "" || clientIdString == "" {
		level.Error(Config.Logger).Log("invalid input params")
		//log.Println("invalid input params")
		return
	}

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		level.Error(Config.Logger).Log("error parsing id value")
		//log.Println("error parsing id value")
		return
	}

	clientId, err := strconv.ParseInt(clientIdString, 10, 64)
	if err != nil {
		level.Error(Config.Logger).Log("error parsing id value")
		//log.Println("error parsing id value")
		return
	}

	data, err:= ioutil.ReadAll(io.LimitReader(r.Body, maxReadLen))
	defer r.Body.Close()

	if err != nil {
		level.Error(Config.Logger).Log(err)
		//log.Print(err)
		return
	}

	inJsonSlice := make([]bookId, 0, 1)
	err = json.Unmarshal(data, &inJsonSlice)

	if err!= nil {
		level.Error(Config.Logger).Log(err)
		//log.Print(err)
		return
	}

	if len(inJsonSlice) == 0 {
		level.Error(Config.Logger).Log("cardCheckout: source json is empty")
		//log.Print("cardCheckout: source json is empty")
		return
	}

	inputIdSlice := make([]int, 0, len(inJsonSlice))

	for _, id := range inJsonSlice {
		inputIdSlice = append(inputIdSlice, id.Instance_id)
	}


	bookSlice, err := getCardProcessing(id, inputIdSlice)
	if err != nil {
		level.Error(Config.Logger).Log(err)
		return
	}


	instanceSlice := make([]int, 0, 1)
	for _, bookInstance := range bookSlice {
		instanceSlice = append(instanceSlice, bookInstance.Instance_id)

	}

	if instanceSlice == nil {
		return
	}
	tx, err := db.Begin()

	defer tx.Rollback()
	if err != nil {
		level.Error(Config.Logger).Log(err)
		//log.Print(err)
		return
	}


	_, err = tx.Exec(`set transaction isolation level serializable`)
	if err != nil {
		tx.Rollback()
		return
	}

	state := StateChange{next: client, curent: processing, tx: tx}

	_ , err = bookChangeState(instanceSlice, state)

	if err != nil {
		level.Error(Config.Logger).Log(err)
		writeData(w, http.StatusInternalServerError, marshalJson(errorJson{Message: err.Error()}))
	}

	cardCleanOperations(instanceSlice, tx)
	checkoutBooks(clientId, bookSlice, tx)

	err = tx.Commit()
	if err != nil {
		level.Error(Config.Logger).Log(err)
		//log.Print(err)
		return
	}
	w.WriteHeader(http.StatusOK)

}


func cardStateHandler(w http.ResponseWriter, r * http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)


	if err != nil {
		level.Error(Config.Logger).Log(err)
		//log.Print(err)
		return
	}
	a := author{}

	err = json.Unmarshal(data, &a)


	if err != nil {
		level.Error(Config.Logger).Log(err)
		//log.Print(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello"))

}


func cardCleanOperations(instanceSlice []int, tx *sql.Tx) (error){
	stmt, err := tx.Prepare(`DELETE FROM operations where operations.instance_id = $1`)

	if err != nil {
		return err
	}

	for _, instance := range instanceSlice {
		stmt.Exec(instance)
	}
	return nil
}


//need added return status
func checkoutBooks(clientId int64, bookSlice []book, tx *sql.Tx) {

	stmt, err := tx.Prepare(`INSERT INTO readers(client_id, instance_id, date_issue, return_date)
	VALUES ($1, $2, $3, $4)`)

	defer stmt.Close()

	if err != nil {
		level.Error(Config.Logger).Log(err)
		//log.Print(err)
		return
	}

	for _, b := range bookSlice {
		stmt.Exec(clientId, b.Instance_id, time.Now(), time.Now().Add(staticBookDuration))
	}

}


func getCardProcessing (id int64, clientSlice []int) ([]book, error){

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
          join book_instances on book_instances.book_id = books.book_id
          JOIN operations on operations.instance_id = book_instances.instance_id where operations.worker_id = $1`


	var prepareString string
	if(clientSlice != nil) {
		prepareString = getPrepareString(len(clientSlice), 2)
		queryString += " AND operations.instance_id IN ( " + prepareString + " ) ORDER BY instance_id"
	} else {
		queryString += " ORDER BY instance_id"
	}

	stmt, err := db.Prepare(queryString)

	if err != nil {
		//level.Error(Config.Logger).Log(err)
		//log.Println(err)
		return nil, err
	}

	var rows *sql.Rows
	if(clientSlice != nil) {
		dataInterface := make([]interface{}, 0, len(clientSlice) + 1)
		dataInterface = append(dataInterface, id)
		dataInterface = append(dataInterface, getPrepareInterface(clientSlice)...)

		rows,err = stmt.Query(dataInterface... )
	} else {
		rows,err = stmt.Query(id)
	}

	if err != nil {
		level.Error(Config.Logger).Log(err)
		//log.Print(err)
		return nil, err
	}

	bookSlice, err := getBookMap(rows, processing)
	if err != nil {
		return nil, err
	}

	if bookSlice == nil {
		return nil, errors.New("empty enities slice")
	}

	return bookSlice, nil

}



func cardGetList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, ok := vars["id"]

	if !ok {
		level.Error(Config.Logger).Log("error parsing arguments\n")
		//log.Print("error parsing arguments\n")
		return
	}

	rows, err := db.Query("select * from operation where worker_id = $1", id)

	if err != nil {
		level.Error(Config.Logger).Log(err)
		//log.Print(err)
		return
	}

	for rows.Next() {

		rows.Scan()
	}

}