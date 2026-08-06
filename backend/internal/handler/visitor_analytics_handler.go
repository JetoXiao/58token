package handler

import (
	"database/sql"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	clientip "github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type VisitorAnalyticsHandler struct {
	service *service.VisitorAnalyticsService
}

func NewVisitorAnalyticsHandler(visitorService *service.VisitorAnalyticsService) *VisitorAnalyticsHandler {
	return &VisitorAnalyticsHandler{service: visitorService}
}

type trackVisitorRequest struct {
	ChannelCode string `json:"channel_code"`
	VisitorID   string `json:"visitor_id"`
	SessionID   string `json:"session_id"`
	Path        string `json:"path"`
	Referrer    string `json:"referrer"`
	LandingURL  string `json:"landing_url"`
	Language    string `json:"language"`
	Screen      string `json:"screen"`
}

type visitorChannelRequest struct {
	Name            string `json:"name" binding:"required"`
	Code            string `json:"code" binding:"required"`
	DestinationPath string `json:"destination_path"`
	Description     string `json:"description"`
	Active          *bool  `json:"active"`
}

func (h *VisitorAnalyticsHandler) Track(c *gin.Context) {
	var req trackVisitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid visit payload")
		return
	}
	ip := visitorClientIP(c)
	if net.ParseIP(ip) == nil {
		response.BadRequest(c, "Unable to determine client IP")
		return
	}
	countryCode := ""
	if visitorTrustsProxyHeaders(c) {
		countryCode = strings.TrimSpace(c.GetHeader("CF-IPCountry"))
	}
	if len(countryCode) > 8 || strings.EqualFold(countryCode, "XX") {
		countryCode = ""
	}
	var userID int64
	if subject, ok := servermiddleware.GetAuthSubjectFromContext(c); ok {
		userID = subject.UserID
	}
	err := h.service.Track(c.Request.Context(), service.VisitorTrackInput{
		UserID:      userID,
		ChannelCode: req.ChannelCode,
		VisitorID:   req.VisitorID,
		SessionID:   req.SessionID,
		IP:          ip,
		CountryCode: countryCode,
		Path:        req.Path,
		Referrer:    req.Referrer,
		LandingURL:  req.LandingURL,
		UserAgent:   c.GetHeader("User-Agent"),
		Language:    req.Language,
		Screen:      req.Screen,
	})
	if err != nil {
		response.InternalError(c, "Failed to record visit")
		return
	}
	response.Accepted(c, gin.H{"recorded": true})
}

// visitorClientIP trusts the reverse-proxy headers only when the TCP peer is
// local/private (Caddy or the Docker bridge). Direct public requests cannot
// spoof X-Real-IP to pollute administrator analytics.
func visitorClientIP(c *gin.Context) string {
	fallback := clientip.GetTrustedClientIP(c)
	if !visitorTrustsProxyHeaders(c) {
		if c == nil || c.Request == nil {
			return fallback
		}
		peerHost, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			peerHost = c.Request.RemoteAddr
		}
		if peerIP := net.ParseIP(strings.TrimSpace(peerHost)); peerIP != nil {
			return peerIP.String()
		}
		return fallback
	}
	for _, header := range []string{"X-Real-IP", "CF-Connecting-IP"} {
		candidate := strings.TrimSpace(c.GetHeader(header))
		if parsed := net.ParseIP(candidate); parsed != nil {
			return parsed.String()
		}
	}
	return fallback
}

func visitorTrustsProxyHeaders(c *gin.Context) bool {
	if c == nil || c.Request == nil {
		return false
	}
	peerHost, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		peerHost = c.Request.RemoteAddr
	}
	peerIP := net.ParseIP(strings.TrimSpace(peerHost))
	if peerIP == nil {
		return false
	}
	return peerIP.IsLoopback() || peerIP.IsPrivate()
}

