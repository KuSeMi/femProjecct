package middleware

import (
	"context"
	"femProject/internal/store"
	"femProject/internal/tokens"
	"femProject/internal/utils"
	"net/http"
	"strings"
)

type UserMiddleware struct {
	UserStore store.UserStore
}

type contextKey string

const UserContextKey = contextKey("userKey")

func SetUser(r *http.Request, user *store.User) *http.Request {
	ctx := context.WithValue(r.Context(), UserContextKey, user)
	return r.WithContext(ctx)
}

func GetUser(r *http.Request) *store.User {
	user, ok := r.Context().Value(UserContextKey).(*store.User)
	if !ok {
		panic("user not found in request context")
	}

	return user
}

func (um *UserMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			r = SetUser(r, store.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.SplitN(authHeader, " ", 2)
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"Error": "Invalid authorization header"})
			return
		}

		token := headerParts[1]
		user, err := um.UserStore.GetUserToken(tokens.ScopeAuth, token)
		if err != nil {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"Error": "Invalid token"})
			return
		}

		if user == nil {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"Error": "Token expired or invalid"})
			return
		}

		r = SetUser(r, user)
		next.ServeHTTP(w, r)
		return
	})
}

func (um *UserMiddleware) RequireUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)
		if user.IsAnonymous() {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"Error": "You must be logged in"})
			return
		}

		next.ServeHTTP(w, r)
	})
}
