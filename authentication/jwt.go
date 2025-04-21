package authentication

import (
    "os"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

// Get JWT key from environment variable or use default (for development only)
var jwtKey = getJWTKey()

func getJWTKey() []byte {
    if key := os.Getenv("JWT_SECRET_KEY"); key != "" {
        return []byte(key)
    }
    // This default key should only be used in development
    return []byte("tempkey") 
}

type Claims struct {
    Username string `json:"username"`
    UID int `json:"uid"`
    jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT for a username
func GenerateToken(username string, uId int) (string, error) {
    claims := &Claims{
        Username: username,
        UID: uId,
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