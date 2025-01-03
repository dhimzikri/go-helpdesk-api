package utils

import (
	"fmt"
	"golang-sqlserver-app/models"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var mySigningKey = []byte("mysecretkey")
var (
	loggedOutTokens = make(map[string]struct{})
	mu              sync.Mutex
)

type MyCustomClaims struct {
	User_name string `json:"user_name"`
	RealName  string `json:"real_name"`
	jwt.RegisteredClaims
}

func CreateToken(user *models.User) (string, error) {
	claims := MyCustomClaims{
		user.Username,
		user.RealName,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)

	// issuedAt := claims.IssuedAt.Time.Format("2006-01-02 15:04:05")
	// expiresAt := claims.ExpiresAt.Time.Format("2006-01-02 15:04:05")

	// // Simpan token ke dalam tabel baru
	// tokenRecord := models.TokenRecord{
	// 	Username:    user.Name, // Gantilah dengan nama kolom yang sesuai di model User
	// 	Password:    user.Password,
	// 	Token:       ss,
	// 	ExpiredDate: expiresAt,
	// 	CreatedDate: issuedAt,
	// }

	// if err := config.DB.Create(&tokenRecord).Error; err != nil {
	// 	return "", err
	// }

	return ss, err
}

func ValidateToken(tokenString string) (*MyCustomClaims, error) {
	mu.Lock()
	defer mu.Unlock()
	if _, found := loggedOutTokens[tokenString]; found {
		return nil, fmt.Errorf("token has been logged out")
	}

	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})

	if err != nil {
		validationErr, isValidationErr := err.(*jwt.ValidationError)
		if isValidationErr {
			if validationErr.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, fmt.Errorf("token has expired, please log in again")
			}
		}

		return nil, fmt.Errorf("unauthorized")
	}

	claims, ok := token.Claims.(*MyCustomClaims)

	if !ok || !token.Valid {
		return nil, fmt.Errorf("unauthorized")
	}

	return claims, nil
}

func AddLoggedOutToken(token string) {
	mu.Lock()
	defer mu.Unlock()
	loggedOutTokens[token] = struct{}{}
}
