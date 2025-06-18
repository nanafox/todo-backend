package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

type GenerateJWTOptions struct {
	AddNewRefreshToken bool
}

type JWTClaims struct {
	jwt.RegisteredClaims

	Sub   uint   `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Exp   int64  `json:"exp"`
}

var secretKey = []byte("strong-secret-key")

func GenerateJWT(jwtClaims *JWTClaims, options GenerateJWTOptions) (tokens map[string]string, err error) {
	jwtClaims.Exp = time.Now().Add(15 * time.Minute).Unix() // Set expiration to 15 minutes

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	tokenString, err := accessToken.SignedString(secretKey)
	if err != nil {
		return nil, err
	}

	var refreshTokenString string

	if options.AddNewRefreshToken {
		jwtClaims.Exp = time.Now().Add(24 * time.Hour).Unix() // Set expiration to 24 hours for refresh token

		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
		refreshTokenString, err = refreshToken.SignedString(secretKey)
		if err != nil {
			return nil, err
		}
	}

	return map[string]string{
		"access_token":  tokenString,
		"refresh_token": refreshTokenString,
	}, nil
}

func VerifyJWT(tokenString string) (userId uint, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	sub, ok := claims["sub"].(float64)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	return uint(sub), nil
}

func BearerTokenAuthenticationMiddleware() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: secretKey},
	})
}
