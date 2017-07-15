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

func writeError(w http.ResponseWriter, b []byte) {
	log.Print("need authenticate")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(b)
}

func auth(f func(w http.ResponseWriter, r *http.Request)) (func(w http.ResponseWriter, r *http.Request)) {

	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			writeError(w, unauth("Authorization header is empty"))
			return
		}

		authValue := strings.TrimSpace(authHeader)

		if authValue == "" {
			writeError(w, unauth("Authorization header is empty"))
			return
		}

		authSlice := strings.Fields(authValue)

		if len(authSlice) != 2 || strings.ToLower(authSlice[0]) != authPrefix {
			//log.Printf("need %s header", authPrefix)
			writeError(w, unauth(fmt.Sprintf("need %s header", authPrefix)))
			return

		}

		authValue = authSlice[1]

		if authValue == "" {
			writeError(w, unauth("token is empty"))
			return
		}

		token, err := jwt.ParseWithClaims(authValue, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return signKey, nil
		})

		if err != nil {
			writeError(w, unauth(err.Error()))
			return
		}

		if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
			log.Print("name ", claims.name)
			log.Print("exp: ", time.Unix(claims.ExpiresAt,0))

		} else {
			writeError(w, unauth("invalid claims"))
			return
		}


		f(w, r)

	}
}



func getToken(w http.ResponseWriter, r *http.Request) {
	log.Print("enter getToken")


	inputData, err := ioutil.ReadAll(io.LimitReader(r.Body, maxReadLen))

	defer r.Body.Close()


	log.Print(err)
	if err != nil {
		log.Print(err)
		return
	}

	auth := authCredentials{}


	err = json.Unmarshal(inputData, &auth)


	if err != nil {
		log.Print(err)
		return
	}

	//
	//need validate login and password
	//

	log.Print("login: ", auth.Login, " password: ", auth.Password)


	if auth.Login == "" {
		log.Print("login param is empty")
		return
	}

	//password := r.PostFormValue("password")

	if auth.Password == "" {
		log.Print("password param is empty")
		return
	}



	expireTime := time.Now().Add(time.Hour * 24).Unix()


	claims := jwtClaims {
		name: auth.Login,
	}
	claims.ExpiresAt = expireTime

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)




	tokenString, err := token.SignedString(signKey)

	if err != nil {
		log.Print(err)
		return
	}
	tokenStruct := jwtToken{}
	tokenStruct.Jwt = tokenString

	outputData, err := json.Marshal(tokenStruct)

	if err != nil {
		log.Print(err)
		return
	}


	log.Print("token:", tokenString)
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(outputData))

}

