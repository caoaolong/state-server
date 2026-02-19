package routers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/caoaolong/state-server/orm"
	"gorm.io/gorm"
)

// 创建状态机请求体
type createStateMachineReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// 保存流程请求体：与前端 FlowData 一致
type flowSaveReq struct {
	Nodes []nodePayload `json:"nodes"`
	Edges []edgePayload `json:"edges"`
}

type nodePayload struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Position positionPayload `json:"position"`
	Data     json.RawMessage `json:"data"`
}

type positionPayload struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type edgePayload struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// 列表/详情返回用（BaseURL 对应前端 baseUrl）
type stateMachineListItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	BaseURL     string `json:"baseUrl"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func RegisterStateMachineRoutes(r *gin.Engine) {
	g := r.Group("/flow")
	db := orm.DB()

	// 获取状态机列表
	g.GET("", func(ctx *gin.Context) {
		page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
		keyword := ctx.Query("keyword")
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		var list []orm.SMFlow
		q := db.Model(&orm.SMFlow{})
		if keyword != "" {
			q = q.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}
		var total int64
		q.Count(&total)
		offset := (page - 1) * pageSize
		if err := q.Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&list).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		items := make([]stateMachineListItem, 0, len(list))
		for _, row := range list {
			items = append(items, stateMachineListItem{
				ID:          strconv.FormatInt(row.ID, 10),
				Name:        row.Name,
				Description: row.Description,
				BaseURL:     row.BaseURL,
				CreatedAt:   row.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
				UpdatedAt:   row.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
			})
		}
		ctx.JSON(http.StatusOK, gin.H{"list": items, "total": total})
	})

	// 创建状态机（事务）
	g.POST("", func(ctx *gin.Context) {
		var req createStateMachineReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
			return
		}
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		row := orm.SMFlow{Name: req.Name, Description: req.Description}
		if err := tx.Create(&row).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败: " + err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"id":          strconv.FormatInt(row.ID, 10),
			"name":        row.Name,
			"description": row.Description,
			"createdAt":   row.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
			"updatedAt":   row.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		})
	})

	// 获取状态机流程数据（必须写在 GET /:id 之前，否则 /123/flow 会被 /:id 匹配成 id=123/flow）
	g.GET("/:id/flow", func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的 id"})
			return
		}
		var flow orm.SMFlow
		if err := db.First(&flow, id).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "状态机不存在"})
			return
		}
		nodes, edges := loadFlowData(db, id)
		ctx.JSON(http.StatusOK, gin.H{"nodes": nodes, "edges": edges})
	})

	// 获取单个状态机详情（含流程数据）
	g.GET("/:id", func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的 id"})
			return
		}
		var flow orm.SMFlow
		if err := db.First(&flow, id).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "状态机不存在"})
			return
		}
		nodes, edges := loadFlowData(db, id)
		ctx.JSON(http.StatusOK, gin.H{
			"id":          strconv.FormatInt(flow.ID, 10),
			"name":        flow.Name,
			"description": flow.Description,
			"baseUrl":     flow.BaseURL,
			"identifier":  flow.Identifier,
			"createdAt":   flow.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
			"updatedAt":   flow.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
			"flowData":    gin.H{"nodes": nodes, "edges": edges},
		})
	})

	// 保存流程 PUT /flow/:id/flow
	g.PUT("/:id/flow", func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的 id"})
			return
		}
		var flow orm.SMFlow
		if err := db.First(&flow, id).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "状态机不存在"})
			return
		}
		var req flowSaveReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误: " + err.Error()})
			return
		}
		if req.Nodes == nil {
			req.Nodes = []nodePayload{}
		}
		if req.Edges == nil {
			req.Edges = []edgePayload{}
		}

		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		// 物理删除该状态机下原有节点与边，再重新插入（避免软删导致表内记录只增不减）
		if err := tx.Unscoped().Where("sm_id = ?", id).Delete(&orm.SMNode{}).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := tx.Unscoped().Where("sm_id = ?", id).Delete(&orm.SMEdge{}).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// 写入节点
		for _, n := range req.Nodes {
			// 存整段 JSON：{ position: {...}, data: {...} }，与前端格式一致
			storeBytes, _ := json.Marshal(map[string]any{"position": n.Position, "data": n.Data})
			if len(storeBytes) <= 2 {
				storeBytes = []byte("{}")
			}
			reqPath, reqMethod, reqData := getRequestFieldsFromData(n.Data)
			node := orm.SMNode{
				SMID:          id,
				NodeID:        n.ID,
				Type:          n.Type,
				Label:         getLabelFromData(n.Data),
				Data:          string(storeBytes),
				RequestPath:   reqPath,
				RequestMethod: reqMethod,
				RequestData:   reqData,
			}
			if node.Type == "" {
				node.Type = "default"
			}
			if err := tx.Create(&node).Error; err != nil {
				tx.Rollback()
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		// 写入边
		for _, e := range req.Edges {
			edge := orm.SMEdge{
				SMID:       id,
				EdgeID:     e.ID,
				FromNodeID: e.Source,
				ToNodeID:   e.Target,
			}
			if err := tx.Create(&edge).Error; err != nil {
				tx.Rollback()
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		// 更新状态机 updated_at
		now := time.Now()
		if err := tx.Model(&orm.SMFlow{}).Where("id = ?", id).Update("updated_at", now).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := tx.Commit().Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// 重新查一次 updated_at
		var updated orm.SMFlow
		db.First(&updated, id)
		ctx.JSON(http.StatusOK, gin.H{
			"ok":        true,
			"updatedAt": updated.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		})
	})

	// 更新状态机（名称、描述等）
	g.PUT("/:id", func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的 id"})
			return
		}
		var flow orm.SMFlow
		if err := db.First(&flow, id).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "状态机不存在"})
			return
		}
		var req struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
			BaseURL     *string `json:"baseUrl"`
			Identifier  *string `json:"identifier"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
			return
		}
		updates := make(map[string]interface{})
		if req.Name != nil {
			updates["name"] = *req.Name
		}
		if req.Description != nil {
			updates["description"] = *req.Description
		}
		if req.BaseURL != nil {
			updates["base_url"] = *req.BaseURL
		}
		if req.Identifier != nil {
			updates["identifier"] = *req.Identifier
		}
		if len(updates) > 0 {
			if err := db.Model(&flow).Updates(updates).Error; err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		db.First(&flow, id)
		ctx.JSON(http.StatusOK, gin.H{
			"id":          strconv.FormatInt(flow.ID, 10),
			"name":        flow.Name,
			"description": flow.Description,
			"baseUrl":     flow.BaseURL,
			"updatedAt":   flow.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		})
	})

	// 删除状态机（事务）
	g.DELETE("/:id", func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的 id"})
			return
		}
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		result := tx.Delete(&orm.SMFlow{}, id)
		if result.Error != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			ctx.JSON(http.StatusNotFound, gin.H{"error": "状态机不存在"})
			return
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败: " + err.Error()})
			return
		}
		ctx.Status(http.StatusNoContent)
	})
}

