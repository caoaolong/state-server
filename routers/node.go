package routers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/caoaolong/state-server/orm"
)

// RunNodeRequest 运行节点请求：Node + 可选 sessionId（默认 0，设计页固定为 0）
type RunNodeRequest struct {
	Node      RunNodePayload `json:"node" binding:"required"`
	SessionID int64          `json:"sessionId"` // 逻辑会话 id，0 表示设计页会话
}

// RunNodePayload 节点结构（与前端/流程中的节点一致）
type RunNodePayload struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Position *struct{ X, Y float64 } `json:"position,omitempty"`
	Data     *RunNodeData     `json:"data,omitempty"`
}

// RunNodeData 节点 data 中与请求相关的字段
type RunNodeData struct {
	RequestPath   string `json:"requestPath"`
	RequestMethod string `json:"requestMethod"`
	RequestData   string `json:"requestData"`
}

// RunNodeResponse 运行节点响应
type RunNodeResponse struct {
	OK         bool   `json:"ok"`
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
	Error      string `json:"error,omitempty"`
}

func RegisterNodeRoutes(r *gin.Engine) {
	r.POST("/nodes/run", runNode)
	// 节点 CRUD 挂到 /flow 下，与前端约定一致
	g := r.Group("/flow")
	g.PUT("/:id/nodes/:nodeId", putFlowNode)
	g.POST("/:id/nodes", postFlowNode)
}

// putFlowNode 更新单个节点 PUT /flow/:id/nodes/:nodeId（编辑窗口保存时调用）
func putFlowNode(c *gin.Context) {
	idStr := c.Param("id")
	nodeID := c.Param("nodeId")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "节点 id 不能为空"})
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态机 id"})
		return
	}
	db := orm.DB()
	if err := db.First(&orm.SMFlow{}, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "状态机不存在"})
		return
	}
	var req nodePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误: " + err.Error()})
		return
	}
	if req.ID != nodeID {
		req.ID = nodeID
	}
	storeBytes, _ := json.Marshal(map[string]any{"position": req.Position, "data": req.Data})
	if len(storeBytes) <= 2 {
		storeBytes = []byte("{}")
	}
	reqPath, reqMethod, reqData := getRequestFieldsFromData(req.Data)
	label := getLabelFromData(req.Data)
	result := db.Model(&orm.SMNode{}).Where("sm_id = ? AND node_id = ?", id, nodeID).Updates(map[string]interface{}{
		"type":           req.Type,
		"label":          label,
		"data":           string(storeBytes),
		"request_path":   reqPath,
		"request_method": reqMethod,
		"request_data":   reqData,
	})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// postFlowNode 创建单个节点 POST /flow/:id/nodes（创建节点时保存到服务端）
func postFlowNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态机 id"})
		return
	}
	db := orm.DB()
	if err := db.First(&orm.SMFlow{}, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "状态机不存在"})
		return
	}
	var req nodePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误: " + err.Error()})
		return
	}
	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "节点 id 不能为空"})
		return
	}
	var existing orm.SMNode
	err = db.Where("sm_id = ? AND node_id = ?", id, req.ID).First(&existing).Error
	if err == nil {
		storeBytes, _ := json.Marshal(map[string]any{"position": req.Position, "data": req.Data})
		if len(storeBytes) <= 2 {
			storeBytes = []byte("{}")
		}
		reqPath, reqMethod, reqData := getRequestFieldsFromData(req.Data)
		label := getLabelFromData(req.Data)
		if err := db.Model(&existing).Updates(map[string]interface{}{
			"type":           req.Type,
			"label":          label,
			"data":           string(storeBytes),
			"request_path":   reqPath,
			"request_method": reqMethod,
			"request_data":   reqData,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	storeBytes, _ := json.Marshal(map[string]any{"position": req.Position, "data": req.Data})
	if len(storeBytes) <= 2 {
		storeBytes = []byte("{}")
	}
	reqPath, reqMethod, reqData := getRequestFieldsFromData(req.Data)
	node := orm.SMNode{
		SMID:          id,
		NodeID:        req.ID,
		Type:          req.Type,
		Label:         getLabelFromData(req.Data),
		Data:          string(storeBytes),
		RequestPath:   reqPath,
		RequestMethod: reqMethod,
		RequestData:   reqData,
	}
	if node.Type == "" {
		node.Type = "default"
	}
	if err := db.Create(&node).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func runNode(c *gin.Context) {
	var req RunNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	if req.Node.Data == nil {
		req.Node.Data = &RunNodeData{}
	}
	nodeID := strings.TrimSpace(req.Node.ID)
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, RunNodeResponse{OK: false, Error: "节点 id 不能为空"})
		return
	}
	db := orm.DB()
	var smNode orm.SMNode
	if err := db.Where("node_id = ?", nodeID).First(&smNode).Error; err != nil {
		c.JSON(http.StatusNotFound, RunNodeResponse{OK: false, Error: "节点不存在或未保存到流程"})
		return
	}
	var flow orm.SMFlow
	if err := db.First(&flow, smNode.SMID).Error; err != nil {
		c.JSON(http.StatusNotFound, RunNodeResponse{OK: false, Error: "所属状态机不存在"})
		return
	}
	baseURL := strings.TrimSuffix(strings.TrimSpace(flow.BaseURL), "/")
	path := strings.TrimSpace(req.Node.Data.RequestPath)
	if path != "" && path[0] != '/' {
		path = "/" + path
	}
	url := baseURL + path
	if baseURL == "" {
		c.JSON(http.StatusBadRequest, RunNodeResponse{OK: false, Error: "请先配置状态机的 Base URL"})
		return
	}

	method := strings.TrimSpace(strings.ToUpper(req.Node.Data.RequestMethod))
	if method == "" {
		method = "GET"
	}

	var body io.Reader
	if method != "GET" && req.Node.Data.RequestData != "" {
		body = bytes.NewBufferString(req.Node.Data.RequestData)
	}
	httpReq, err := http.NewRequest(method, url, body)
	if err != nil {
		c.JSON(http.StatusBadRequest, RunNodeResponse{OK: false, Error: "构建请求失败: " + err.Error()})
		return
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusOK, RunNodeResponse{OK: false, Error: "请求失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	respBodyStr := string(respBody)
	ok := resp.StatusCode >= 200 && resp.StatusCode < 300

	// 记录会话历史（事务）：会话不存在则创建，按 flowId + nodeId + sessionId 确定唯一，存在则更新
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	sessionID := req.SessionID
	var session orm.SessionInfo
	err = tx.Where("sm_id = ? AND logical_session_id = ?", smNode.SMID, sessionID).First(&session).Error
	if err != nil {
		session = orm.SessionInfo{SMID: smNode.SMID, LogicalSessionID: sessionID, State: "", Status: "running"}
		if err = tx.Create(&session).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, RunNodeResponse{OK: false, Error: "创建会话失败: " + err.Error()})
			return
		}
	}
	reqDataStr := ""
	if req.Node.Data != nil && req.Node.Data.RequestData != "" {
		reqDataStr = req.Node.Data.RequestData
	}
	detail := orm.SessionDetail{
		SessionID:    session.ID,
		NodeID:       nodeID,
		SMID:         smNode.SMID,
		Event:        "run_node",
		FromState:    "",
		ToState:      nodeID,
		Path:         path,
		RequestData:  reqDataStr,
		ResponseData: respBodyStr,
	}
	var existing orm.SessionDetail
	errDetail := tx.Where("session_id = ? AND node_id = ?", session.ID, nodeID).First(&existing).Error
	if errDetail != nil {
		if err = tx.Create(&detail).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, RunNodeResponse{OK: false, Error: "记录会话历史失败: " + err.Error()})
			return
		}
	} else {
		if err = tx.Model(&existing).Updates(map[string]interface{}{
			"path":          detail.Path,
			"request_data":  detail.RequestData,
			"response_data": detail.ResponseData,
			"to_state":      detail.ToState,
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, RunNodeResponse{OK: false, Error: "更新会话历史失败: " + err.Error()})
			return
		}
	}
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, RunNodeResponse{OK: false, Error: "提交事务失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, RunNodeResponse{
		OK:         ok,
		StatusCode: resp.StatusCode,
		Body:       respBodyStr,
	})
}
