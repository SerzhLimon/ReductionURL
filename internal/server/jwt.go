package server

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/SerzhLimon/ReductionURL/internal/model"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

const (
	secretKey = "super_secret_key_for_shortener"
)

var userIDCounter int

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func createJWT(res http.ResponseWriter) error {
	userIDCounter++
	userID := userIDCounter

	expirationTime := time.Now().Add(24 * time.Hour) // Токен на 24 часа

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(secretKey)) // Подписание токена секретным ключом
	if err != nil {
		return fmt.Errorf("failed to create token %w", err)
	}

	http.SetCookie(res, &http.Cookie{
		Name:    "jwt_token",
		Value:   tokenStr,
		Expires: expirationTime,
		// Path:     "/",
	})
	return nil
}

func validateJWT(res http.ResponseWriter, req *http.Request) error {
	cookie, err := req.Cookie("jwt_token")
	if err != nil {
		logrus.Error("here0")
		return createJWT(res)
	}

	tokenStr := cookie.Value
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		},
	)

	id := claims.UserID
	if id < 1 {
		return model.ErrEmptyUserID
	}
	if err != nil {
		logrus.Error(err, "here1")
		return fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		logrus.Error("here2")
		return fmt.Errorf("invalid token")
	}

	return nil
}

func cookies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := validateJWT(w, r); err != nil {
			if errors.Is(err, model.ErrEmptyUserID) {
				http.Error(w, "invalid JWT", http.StatusUnauthorized)
				return
			}
			http.Error(w, "invalid JWT", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}