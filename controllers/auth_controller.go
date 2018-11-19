package controllers

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gosu-team/cfapp-api/lib"
	"github.com/gosu-team/cfapp-api/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthParams ...
type AuthParams struct {
	Email    string `json:"email" gorm:"not null; type:varchar(250); unique_index"`
	Password string `json:"password" gorm:"not null; type:varchar(250)"`
}

// AuthResponse ...
type AuthResponse struct {
	JWT       string `json:"jwt"`
	ExpiresAt int64  `json:"expire_at"`
	ID        int    `json:"id"`
}

// Verify password
func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

// LoginHandler ...
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	req := lib.Request{ResponseWriter: w, Request: r}
	res := lib.Response{ResponseWriter: w}

	authParams := new(AuthParams)
	req.GetJSONBody(authParams)

	user := models.User{
		Email: authParams.Email,
	}

	if err := user.FetchByEmail(); err != nil {
		res.SendBadRequest("User not found")
		return
	}

	verifyPassword := comparePasswords(user.Password, []byte(authParams.Password))

	if !verifyPassword {
		res.SendBadRequest("Password is wrong")
		return
	}

	mySigningKey := []byte(os.Getenv("JWT_SECRET"))
	expireTime := time.Now().Add(time.Hour * 24 * 1).Unix()

	// Create the Claims
	claims := &jwt.StandardClaims{
		ExpiresAt: expireTime,
		Id:        strconv.Itoa(user.ID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)

	if err != nil {
		res.SendBadRequest("unknown error")
		return
	}

	jwtResponse := AuthResponse{
		JWT:       ss,
		ExpiresAt: expireTime,
		ID:        user.ID,
	}

	res.SendOK(jwtResponse)
}
