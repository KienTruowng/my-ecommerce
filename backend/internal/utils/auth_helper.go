package utils

import (
	"net/http"
	"strconv"

	"github.com/AiEmStylix/ecommerce/backend/internal/appctx"
)

func ValidateOwnership(w http.ResponseWriter, r *http.Request) (uint, bool) {
	callerID, ok := appctx.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Chưa xác thực", http.StatusUnauthorized)
		return 0, false
	}

	// Lấy ID từ Path
	idStr := r.PathValue("id")
	if idStr == "" {
		// Nếu API dạng /me thì mặc định lấy luôn callerID
		idStr = callerID
	}

	if callerID != idStr {
		http.Error(w, "Bạn không có quyền truy cập", http.StatusForbidden)
		return 0, false
	}

	idUint64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
		return 0, false
	}

	return uint(idUint64), true
}
