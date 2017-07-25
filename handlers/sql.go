package handlers

import (
	"database/sql"
	"log"
	_ "github.com/lib/pq"
	"strconv"
	"fmt"
	//"github.com/pkg/errors"
	"errors"
)

func init()  {
	var err error;

	db, err = sql.Open("postgres", "postgres://postgres:postgres@library-db:5432/library?sslmode=disable")

	if err != nil {
		log.Print(err)
		panic("db init error")
	}
}


func getPrepareString(stringLen int, start int) string {

	var request string
	for i := 0; i < stringLen; i++ {
		request += "$" + strconv.Itoa(start + i)
		if (i + 1 < stringLen) {
			request +=", "
		}
	}

	return request

}


func getPrepareInterface(data []int) []interface{} {
	dataInterfaces := make([]interface{}, 0, len(data))

	for idx := range data {
		dataInterfaces = append(dataInterfaces, data[idx])
	}

	return dataInterfaces
}


func getBookMap(rows *sql.Rows, state BookState) ([]book, error) {

	bookSlice := make([]book, 0, 1)

	for rows.Next() {
		b := book{}
		var row_state BookState
		if err := rows.Scan(&b.Instance_id, &b.Name, &b.Year, &b.Publisher, &b.Author, &row_state); err != nil {
			return nil, err
		}

		if state != showAll && row_state != state {
			//log.Print("error state")
			return nil, errors.New("invalid data state")
		}

		bookSlice = append(bookSlice, b)

	}
	return  bookSlice, nil

}



func getBookChangeState(rows *sql.Rows, state BookState) ([]book, error) {

	bookSlice := make([]book, 0, 1)

	for rows.Next() {
		b := book{}
		var row_state BookState

		if err := rows.Scan(&b.Instance_id, &b.Name, &b.Year, &b.Publisher, &b.Author, &row_state); err != nil {
			return nil, err
		}

		if row_state != state{
			return nil, errors.New("invalid data state")
		}

		bookSlice = append(bookSlice, b)
	}
	return  bookSlice, nil
}


func createBookDb(b *book) {


	book_id := createBookRelations(b)


	_, err := db.Exec(`insert into book_instances (book_id, state) VALUES($1, $2)`, book_id, library)

	if err != nil {
		log.Print(err)
		return
	}

}


func createBookRelations(b *book) (int) {
	tx, err := db.Begin()

	defer tx.Rollback()
	rows, err := tx.Query("select publishers.publisher_id from publishers where publishers.publisher_name = $1", b.Publisher)

	if err != nil {
		log.Print(err)
		return 0
	}

	var publisher_id int

	if rows.Next() {
		err = rows.Scan(&publisher_id)
	} else {
		err = tx.QueryRow("insert into publishers (publisher_name ) VALUES ($1) RETURNING publisher_id", b.Publisher).Scan(&publisher_id)
	}

	if err != nil {
		log.Print(err)
		return 0
	}

	rows.Close()
	rows, err = tx.Query(`select books.book_id from books where books.book_name = $1 AND books.year = $2 and books.publisher_id = $3`, b.Name, b.Year, publisher_id)


	if err != nil {
		log.Print(err)
		return 0
	}

	var book_id int

	if rows.Next() {
		err = rows.Scan(&book_id)
	} else {
		err = tx.QueryRow(`insert into books (book_name, year, publisher_id) VALUES ($1, $2, $3) RETURNING book_id`, b.Name, b.Year, publisher_id).Scan(&book_id)
	}

	if err != nil {
		log.Print(err)
		return 0
	}

	rows.Close()

	authorsId := make([]int, 0, len(b.Author))

	stmt, err := tx.Prepare(`select authors.author_id from authors where authors.first_name = $1 AND authors.last_name = $2`)

	for _, a := range b.Author {
		rows, err := stmt.Query(a.FirstName, a.LastName)
		if err != nil {
			log.Print(err)
			continue
		}

		var a_tmp int

		//need fix, if namesakes
		if rows.Next() {
			err = rows.Scan(&a_tmp)
			if err != nil {
				log.Print(err)
				continue

			}

		} else {
			err = tx.QueryRow(`insert into authors (first_name, last_name ) VALUES ($1, $2) RETURNING author_id`, a.FirstName, a.LastName).Scan(&a_tmp)
			if err != nil {
				log.Print(err)
				continue
			}

		}
		authorsId = append(authorsId, a_tmp)
		rows.Close()
	}

	if err != nil {
		log.Print(err)
		return 0

	}

	rows, err = tx.Query(`select book_id from authors_books where book_id = $1 `, book_id)

	if err != nil {
		log.Print(err)
		return 0
	}



	if !rows.Next() {
		for _, id := range authorsId {
			_, err := tx.Exec(`insert into authors_books(author_id,book_id) VALUES ($1, $2)`, id, book_id);
			if err != nil {
				log.Print(err)
				return 0
			}

		}

	}

	rows.Close()
	tx.Commit()
	return book_id

}


func updateBookInstance(b *book) (error) {
	book_id := createBookRelations(b)

	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Print(err)
		return fmt.Errorf("error starting transaction")
	}
	_, err = tx.Exec(`update book_instances set book_id = $1 where instance_id = $2`, book_id, b.Instance_id)

	if err != nil {
		log.Print(err)
		return fmt.Errorf("error updating instance")
	}

	tx.Commit()
	return nil
}


func deleteBookInstance(bSlice []book) (error){

	var instance_count int

	tx, err := db.Begin()
	defer tx.Rollback()

	if err != nil {
		log.Print(err)
		return fmt.Errorf("cannot get transaction descriptor")
	}
	for _, b := range bSlice {
		err := db.QueryRow(`select COUNT(instance_id) from book_instances where instance_id = $1`, b.Instance_id).Scan(&instance_count)

		if err == sql.ErrNoRows {
			log.Print("instance not exist")
			return fmt.Errorf("instance not exist")
		}

		if err != nil {
			log.Print(err)
			return fmt.Errorf("querry row call error")
		}

		_, err = db.Exec(`delete from book_instances where instance_id = $1`, b.Instance_id)
		if err != nil {
			log.Print(err)
			return fmt.Errorf("delete call error")
		}

	}

	tx.Commit()
	return nil
}



