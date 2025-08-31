package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	customError "github.com/toluhikay/fx-exchange/internal/errors"
	"github.com/toluhikay/fx-exchange/internal/models"
)

type Auth struct {
	Issuer               string
	Audience             string
	Secret               string
	TokenExpireAt        time.Duration
	RefreshTokenExpireAt time.Duration
}

func NewAuth(auth Auth) *Auth {
	return &Auth{
		Issuer:               auth.Issuer,
		Audience:             auth.Audience,
		Secret:               auth.Secret,
		TokenExpireAt:        auth.TokenExpireAt,
		RefreshTokenExpireAt: auth.RefreshTokenExpireAt.Abs(),
	}
}

type JwtClaims struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	jwt.RegisteredClaims
}

type TokenPairs struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (a *Auth) GetTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *JwtClaims, error) {

	// add token to header
	w.Header().Add("Vary", "Authorization")

	tokenFromHeader := r.Header.Get("Authorization")
	if tokenFromHeader == "" {
		return "", &JwtClaims{}, customError.ErrUnauthorized
	}

	tokenParts := strings.Split(tokenFromHeader, " ")
	if len(tokenParts) != 2 {
		return "", &JwtClaims{}, customError.ErrUnauthorized

	}

	if tokenParts[0] != "Bearer" {
		return "", &JwtClaims{}, customError.ErrUnauthorized
	}

	token := tokenParts[1]

	claims := &JwtClaims{}

	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		// check for the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method expected: " + token.Header["alg"].(string))
		}
		return []byte(a.Secret), nil
	})

	if err != nil {
		if strings.HasPrefix(err.Error(), "token expired by") {
			return "", nil, errors.New("token expired")
		}
		return "", nil, err
	}

	if claims.Issuer != a.Issuer {
		return "", nil, errors.New("issuer not verified")
	}

	return token, claims, nil
}

func (a *Auth) GenerateTokens(u models.User) (TokenPairs, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":   a.Issuer,
		"email": fmt.Sprintf("%s ", u.Email),
		"id":    u.ID,
		"aud":   a.Audience,
		"sub":   fmt.Sprint(u.ID),
		"exp":   time.Now().Add(a.TokenExpireAt).Unix(),
	})

	accessToken, err := token.SignedString([]byte(a.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: a.Issuer,
		Audience: jwt.ClaimStrings{
			a.Audience,
		},
		Subject: a.Issuer,
		ID:      u.ID.String(),
	})

	refreshToken, err := token.SignedString([]byte(a.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	tokenPairs := TokenPairs{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return tokenPairs, nil
}
