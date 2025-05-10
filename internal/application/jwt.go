	package application

	import (
		"os"
		"strconv"
		"time"

		"github.com/golang-jwt/jwt/v5"
	)

	var (
		jwtSecret           = []byte(os.Getenv("JWT_SECRET"))
		jwtExpirationMinute int
	)

	func init() {
		min, err := strconv.Atoi(os.Getenv("JWT_EXPIRATION_MINUTES"))
		if err != nil {
			jwtExpirationMinute = 60
		} else {
			jwtExpirationMinute = min
		}
	}

	type Claims struct {
		UserID string `json:"user_id"`
		jwt.RegisteredClaims
	}

	// GenerateJWT создаёт JWT с HS256 и RegisteredClaims
	func GenerateJWT(userID string) (string, error) {
		claims := Claims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtExpirationMinute) * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString(jwtSecret)
	}

		
	// ParseJWT проверяет и парсит токен, возвращает Claims
	func ParseJWT(tokenStr string) (*Claims, error) {
		var claims Claims
		// WithValidMethods — явно указываем ожидаемый алгоритм
		parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
		_, err := parser.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil {
			return nil, err
		}
		// В jwt/v5 валидность проверяется в ParseWithClaims, если нужно — дополнительно проверяйте Claims.VerifyExpiresAt
		return &claims, nil
	}