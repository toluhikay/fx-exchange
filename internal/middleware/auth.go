package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/toluhikay/fx-exchange/pkg/jwt"
	"github.com/toluhikay/fx-exchange/pkg/utils"
)

type AuthMiddleware struct {
	auth *jwt.Auth
}

func NewMiddleware(auth *jwt.Auth) AuthMiddleware {
	return AuthMiddleware{auth: auth}
}

type ContextUserClaims any

func (mw AuthMiddleware) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := mw.auth.GetTokenFromHeaderAndVerify(w, r)
		if err != nil {
			fmt.Println(err)
			utils.ErrorJSON(w, err, http.StatusUnauthorized)
			return
		}

		var ctxClaimsKey ContextUserClaims = "user_claims"

		ctx := context.WithValue(r.Context(), ctxClaimsKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
