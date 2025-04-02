package authentication

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("tempkey")

type Claims struct {
    Username string `json:"username"`
    jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT for a username
func GenerateToken(username string) (string, error) {
    claims := &Claims{
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(730 * time.Hour)),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtKey)
}

// ParseToken validates and extracts claims from a JWT
func ParseToken(tokenStr string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })
    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, jwt.ErrTokenSignatureInvalid
}

