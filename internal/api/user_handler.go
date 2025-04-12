package api

import (
	"encoding/json"
	"errors"
	"femProject/internal/store"
	"femProject/internal/utils"
	"log"
	"net/http"
	"regexp"
)

type registerUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

type UserHandler struct {
	userStore store.UserStore
	logger    *log.Logger
}

func NewUserHandler(userStore store.UserStore, logger *log.Logger) *UserHandler {
	return &UserHandler{
		userStore: userStore,
		logger:    logger,
	}
}

func (h *UserHandler) validateRegisterUserRequest(request *registerUserRequest) error {
	if request.Username == "" {
		return errors.New("username is required")
	}
	if request.Email == "" {
		return errors.New("email is required")
	}
	if request.Password == "" {
		return errors.New("password is required")
	}

	if len(request.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}
	if len(request.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(request.Email) {
		return errors.New("invalid email format")
	}

	return nil
}

func (h *UserHandler) HandleRegisterUser(w http.ResponseWriter, r *http.Request) {
	var request registerUserRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		h.logger.Printf("ERROR: decoding register user request: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid payload"})
		return
	}

	err = h.validateRegisterUserRequest(&request)
	if err != nil {
		h.logger.Printf("ERROR: validating register user request: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}

	user := &store.User{
		Username: request.Username,
		Email:    request.Email,
	}

	if request.Bio != "" {
		user.Bio = request.Bio
	}

	err = user.PasswordHash.Set(request.Password)
	if err != nil {
		h.logger.Printf("ERROR: hashing password: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	err = h.userStore.CreateUser(user)
	if err != nil {
		h.logger.Printf("ERROR: register user: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"user": user})
}