// loadFlowData 从 DB 加载节点与边，返回前端格式的 nodes/edges
func loadFlowData(db *gorm.DB, smID int64) (nodes []map[string]any, edges []map[string]any) {
	var nodeRows []orm.SMNode
	db.Where("sm_id = ?", smID).Find(&nodeRows)
	nodes = make([]map[string]any, 0, len(nodeRows))
	for _, r := range nodeRows {
		var dataMap map[string]any
		if r.Data != "" {
			_ = json.Unmarshal([]byte(r.Data), &dataMap)
		}
		if dataMap == nil {
			dataMap = make(map[string]any)
		}
		position, _ := dataMap["position"].(map[string]any)
		data, _ := dataMap["data"]
		if data == nil {
			data = map[string]any{"label": r.Label}
		}
		dataObj, _ := data.(map[string]any)
		if dataObj == nil {
			dataObj = map[string]any{"label": r.Label}
		}
		if r.RequestPath != "" {
			dataObj["requestPath"] = r.RequestPath
		}
		if r.RequestMethod != "" {
			dataObj["requestMethod"] = r.RequestMethod
		}
		if r.RequestData != "" {
			dataObj["requestData"] = r.RequestData
		}
		nodes = append(nodes, map[string]any{
			"id":       r.NodeID,
			"type":     r.Type,
			"position": position,
			"data":     dataObj,
		})
	}
	var edgeRows []orm.SMEdge
	db.Where("sm_id = ?", smID).Find(&edgeRows)
	edges = make([]map[string]any, 0, len(edgeRows))
	for _, r := range edgeRows {
		edges = append(edges, map[string]any{
			"id":     r.EdgeID,
			"source": r.FromNodeID,
			"target": r.ToNodeID,
		})
	}
	return nodes, edges
}

// getLabelFromData 从 data JSON 中取出 label
func getLabelFromData(data json.RawMessage) string {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	if l, ok := m["label"].(string); ok {
		return l
	}
	return ""
}

// getRequestFieldsFromData 从 data JSON 中取出 requestPath、requestMethod、requestData
func getRequestFieldsFromData(data json.RawMessage) (path, method, dataStr string) {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return "", "", ""
	}
	if p, ok := m["requestPath"].(string); ok {
		path = p
	}
	if mth, ok := m["requestMethod"].(string); ok {
		method = mth
	}
	if d, ok := m["requestData"].(string); ok {
		dataStr = d
	}
	return path, method, dataStr
}
