package routers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourname/state-monitor/orm"
)

const apiKeyPrefix = "smKey-"
const apiKeyRandomBytes = 16 // 32 个十六进制字符

func generateApiKey() string {
	b := make([]byte, apiKeyRandomBytes)
	_, _ = rand.Read(b)
	return apiKeyPrefix + hex.EncodeToString(b)
}

func maskApiKey(key string) string {
	if len(key) <= 8 {
		return "••••••••"
	}
	return key[:8] + "••••••••"
}

func RegisterApiKeyRoutes(r *gin.Engine) {
	g := r.Group("/api-keys")
	db := orm.DB()

	// 获取列表（apiKey 返回脱敏）
	g.GET("", func(ctx *gin.Context) {
		var rows []orm.SMApiKey
		if err := db.Order("created_at DESC").Find(&rows).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			list = append(list, gin.H{
				"id":        strconv.FormatInt(r.ID, 10),
				"name":      r.Name,
				"apiKey":    maskApiKey(r.ApiKey),
				"createdAt": r.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
			})
		}
		ctx.JSON(http.StatusOK, gin.H{"list": list})
	})

	// 创建
	g.POST("", func(ctx *gin.Context) {
		var req struct {
			Name string `json:"name" binding:"required"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
			return
		}
		name := req.Name
		if name == "" {
			name = "未命名"
		}
		key := generateApiKey()
		row := orm.SMApiKey{Name: name, ApiKey: key}
		if err := db.Create(&row).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"id":        strconv.FormatInt(row.ID, 10),
			"name":      row.Name,
			"apiKey":    row.ApiKey,
			"createdAt": row.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		})
	})

	// 刷新（重新生成 Key），仅返回时带完整 apiKey
	g.PUT("/:id/refresh", func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的 id"})
			return
		}
		var row orm.SMApiKey
		if err := db.First(&row, id).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "API Key 不存在"})
			return
		}
		newKey := generateApiKey()
		if err := db.Model(&row).Update("api_key", newKey).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		db.First(&row, id)
		ctx.JSON(http.StatusOK, gin.H{
			"apiKey":    newKey,
			"updatedAt": row.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		})
	})

	// 查看完整 Key（用于复制，仅此接口返回明文）
	g.GET("/:id/reveal", func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的 id"})
			return
		}
		var row orm.SMApiKey
		if err := db.First(&row, id).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "API Key 不存在"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"apiKey": row.ApiKey})
	})

	// 删除
	g.DELETE("/:id", func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的 id"})
			return
		}
		result := db.Delete(&orm.SMApiKey{}, id)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		if result.RowsAffected == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "API Key 不存在"})
			return
		}
		ctx.Status(http.StatusNoContent)
	})
}
