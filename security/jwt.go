package security

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtIDBytes — длина случайного идентификатора токена (`jti`), байт: 128 бит,
// совпадение двух токенов исключено практически.
const jwtIDBytes = 16

/*
GenerateJWT creates a new Json Web Token, based on provided data.

value - any value to store in JWT payload;

secretKey - is a secret, on base of which will be checked, if JWT can be trusted;

ttl - time JWT is appropriate and after which will be expired;

algorithm - JWT hashing algorithm like HS256 and so on.
*/
func GenerateJWT(
	value any,
	secretKey string,
	ttl time.Duration,
	algorithm string,
	opts ...jwt.TokenOption,
) (string, error) {
	token := jwt.New(jwt.GetSigningMethod(algorithm), opts...)

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", &JWTClaimsError{}
	}

	// jti делает токен уникальным: без него два токена с одним значением,
	// выданные в одну секунду, совпадали побайтно — `exp` считается в
	// секундах, — и хранилище с уникальным индексом по значению отвергало
	// второй (двойной вход одного пользователя).
	id := make([]byte, jwtIDBytes)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}

	claims["value"] = value
	claims["exp"] = time.Now().UTC().Add(ttl).Unix()
	claims["jti"] = hex.EncodeToString(id)

	return token.SignedString([]byte(secretKey))
}

/*
ParseJWT decodes a Json Web Token payload.

tokenString - a JWT, which will be parsed;

secretKey - is a secret, on base of which will be checked, if JWT can be trusted;.
*/
func ParseJWT(tokenString, secretKey string, opts ...jwt.ParserOption) (any, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		jwt.MapClaims{},
		func(_ *jwt.Token) (any, error) {
			return []byte(secretKey), nil
		},
		opts...,
	)

	if err != nil || !token.Valid {
		return nil, &InvalidJWTError{}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, &JWTClaimsError{}
	}

	return claims["value"], nil
}
