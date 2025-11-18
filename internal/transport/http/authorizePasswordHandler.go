package http
import (
	"net/http"
	"strconv"
	"file-sharing/internal/share"
	"github.com/gin-gonic/gin"
	"file-sharing/internal/files"
	"errors"
)

type authorizePasswordHandler struct {
	aService share.AuthServices
}

func NewAuthorizePasswordHandler(aService share.AuthServices) *authorizePasswordHandler {
	return &authorizePasswordHandler{aService: aService}
}

// POST /v1/shares/:id/authorize
func (h *authorizePasswordHandler) HandleAuthorizePassword(c *gin.Context) {
	type authorizeRequest struct {
		Password string `json:"password" binding:"required"`
	}
	// Lấy share ID và plain password từ URL
	shareID := c.Param("id")
	id , err := strconv.ParseInt(shareID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Share ID"})
		return
	}
	var req authorizeRequest
	if err := c.ShouldBindJSON(&req);  err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	// Gọi Service để xác thực mật khẩu
	token, err := h.aService.AuthorizeSharePasswordAndIssueToken(c.Request.Context(), id, req.Password)
	if err != nil {
		if errors.Is(err, files.InvalidPasswordError()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	// Trả về token truy cập tạm thời
	c.JSON(http.StatusOK, gin.H{"token": token})

}