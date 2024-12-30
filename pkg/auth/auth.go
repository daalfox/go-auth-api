package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/daalfox/go-auth-microservice/internal/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func NewAuthService(db *gorm.DB) AuthService {
	db.AutoMigrate(&User{})

	a := AuthService{
		db:     db,
		Router: chi.NewRouter(),
	}

	a.mountHandlers()

	return a
}

type AuthService struct {
	db     *gorm.DB
	Router *chi.Mux
}

func (a *AuthService) mountHandlers() {
	a.Router.Use(chiMiddleware.Logger)
	a.Router.Use(middleware.Json)

	a.Router.Post("/register", a.Register)
	a.Router.Post("/login", a.Login)
}

func (a *AuthService) Register(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	issues := user.validate()
	if len(issues) != 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&issues)
		return
	}

	pw, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(pw)

	result := a.db.Create(&user)
	if result.Error == gorm.ErrDuplicatedKey {
		w.WriteHeader(http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *AuthService) Login(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	issues := user.validate()
	if len(issues) != 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&issues)
		return
	}

	rawPassword := user.Password

	result := a.db.Where("username = ?", user.Username).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(rawPassword))

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": user.ID,
	})

	token, err := t.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}

	json.NewEncoder(w).Encode(&token)
}