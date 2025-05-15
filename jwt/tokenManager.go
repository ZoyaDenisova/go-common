package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenPair struct {
	AccessToken    string
	RefreshToken   string
	AccessExpires  time.Time
	RefreshExpires time.Time
}

type TokenManager interface {
	Generate(userID int64, role string) (TokenPair, error)
	ValidateAccess(token string) (userID int64, role string, err error)
	ValidateRefresh(token string) (userID int64, role string, err error)
}
type JWTManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTManager(secret string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

type customClaims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (j *JWTManager) Generate(userID int64, role string) (TokenPair, error) {
	now := time.Now().UTC()

	// Access
	accessExp := now.Add(j.accessTTL)
	accessClaims := customClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExp),
		},
	}
	accessToken, err := j.createToken(accessClaims)
	if err != nil {
		return TokenPair{}, err
	}

	// Refresh
	refreshExp := now.Add(j.refreshTTL)
	refreshClaims := customClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExp),
		},
	}
	refreshToken, err := j.createToken(refreshClaims)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		AccessExpires:  accessExp,
		RefreshExpires: refreshExp,
	}, nil
}

func (j *JWTManager) createToken(claims customClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTManager) ValidateAccess(tokenStr string) (int64, string, error) {
	return j.validate(tokenStr)
}

func (j *JWTManager) ValidateRefresh(tokenStr string) (int64, string, error) {
	return j.validate(tokenStr)
}

func (j *JWTManager) validate(tokenStr string) (int64, string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &customClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil {
		return 0, "", err
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return 0, "", errors.New("invalid token")
	}

	return claims.UserID, claims.Role, nil
}