func (h *VisitorAnalyticsHandler) Overview(c *gin.Context) {
	start, end, _, err := visitorDateRange(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.service.GetOverview(c.Request.Context(), start, end)
	if err != nil {
		response.InternalError(c, "Failed to load visitor overview")
		return
	}
	response.Success(c, item)
}

func (h *VisitorAnalyticsHandler) Trend(c *gin.Context) {
	start, end, timezone, err := visitorDateRange(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	items, err := h.service.GetTrend(c.Request.Context(), start, end, timezone)
	if err != nil {
		response.InternalError(c, "Failed to load visitor trend")
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *VisitorAnalyticsHandler) ChannelStats(c *gin.Context) {
	start, end, _, err := visitorDateRange(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	items, err := h.service.GetChannelStats(c.Request.Context(), start, end)
	if err != nil {
		response.InternalError(c, "Failed to load channel statistics")
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *VisitorAnalyticsHandler) ListEvents(c *gin.Context) {
	start, end, _, err := visitorDateRange(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.ListEvents(c.Request.Context(), service.VisitorEventQuery{
		Start: start, End: end, ChannelCode: strings.TrimSpace(c.Query("channel_code")),
		IP: strings.TrimSpace(c.Query("ip")), Search: strings.TrimSpace(c.Query("search")),
		Page: page, PageSize: pageSize,
	})
	if err != nil {
		response.InternalError(c, "Failed to load visit records")
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *VisitorAnalyticsHandler) ListChannels(c *gin.Context) {
	items, err := h.service.ListChannels(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to load visitor channels")
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *VisitorAnalyticsHandler) CreateChannel(c *gin.Context) {
	var req visitorChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid channel payload")
		return
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	item, err := h.service.CreateChannel(c.Request.Context(), service.VisitorChannel{
		Name: req.Name, Code: req.Code, DestinationPath: req.DestinationPath,
		Description: req.Description, Active: active,
	})
	if err != nil {
		response.BadRequest(c, userSafeVisitorError(err))
		return
	}
	response.Created(c, item)
}

func (h *VisitorAnalyticsHandler) UpdateChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid channel ID")
		return
	}
	var req visitorChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid channel payload")
		return
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	item, err := h.service.UpdateChannel(c.Request.Context(), id, service.VisitorChannel{
		Name: req.Name, Code: req.Code, DestinationPath: req.DestinationPath,
		Description: req.Description, Active: active,
	})
	if errors.Is(err, sql.ErrNoRows) {
		response.NotFound(c, "Channel not found")
		return
	}
	if err != nil {
		response.BadRequest(c, userSafeVisitorError(err))
		return
	}
	response.Success(c, item)
}

func (h *VisitorAnalyticsHandler) DeleteChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid channel ID")
		return
	}
	err = h.service.DeleteChannel(c.Request.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		response.NotFound(c, "Channel not found")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to delete channel")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *VisitorAnalyticsHandler) LookupIP(c *gin.Context) {
	var req struct {
		IP string `json:"ip" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "IP address is required")
		return
	}
	item, err := h.service.LookupIP(c.Request.Context(), req.IP)
	if err != nil {
		response.Error(c, http.StatusBadGateway, userSafeVisitorError(err))
		return
	}
	response.Success(c, item)
}

func (h *VisitorAnalyticsHandler) GetSettings(c *gin.Context) {
	item, err := h.service.GetSettings(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to load visitor analytics settings")
		return
	}
	response.Success(c, item)
}

func (h *VisitorAnalyticsHandler) UpdateSettings(c *gin.Context) {
	var req struct {
		Enabled       bool `json:"enabled"`
		RetentionDays int  `json:"retention_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid settings payload")
		return
	}
	item, err := h.service.UpdateSettings(c.Request.Context(), req.Enabled, req.RetentionDays)
	if err != nil {
		response.BadRequest(c, userSafeVisitorError(err))
		return
	}
	response.Success(c, item)
}

func visitorDateRange(c *gin.Context) (time.Time, time.Time, string, error) {
	timezone := strings.TrimSpace(c.DefaultQuery("timezone", "Asia/Shanghai"))
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, time.Time{}, timezone, errors.New("invalid timezone")
	}
	today := time.Now().In(location)
	defaultStart := today.AddDate(0, 0, -6).Format("2006-01-02")
	defaultEnd := today.Format("2006-01-02")
	startDate, err := time.ParseInLocation("2006-01-02", c.DefaultQuery("start_date", defaultStart), location)
	if err != nil {
		return time.Time{}, time.Time{}, timezone, errors.New("invalid start_date")
	}
	endDate, err := time.ParseInLocation("2006-01-02", c.DefaultQuery("end_date", defaultEnd), location)
	if err != nil {
		return time.Time{}, time.Time{}, timezone, errors.New("invalid end_date")
	}
	endExclusive := endDate.AddDate(0, 0, 1)
	if endExclusive.Before(startDate) || endExclusive.Sub(startDate) > 367*24*time.Hour {
		return time.Time{}, time.Time{}, timezone, errors.New("date range must be between 1 and 366 days")
	}
	return startDate.UTC(), endExclusive.UTC(), timezone, nil
}

func userSafeVisitorError(err error) string {
	if err == nil {
		return "Unknown error"
	}
	message := err.Error()
	if strings.Contains(strings.ToLower(message), "duplicate key") {
		return "Channel code already exists"
	}
	return message
}
