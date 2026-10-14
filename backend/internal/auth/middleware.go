package auth

import (
	"net/http"
	"strings"

	"github.com/AiEmStylix/ecommerce/backend/internal/appctx"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	jwtService JWTService
}

func NewAuthMiddleware(jwtSvc JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtSvc,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Thiếu token xác thực", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Token không hợp lệ", http.StatusBadRequest)
			return
		}

		tokenString := parts[1]

		token, err := m.jwtService.ValidateToken(tokenString)

		if err != nil || !token.Valid {
			http.Error(w, "Token không hợp lệ hoặc đã hết hạn", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			http.Error(w, "Không thể đọc dữ liệu từ token", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, "Dữ liệu UserID trong token không hợp lệ", http.StatusUnauthorized)
			return
		}

		role, _ := claims["role"].(string)

		// Set userid và role vào context
		ctx := appctx.WithUserID(r.Context(), userID)
		ctx = appctx.WithUserRole(ctx, role)

		reqWithCtx := r.WithContext(ctx)
		next.ServeHTTP(w, reqWithCtx)
	}
}

// RequireAdmin wraps a handler with auth + admin role enforcement.
// It reuses RequireAuth (which sets userID in context) then fetches the user
// from the DB and rejects non-admins with 403 Forbidden.
func (m *AuthMiddleware) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return m.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		role, ok := appctx.GetUserRole(r.Context())
		if !ok || role != "admin" {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
