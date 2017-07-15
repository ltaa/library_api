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
	"time"
)

func cardQueryOperation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	idString:= vars["id"]

	if idString == "" {
		log.Println("idString is empty")
		return
	}

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Println("error parsing id value")
		return
	}


	bookSlice := cardOperationDb(id, nil)



	b, err := json.Marshal(bookSlice)

	if err != nil {
		log.Print(err)
		return
	}

	//w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(b)


}




func cardAddHandler(w http.ResponseWriter, r *http.Request) {

	muxVars := mux.Vars(r)

	idString := muxVars["id"]
	defer r.Body.Close()
	data, err := ioutil.ReadAll(io.LimitReader(r.Body, maxReadLen))

	if err != nil {
		log.Print(err)
		return
	}

	inBookSlice  := make([]bookId, 0, 1)

	err = json.Unmarshal(data, &inBookSlice)

	if err != nil {
		log.Print(err)
		return
	}

	bookSlice := make([]int, 0 , 1)
	for _,cardInstance := range inBookSlice {
		bookSlice = append(bookSlice, cardInstance.Instance_id)
	}

	id, err := strconv.Atoi(idString)

	if err != nil {
		log.Printf("invalid request param value")
		return
	}



	tx, err := db.Begin()
	//defer tx.Rollback()

	if err != nil {
		log.Print(err)
		return
	}

	outBookSlice := bookRequestDb(bookSlice, library, tx)

	stmt, err := tx.Prepare("INSERT  INTO operations(instance_id, worker_id) VALUES ($1, $2)")

	for _, bookVal := range bookSlice {
		_, err := stmt.Exec(bookVal, id)
		if err != nil {
			log.Print(err)
		}
	}



	err = tx.Commit()
	if err != nil {
		log.Print(err)
		return
	}

	//outBookSlice := make([]*books, 0, 1)
	//
	//for _, b := range outbooksMap  {
	//	outBookSlice = append(outBookSlice, b)
	//}

	outData, err := json.Marshal(&outBookSlice)

	if err != nil {
		log.Print(err)
		return
	}

	//w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(outData)

}




func cardCheckout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	idString:= vars["id"]
	clientIdString := vars["clientId"]


	if idString == "" || clientIdString == "" {
		log.Println("invalid input params")
		return
	}

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Println("error parsing id value")
		return
	}

	clientId, err := strconv.ParseInt(clientIdString, 10, 64)
	if err != nil {
		log.Println("error parsing id value")
		return
	}

	data, err:= ioutil.ReadAll(io.LimitReader(r.Body, maxReadLen))
	defer r.Body.Close()

	if err != nil {
		log.Print(err)
		return
	}

	inJsonSlice := make([]bookId, 0, 1)
	err = json.Unmarshal(data, &inJsonSlice)

	if err!= nil {
		log.Print(err)
		return
	}

	if len(inJsonSlice) == 0 {
		log.Print("cardCheckout: source json is empty")
		return
	}

	inputIdSlice := make([]int, 0, len(inJsonSlice))

	for _, id := range inJsonSlice {
		inputIdSlice = append(inputIdSlice, id.Instance_id)
	}



	bookSlice := cardOperationDb(id, inputIdSlice)


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
		log.Print(err)
		return
	}

	bookRequestDb(instanceSlice, processing, tx)
	cardCleanOperations(instanceSlice, tx)


	checkoutBooks(clientId, bookSlice, tx)

	err = tx.Commit()
	if err != nil {
		log.Print(err)
		return
	}


	//w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

}




func cardStateHandler(w http.ResponseWriter, r * http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)


	if err != nil {
		log.Print(err)
		return
	}
	a := author{}

	err = json.Unmarshal(data, &a)


	if err != nil {
		log.Print(err)
		return
	}





	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello"))

}





func cardCleanOperations(instanceSlice []int, tx *sql.Tx) {


	var queryString string
	var interfaceSlice = make([]interface{}, 0, 1)
	for instanceIdx := range instanceSlice {
		queryString += "$" + strconv.Itoa(instanceIdx + 1)
		interfaceSlice = append(interfaceSlice, instanceSlice[instanceIdx])

		if instanceIdx + 1 < len(instanceSlice) {
			queryString += ", "
		}

	}


	tx.Exec(`
		DELETE FROM operations where operations.instance_id IN (` + queryString + ")", interfaceSlice...)

}





//need added return status
func checkoutBooks(clientId int64, bookSlice []book, tx *sql.Tx) {

	stmt, err := tx.Prepare(`INSERT INTO readers(client_id, instance_id, date_issue, return_date)
	VALUES ($1, $2, $3, $4)`)

	defer stmt.Close()

	if err != nil {
		log.Print(err)
		return
	}

	for _, b := range bookSlice {
		stmt.Exec(clientId, b.Instance_id, time.Now(), time.Now().Add(staticBookDuration))
	}

}




func cardOperationDb (id int64, clientSlice []int) ([]book){

	//queryString := `select book_instances.instance_id,
	//	books.book_name,
	//	books.year,
	//	authors.first_name,
	//	authors.last_name,
	//	publishers.publisher_name
	//
	//      from
	//	book_instances left join books on book_instances.book_id = books.book_id
	//	join publishers on books.publisher_id = publishers.publisher_id
	//	join authors_books on books.book_id = authors_books.book_id
	//	join authors on authors.author_id = authors_books.author_id
	//	JOIN operations on operations.instance_id = book_instances.instance_id where operations.worker_id = $1`



	queryString := `select book_instances.instance_id,
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
          JOIN operations on operations.instance_id = book_instances.instance_id where operations.worker_id = $1`






	//book_instances.state

	var prepareString string
	if(clientSlice != nil) {
		prepareString = getPrepareString(len(clientSlice), 2)
		log.Print("prepareString = ", prepareString)
		queryString += " AND operations.instance_id IN ( " + prepareString + " )"
	}

	stmt, err := db.Prepare(queryString)

	if err != nil {
		log.Println(err)
		return nil
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
		log.Print(err)
		return nil
	}

	bookSlice := getBookMap(rows)
	if bookSlice == nil {
		return nil
	}

	//bookSlice := make([]*books, 0, 1)
	//
	//for _, bookInstance := range booksMap {
	//	bookSlice = append(bookSlice, bookInstance)
	//}

	return bookSlice

}



func cardGetList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, ok := vars["id"]

	if !ok {
		log.Print("error parsing arguments\n")
		return
	}

	rows, err := db.Query("select * from operation where worker_id = $1", id)

	if err != nil {
		log.Print(err)
		return
	}

	for rows.Next() {

		rows.Scan()
	}

}