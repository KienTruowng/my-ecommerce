package appctx

import "context"

type contextKey string

const (
	userIdKey   contextKey = "userID"
	userRoleKey contextKey = "userRole"
)

// Setter: Gán userID vào context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIdKey, userID) // Dùng key nội bộ
}

// Getter: Lấy userID từ context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIdKey).(string) // Dùng key nội bộ
	return userID, ok
}

// Setter: Gán userRole vào context
func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, userRoleKey, role)
}

// Getter: Lấy userRole từ context
func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(userRoleKey).(string)
	return role, ok
}
