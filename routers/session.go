package routers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourname/state-monitor/orm"
)

func RegisterSessionRoutes(r *gin.Engine) {
	g := r.Group("/sessions")
	db := orm.DB()

	// 获取会话历史（必须写在 GET /:id 之前，否则 /history 会被匹配成 id=history）
	g.GET("/history", func(ctx *gin.Context) {
		sessionIdStr := ctx.Query("sessionId")
		page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		q := db.Model(&orm.SessionDetail{})
		if sessionIdStr != "" {
			sessionId, err := strconv.ParseInt(sessionIdStr, 10, 64)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的 sessionId"})
				return
			}
			q = q.Where("session_id = ?", sessionId)
		}
		var total int64
		q.Count(&total)
		var rows []orm.SessionDetail
		offset := (page - 1) * pageSize
		if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		list := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			list = append(list, gin.H{
				"id":        strconv.FormatInt(r.ID, 10),
				"sessionId": strconv.FormatInt(r.SessionID, 10),
				"event":     r.Event,
				"fromState": r.FromState,
				"toState":   r.ToState,
				"createdAt": r.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
			})
		}
		ctx.JSON(http.StatusOK, gin.H{"list": list, "total": total})
	})

	// 获取会话列表
	g.GET("", func(ctx *gin.Context) {
		page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
		stateMachineId := ctx.Query("stateMachineId")
		status := ctx.Query("status")
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		q := db.Model(&orm.SessionInfo{})
		if stateMachineId != "" {
			smId, err := strconv.ParseInt(stateMachineId, 10, 64)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的 stateMachineId"})
				return
			}
			q = q.Where("sm_id = ?", smId)
		}
		if status != "" {
			q = q.Where("status = ?", status)
		}
		var total int64
		q.Count(&total)
		var rows []orm.SessionInfo
		offset := (page - 1) * pageSize
		if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		list := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			list = append(list, gin.H{
				"id":             strconv.FormatInt(r.ID, 10),
				"sessionId":      strconv.FormatInt(r.ID, 10),
				"stateMachineId": strconv.FormatInt(r.SMID, 10),
				"status":         r.Status,
				"createdAt":      r.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
			})
		}
		ctx.JSON(http.StatusOK, gin.H{"list": list, "total": total})
	})

	// 获取单个会话详情
	g.GET("/:id", func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的 id"})
			return
		}
		var s orm.SessionInfo
		if err := db.First(&s, id).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"id":             strconv.FormatInt(s.ID, 10),
			"sessionId":      strconv.FormatInt(s.ID, 10),
			"stateMachineId": strconv.FormatInt(s.SMID, 10),
			"status":         s.Status,
			"createdAt":      s.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
			"updatedAt":      s.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		})
	})
}
