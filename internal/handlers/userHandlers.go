package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/toluhikay/fx-exchange/internal/dtos"
	customErrors "github.com/toluhikay/fx-exchange/internal/errors"
	"github.com/toluhikay/fx-exchange/internal/models"
	"github.com/toluhikay/fx-exchange/internal/services"
	"github.com/toluhikay/fx-exchange/pkg/jwt"
	"github.com/toluhikay/fx-exchange/pkg/utils"
)

type UserHandler struct {
	userService services.UserServiceImpl
	auth        jwt.Auth
}

func NewUserHandler(service services.UserServiceImpl, auth jwt.Auth) *UserHandler {
	return &UserHandler{
		userService: service,
		auth:        auth,
	}
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dtos.RegisterUser
	if err := utils.ReadJSON(w, r, &req); err != nil {
		fmt.Println("error reading data at create user handler with error: ", err)
		utils.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	user, err := uh.userService.CreateUser(r.Context(), req)
	if err != nil {
		utils.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	response := utils.JSONResponse{
		Error:   false,
		Message: "sign up successful",
		Data:    user,
	}

	utils.WriteJson(w, http.StatusCreated, response)

}

func (uh *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("user_claims").(*jwt.JwtClaims)

	fmt.Println(claims)

	id := claims.ID

	if id.String() == "" {
		fmt.Println("id not found")
		utils.ErrorJSON(w, customErrors.ErrUnauthorized, http.StatusUnauthorized)
	}

	user, err := uh.userService.GetUserById(r.Context(), id)
	if err != nil {
		fmt.Println(err, "at log - 01")
		utils.ErrorJSON(w, err, customErrors.ResolveHTTPStatus(err))
		return
	}

	jsonResponse := utils.JSONResponse{
		Error:   false,
		Message: "success",
		Data:    user,
	}

	utils.WriteJson(w, http.StatusOK, jsonResponse)
}

func (uh *UserHandler) UserLogin(w http.ResponseWriter, r *http.Request) {
	var req dtos.UserLogin
	if err := utils.ReadJSON(w, r, &req); err != nil {
		utils.ErrorJSON(w, errors.Join(customErrors.ErrInvalidPayload, err))
		return
	}

	user, err := uh.userService.GetUserByEmail(r.Context(), req)
	if err != nil {
		fmt.Println(err, "at log - 01")
		utils.ErrorJSON(w, err, customErrors.ResolveHTTPStatus(err))
		return
	}

	fmt.Println(user)

	tokenPairs, err := uh.auth.GenerateTokens(*user)
	if err != nil {
		utils.ErrorJSON(w, customErrors.ErrInternalServer)
	}

	response := utils.JSONResponse{
		Error:   false,
		Message: "sign in successful",
		Data: struct {
			AccessToken  string       `json:"access_token"`
			RefreshToken string       `json:"refresh_token"`
			Data         *models.User `json:"data"`
		}{
			AccessToken:  tokenPairs.AccessToken,
			RefreshToken: tokenPairs.RefreshToken,
			Data:         user,
		},
	}

	utils.WriteJson(w, http.StatusOK, response)

}
