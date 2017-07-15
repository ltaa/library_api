package handlers

import (
	"database/sql"
	"github.com/dgrijalva/jwt-go"
	"time"
	"database/sql/driver"
	"errors"
	"strings"
	"log"
)

type author struct{
	Id int `json:"id"`
	First_name string `json:"first_name"`
	Last_name string `json:"last_name"`

}

type getCard struct {
	Id int `json:"id"`
	books []int
}

//type books struct {
//	Instance_id int `json:"instance_id"`
//	Name string `json:"name"`
//	Year int `json:"year"`
//	Author []string `json:"author"`
//	Publisher string `json:"publisher"`
//}

type bookId struct{
	Instance_id int `json:"instance_id"`
}



type jwtToken struct {
	Jwt string `json:"jwt"`
}

type authCredentials struct {
	Login string `json:"login"`
	Password string `json:"password"`
}

type jwtClaims struct {
	name string `json:"name"`
	jwt.StandardClaims
}

type unauthJson struct {
	Error string `json:"error"`
	Message string `json:"message"`
}

type createBookJson struct {
	Message string `json:"message"`
}

type updateBookJson struct {
	Message string `json:"message"`
}


var db *sql.DB
const maxReadLen = 1048576

const staticBookDuration = 20 * time.Hour * 24
const secureKeyLen = 128

var signKey[]byte

var authPrefix = "bearer"



type BookState string

const (
	library BookState = "library"
	processing = "processing"
	client = "client"

)


func (s *BookState) Scan(value interface{}) error {
	byteValue, ok := value.([]byte)

	if !ok {
		return errors.New("scaned value is not []byte")
	}
	*s = BookState(string(byteValue))

	return nil
}


func (s BookState) Value() (driver.Value, error) {
	return string(s), nil
}


type bookAuthor struct {
	FirstName string	`json:"first_name"`
	LastName string		`json:"last_name"`
}


type bookAuthors []bookAuthor

func (a *bookAuthors) Scan(value interface{}) error {
	byteSlice, ok := value.([]byte)

	s := string(byteSlice)
	log.Print(s)

	if !ok {
		return errors.New("scaned value is not []byte")
	}

	//s := string(byteValue)

	s = strings.Trim(s, "{}")

	autorsSlice := strings.Split(s, ",")
	b := make([]bookAuthor, 0, len(autorsSlice))

	for _, fullname := range autorsSlice {
		nameSlice := strings.Split(fullname, "_")
		b = append(b, bookAuthor{FirstName: nameSlice[0], LastName: nameSlice[1]})
	}

	*a = b

	return nil
}


//func (s *authors) Value() (driver.Value, error) {
//	authors.
//
//	return string(s), nil
//}