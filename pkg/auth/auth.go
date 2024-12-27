package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	a.Router.Use(middleware.Logger)

	a.Router.Post("/register", a.Register)
}

func (a *AuthService) Register(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	issues := user.validate()
	if len(issues) != 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&issues)
		return
	}

	a.db.Create(&user)

	w.WriteHeader(http.StatusCreated)
}

type User struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u *User) validate() []string {
	issues := []string{}
	if u.Username == "" {
		issues = append(issues, "`username` is required")
	}
	if u.Password == "" {
		issues = append(issues, "`password` is required")
	}

	return issues
}
