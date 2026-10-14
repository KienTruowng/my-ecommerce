package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrEmptySecretKey   = errors.New("Secret Key Cannot Be Empty")
	ErrInvalidAlgorithm = errors.New("lỗi: thuật toán mã hóa không hợp lệ")
)

type JWTService interface {
	GenerateAccessToken(userID string, role string) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
}

type jwtCustomClaim struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type jwtService struct {
	secretKey []byte
	issuer    string
}

func NewJWTService(secretKey string) (JWTService, error) {
	if secretKey == "" {
		return nil, ErrEmptySecretKey
	}
	return &jwtService{
		secretKey: []byte(secretKey),
		issuer:    "ecommerce_backend",
	}, nil
}

func (j *jwtService) GenerateAccessToken(userID string, role string) (string, error) {
	claims := &jwtCustomClaim{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(j.secretKey)

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (j *jwtService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Kiểm tra method ký có phải HMAC không (bảo mật)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidAlgorithm
		}
		return j.secretKey, nil
	})
}
