package handlers

import (
	"log"
	"github.com/dgrijalva/jwt-go"
	"fmt"
	"net/http"
	"strings"
	"encoding/json"
	"io/ioutil"
	"io"
	"time"
	"github.com/gorilla/securecookie"
	"github.com/go-kit/kit/log/level"
	//"github.com/pkg/errors"
	"errors"
)


func init() {
	signKey = securecookie.GenerateRandomKey(secureKeyLen)
}




func unauth(message string) []byte {
	u := unauthJson{Error: "auth error", Message: message}
	data, err := json.Marshal(u)

	if err != nil {
		log.Print(err)
		return nil
	}

	return data
}

func writeAuthError(w http.ResponseWriter, b []byte) {
	log.Print("need authenticate")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(b)
}



func auth(f func(w http.ResponseWriter, r *http.Request)) (func(w http.ResponseWriter, r *http.Request)) {

	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			writeAuthError(w, unauth("Authorization header is empty"))
			return
		}

		authValue := strings.TrimSpace(authHeader)

		if authValue == "" {
			writeAuthError(w, unauth("Authorization header is empty"))
			return
		}

		authSlice := strings.Fields(authValue)

		if len(authSlice) != 2 || strings.ToLower(authSlice[0]) != authPrefix {
			writeAuthError(w, unauth(fmt.Sprintf("need %s header", authPrefix)))
			return

		}

		authValue = authSlice[1]

		if authValue == "" {
			writeAuthError(w, unauth("token is empty"))
			return
		}

		token, err := jwt.ParseWithClaims(authValue, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return signKey, nil
		})

		if err != nil {
			writeAuthError(w, unauth(err.Error()))
			return
		}

		if _, ok := token.Claims.(*jwtClaims); ok && token.Valid {
			//log.Print("name ", claims.name)
			//log.Print("exp: ", time.Unix(claims.ExpiresAt,0))

		} else {
			writeAuthError(w, unauth("invalid claims"))
			return
		}

		err = level.Error(Config.Logger).Log("error:", "handle request from host " + r.URL.Host)

		f(w, r)

	}
}



func getTokenInputValidate(inputData []byte) (*authCredentials, error) {
	auth := new (authCredentials)
	err := json.Unmarshal(inputData, &auth)

	if err != nil {
		//log.Print(err)
		return nil, err
	}


	auth.Login = strings.TrimSpace(auth.Login)
	auth.Password = strings.TrimSpace(auth.Password)


	log.Print("login: ", auth.Login, " password: ", auth.Password)

	if auth.Login == "" {
		return nil, errors.New("login is empty")
	}

	if auth.Password == "" {
		return nil, errors.New("password is empty")
	}
	return auth, nil
}


func getTokenCredentialsValidate(auth *authCredentials ) (int,error) {
	queryString := `SELECT worker_id, (password = crypt($1, password)) from workers where login = $2`

	//queryString =  `SELECT (password = crypt($1, password)) FROM workers where login = $2`
	//var login string
	//var passwd string


	var validFlag bool
	var worker_id int
	err := db.QueryRow(queryString, auth.Password, auth.Login).Scan(&worker_id, &validFlag)

	if err != nil {
		return 0, errors.New("invalid credentials")
		//return err
	}



	if validFlag {
		return worker_id, nil
	} else {
		return 0,errors.New("invalid credentials")
	}


}

func getTokenInstance(auth *authCredentials, worker_id int) (*jwtToken, error) {

	expireTime := time.Now().Add(time.Hour * 24).Unix()
	claims := jwtClaims {
		name: auth.Login,
	}
	claims.ExpiresAt = expireTime

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenString, err := token.SignedString(signKey)

	if err != nil {
		log.Print(err)
		return nil, err
	}

	tokenStruct := new (jwtToken)
	tokenStruct.Jwt = tokenString
	tokenStruct.Worker_id = worker_id

	return tokenStruct, nil

}

func getToken(w http.ResponseWriter, r *http.Request) {


	inputData, err := ioutil.ReadAll(io.LimitReader(r.Body, maxReadLen))

	defer r.Body.Close()

	log.Print(err)
	if err != nil {
		log.Print(err)
		writeData(w, http.StatusBadRequest, marshalJson(emptyJson{}))
		return
	}

	auth, err := getTokenInputValidate(inputData)
	if err != nil {
		log.Print(err)
		return
	}


	id, err := getTokenCredentialsValidate(auth)
	if err != nil {
		log.Print(err)
		writeData(w, http.StatusForbidden, marshalJson(errorJson{Message: err.Error()}))
		return
	}

	token, err := getTokenInstance(auth, id)

	if err != nil {
		log.Print(err)
		writeData(w, http.StatusForbidden, marshalJson(errorJson{Message: err.Error()}))
		return
	}


	if err := writeData(w, http.StatusOK, marshalJson(token)); err != nil {
		log.Print(err)
		writeData(w, http.StatusInternalServerError, marshalJson(errorJson{Message: err.Error()}))
	}

}

