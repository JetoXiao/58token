package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

const defaultIPGeolocationEndpoint = "https://ipwho.is"

var visitorChannelCodePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

type VisitorAnalyticsService struct {
	db              *sql.DB
	httpClient      *http.Client
	geoEndpoint     string
	cleanupMu       sync.Mutex
	lastCleanupTime time.Time
}

type VisitorAnalyticsSettings struct {
	Enabled       bool      `json:"enabled"`
	RetentionDays int       `json:"retention_days"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type VisitorChannel struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Code            string    `json:"code"`
	DestinationPath string    `json:"destination_path"`
	Description     string    `json:"description"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type VisitorTrackInput struct {
	UserID      int64
	ChannelCode string
	VisitorID   string
	SessionID   string
	IP          string
	CountryCode string
	Path        string
	Referrer    string
	LandingURL  string
	UserAgent   string
	Language    string
	Screen      string
}

type VisitorAnalyticsOverview struct {
	PageViews      int64 `json:"page_views"`
	UniqueVisitors int64 `json:"unique_visitors"`
	UniqueIPs      int64 `json:"unique_ips"`
	ActiveChannels int64 `json:"active_channels"`
}

type VisitorTrendPoint struct {
	Date           string `json:"date"`
	PageViews      int64  `json:"page_views"`
	UniqueVisitors int64  `json:"unique_visitors"`
	UniqueIPs      int64  `json:"unique_ips"`
}

type VisitorChannelStats struct {
	ID              *int64 `json:"id"`
	Name            string `json:"name"`
	Code            string `json:"code"`
	DestinationPath string `json:"destination_path"`
	Active          bool   `json:"active"`
	PageViews       int64  `json:"page_views"`
	UniqueVisitors  int64  `json:"unique_visitors"`
	UniqueIPs       int64  `json:"unique_ips"`
}

type VisitorEvent struct {
	ID            int64      `json:"id"`
	UserID        *int64     `json:"user_id,omitempty"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	ChannelName   string     `json:"channel_name"`
	ChannelCode   string     `json:"channel_code"`
	VisitorID     string     `json:"visitor_id"`
	SessionID     string     `json:"session_id"`
	IP            string     `json:"ip"`
	CountryCode   string     `json:"country_code"`
	Path          string     `json:"path"`
	Referrer      string     `json:"referrer"`
	LandingURL    string     `json:"landing_url"`
	UserAgent     string     `json:"user_agent"`
	Language      string     `json:"language"`
	Screen        string     `json:"screen"`
	IsBot         bool       `json:"is_bot"`
	OccurredAt    time.Time  `json:"occurred_at"`
	GeoCountry    string     `json:"geo_country"`
	GeoRegion     string     `json:"geo_region"`
	GeoCity       string     `json:"geo_city"`
	GeoResolvedAt *time.Time `json:"geo_resolved_at,omitempty"`
}

type VisitorEventQuery struct {
	Start       time.Time
	End         time.Time
	ChannelCode string
	IP          string
	Search      string
	Page        int
	PageSize    int
}

type IPGeolocation struct {
	IP          string    `json:"ip"`
	Country     string    `json:"country"`
	CountryCode string    `json:"country_code"`
	Region      string    `json:"region"`
	City        string    `json:"city"`
	Timezone    string    `json:"timezone"`
	Latitude    *float64  `json:"latitude,omitempty"`
	Longitude   *float64  `json:"longitude,omitempty"`
	Provider    string    `json:"provider"`
	ResolvedAt  time.Time `json:"resolved_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func NewVisitorAnalyticsService(db *sql.DB) *VisitorAnalyticsService {
	return &VisitorAnalyticsService{
		db:          db,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		geoEndpoint: defaultIPGeolocationEndpoint,
	}
}

func (s *VisitorAnalyticsService) GetSettings(ctx context.Context) (*VisitorAnalyticsSettings, error) {
	var settings VisitorAnalyticsSettings
	err := s.db.QueryRowContext(ctx, `
		SELECT enabled, retention_days, updated_at
		FROM visitor_analytics_settings WHERE id = 1
	`).Scan(&settings.Enabled, &settings.RetentionDays, &settings.UpdatedAt)
	return &settings, err
}

func (s *VisitorAnalyticsService) UpdateSettings(ctx context.Context, enabled bool, retentionDays int) (*VisitorAnalyticsSettings, error) {
	if retentionDays < 7 || retentionDays > 730 {
		return nil, errors.New("retention_days must be between 7 and 730")
	}
	var settings VisitorAnalyticsSettings
	err := s.db.QueryRowContext(ctx, `
		UPDATE visitor_analytics_settings
		SET enabled = $1, retention_days = $2, updated_at = NOW()
		WHERE id = 1
		RETURNING enabled, retention_days, updated_at
	`, enabled, retentionDays).Scan(&settings.Enabled, &settings.RetentionDays, &settings.UpdatedAt)
	return &settings, err
}

func (s *VisitorAnalyticsService) Track(ctx context.Context, input VisitorTrackInput) error {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return err
	}
	if !settings.Enabled {
		return nil
	}

	channelCode := normalizeVisitorChannelCode(input.ChannelCode)
	var channelID sql.NullInt64
	if channelCode != "direct" {
		err = s.db.QueryRowContext(ctx, `SELECT id FROM visitor_channels WHERE code = $1 AND active = TRUE`, channelCode).Scan(&channelID)
		if errors.Is(err, sql.ErrNoRows) {
			channelCode = "direct"
			channelID = sql.NullInt64{}
		} else if err != nil {
			return err
		}
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO visitor_events (
			channel_id, channel_code, visitor_id, session_id, user_id, ip, country_code,
			path, referrer, landing_url, user_agent, language, screen, is_bot
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, nullableInt64(channelID), channelCode, clamp(input.VisitorID, 64), clamp(input.SessionID, 64),
		nullablePositiveInt64(input.UserID), clamp(input.IP, 45), strings.ToUpper(clamp(input.CountryCode, 8)), cleanPath(input.Path),
		clamp(input.Referrer, 1024), clamp(input.LandingURL, 1024), clamp(input.UserAgent, 512),
		clamp(input.Language, 32), clamp(input.Screen, 32), looksLikeBot(input.UserAgent))
	if err == nil {
		s.cleanupExpiredEvents(ctx, settings.RetentionDays)
	}
	return err
}

func (s *VisitorAnalyticsService) GetOverview(ctx context.Context, start, end time.Time) (*VisitorAnalyticsOverview, error) {
	var result VisitorAnalyticsOverview
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COUNT(DISTINCT CASE WHEN visitor_id <> '' THEN visitor_id ELSE 'ip:' || ip END),
		       COUNT(DISTINCT ip),
		       (SELECT COUNT(*) FROM visitor_channels WHERE active = TRUE)
		FROM visitor_events
		WHERE occurred_at >= $1 AND occurred_at < $2
	`, start, end).Scan(&result.PageViews, &result.UniqueVisitors, &result.UniqueIPs, &result.ActiveChannels)
	return &result, err
}

func (s *VisitorAnalyticsService) GetTrend(ctx context.Context, start, end time.Time, timezone string) ([]VisitorTrendPoint, error) {
	rows, err := s.db.QueryContext(ctx, `
		WITH days AS (
			SELECT generate_series(
				(timezone($3, $1::timestamptz))::date,
				(timezone($3, ($2::timestamptz - interval '1 second')))::date,
				interval '1 day'
			)::date AS day
		)
		SELECT to_char(d.day, 'YYYY-MM-DD'),
		       COUNT(e.id),
		       COUNT(DISTINCT CASE WHEN e.visitor_id <> '' THEN e.visitor_id ELSE 'ip:' || e.ip END),
		       COUNT(DISTINCT e.ip)
		FROM days d
		LEFT JOIN visitor_events e
		  ON (timezone($3, e.occurred_at))::date = d.day
		 AND e.occurred_at >= $1 AND e.occurred_at < $2
		GROUP BY d.day ORDER BY d.day
	`, start, end, timezone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]VisitorTrendPoint, 0)
	for rows.Next() {
		var item VisitorTrendPoint
		if err := rows.Scan(&item.Date, &item.PageViews, &item.UniqueVisitors, &item.UniqueIPs); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *VisitorAnalyticsService) GetChannelStats(ctx context.Context, start, end time.Time) ([]VisitorChannelStats, error) {
	rows, err := s.db.QueryContext(ctx, `
		WITH event_stats AS (
			SELECT channel_code,
			       COUNT(*) AS page_views,
			       COUNT(DISTINCT CASE WHEN visitor_id <> '' THEN visitor_id ELSE 'ip:' || ip END) AS unique_visitors,
			       COUNT(DISTINCT ip) AS unique_ips
			FROM visitor_events
			WHERE occurred_at >= $1 AND occurred_at < $2
			GROUP BY channel_code
		), channel_catalog AS (
			SELECT id, name, code, destination_path, active FROM visitor_channels
			UNION ALL
			SELECT NULL::BIGINT, 'Direct', 'direct', '/home', TRUE
			UNION ALL
			SELECT NULL::BIGINT, stats.channel_code, stats.channel_code, '/home', FALSE
			FROM event_stats stats
			WHERE stats.channel_code <> 'direct'
			  AND NOT EXISTS (SELECT 1 FROM visitor_channels c WHERE c.code = stats.channel_code)
		)
		SELECT catalog.id, catalog.name, catalog.code, catalog.destination_path, catalog.active,
		       COALESCE(stats.page_views, 0), COALESCE(stats.unique_visitors, 0), COALESCE(stats.unique_ips, 0)
		FROM channel_catalog catalog
		LEFT JOIN event_stats stats ON stats.channel_code = catalog.code
		ORDER BY COALESCE(stats.page_views, 0) DESC, catalog.name
	`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]VisitorChannelStats, 0)
	for rows.Next() {
		var item VisitorChannelStats
		var id sql.NullInt64
		if err := rows.Scan(&id, &item.Name, &item.Code, &item.DestinationPath, &item.Active, &item.PageViews, &item.UniqueVisitors, &item.UniqueIPs); err != nil {
			return nil, err
		}
		if id.Valid {
			value := id.Int64
			item.ID = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *VisitorAnalyticsService) ListChannels(ctx context.Context) ([]VisitorChannel, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, code, destination_path, description, active, created_at, updated_at
		FROM visitor_channels ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]VisitorChannel, 0)
	for rows.Next() {
		var item VisitorChannel
		if err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.DestinationPath, &item.Description, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *VisitorAnalyticsService) CreateChannel(ctx context.Context, input VisitorChannel) (*VisitorChannel, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Code = strings.ToLower(strings.TrimSpace(input.Code))
	input.DestinationPath = cleanPath(input.DestinationPath)
	if input.Name == "" || !visitorChannelCodePattern.MatchString(input.Code) || input.Code == "direct" {
		return nil, errors.New("invalid channel name or code")
	}
	var item VisitorChannel
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO visitor_channels (name, code, destination_path, description, active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, code, destination_path, description, active, created_at, updated_at
	`, clamp(input.Name, 100), input.Code, input.DestinationPath, clamp(input.Description, 500), input.Active).
		Scan(&item.ID, &item.Name, &item.Code, &item.DestinationPath, &item.Description, &item.Active, &item.CreatedAt, &item.UpdatedAt)
	return &item, err
}

func (s *VisitorAnalyticsService) UpdateChannel(ctx context.Context, id int64, input VisitorChannel) (*VisitorChannel, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return nil, errors.New("invalid channel name or code")
	}
	var item VisitorChannel
	err := s.db.QueryRowContext(ctx, `
		UPDATE visitor_channels SET name = $2, destination_path = $3,
		       description = $4, active = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, code, destination_path, description, active, created_at, updated_at
	`, id, clamp(input.Name, 100), cleanPath(input.DestinationPath), clamp(input.Description, 500), input.Active).
		Scan(&item.ID, &item.Name, &item.Code, &item.DestinationPath, &item.Description, &item.Active, &item.CreatedAt, &item.UpdatedAt)
	return &item, err
}

func (s *VisitorAnalyticsService) DeleteChannel(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM visitor_channels WHERE id = $1`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (s *VisitorAnalyticsService) ListEvents(ctx context.Context, query VisitorEventQuery) ([]VisitorEvent, int64, error) {
	where, args := visitorEventWhere(query)
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM visitor_events e `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (query.Page - 1) * query.PageSize
	args = append(args, query.PageSize, offset)
	limitArg, offsetArg := len(args)-1, len(args)
	rows, err := s.db.QueryContext(ctx, `
		SELECT e.id, e.user_id, COALESCE(u.username, ''), COALESCE(u.email, ''),
		       COALESCE(c.name, CASE WHEN e.channel_code = 'direct' THEN 'Direct' ELSE e.channel_code END),
		       e.channel_code, e.visitor_id, e.session_id, e.ip, e.country_code, e.path,
		       e.referrer, e.landing_url, e.user_agent, e.language, e.screen, e.is_bot, e.occurred_at,
		       COALESCE(g.country, ''), COALESCE(g.region, ''), COALESCE(g.city, ''), g.resolved_at
		FROM visitor_events e
		LEFT JOIN visitor_channels c ON c.code = e.channel_code
		LEFT JOIN users u ON u.id = e.user_id
		LEFT JOIN visitor_ip_geolocation_cache g ON g.ip = e.ip AND g.expires_at > NOW()
	`+where+fmt.Sprintf(` ORDER BY e.occurred_at DESC LIMIT $%d OFFSET $%d`, limitArg, offsetArg), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]VisitorEvent, 0, query.PageSize)
	for rows.Next() {
		var item VisitorEvent
		var userID sql.NullInt64
		var resolvedAt sql.NullTime
		if err := rows.Scan(&item.ID, &userID, &item.Username, &item.Email, &item.ChannelName, &item.ChannelCode, &item.VisitorID, &item.SessionID,
			&item.IP, &item.CountryCode, &item.Path, &item.Referrer, &item.LandingURL, &item.UserAgent,
			&item.Language, &item.Screen, &item.IsBot, &item.OccurredAt, &item.GeoCountry, &item.GeoRegion,
			&item.GeoCity, &resolvedAt); err != nil {
			return nil, 0, err
		}
		if userID.Valid {
			value := userID.Int64
			item.UserID = &value
		}
		if resolvedAt.Valid {
			value := resolvedAt.Time
			item.GeoResolvedAt = &value
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *VisitorAnalyticsService) LookupIP(ctx context.Context, rawIP string) (*IPGeolocation, error) {
	ip := net.ParseIP(strings.TrimSpace(rawIP))
	if ip == nil {
		return nil, errors.New("invalid IP address")
	}
	normalized := ip.String()
	if isNonPublicIP(ip) {
		now := time.Now()
		return &IPGeolocation{IP: normalized, Country: "Private network", Provider: "local", ResolvedAt: now, ExpiresAt: now.Add(24 * time.Hour)}, nil
	}
	if cached, err := s.getCachedIPLocation(ctx, normalized); err == nil && cached != nil {
		return cached, nil
	}

	requestURL := strings.TrimRight(s.geoEndpoint, "/") + "/" + url.PathEscape(normalized)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("IP geolocation service unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IP geolocation service returned %s", resp.Status)
	}
	var payload struct {
		Success     bool    `json:"success"`
		Message     string  `json:"message"`
		Country     string  `json:"country"`
		CountryCode string  `json:"country_code"`
		Region      string  `json:"region"`
		City        string  `json:"city"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Timezone    struct {
			ID string `json:"id"`
		} `json:"timezone"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if !payload.Success {
		return nil, fmt.Errorf("IP geolocation failed: %s", payload.Message)
	}
	now := time.Now()
	lat, lng := payload.Latitude, payload.Longitude
	result := &IPGeolocation{
		IP: normalized, Country: payload.Country, CountryCode: payload.CountryCode,
		Region: payload.Region, City: payload.City, Timezone: payload.Timezone.ID,
		Latitude: &lat, Longitude: &lng, Provider: "ipwho.is", ResolvedAt: now, ExpiresAt: now.Add(7 * 24 * time.Hour),
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO visitor_ip_geolocation_cache (
			ip, country, country_code, region, city, timezone, latitude, longitude, provider, resolved_at, expires_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (ip) DO UPDATE SET country = EXCLUDED.country, country_code = EXCLUDED.country_code,
			region = EXCLUDED.region, city = EXCLUDED.city, timezone = EXCLUDED.timezone,
			latitude = EXCLUDED.latitude, longitude = EXCLUDED.longitude, provider = EXCLUDED.provider,
			resolved_at = EXCLUDED.resolved_at, expires_at = EXCLUDED.expires_at
	`, result.IP, result.Country, result.CountryCode, result.Region, result.City, result.Timezone,
		result.Latitude, result.Longitude, result.Provider, result.ResolvedAt, result.ExpiresAt)
	return result, err
}

func (s *VisitorAnalyticsService) getCachedIPLocation(ctx context.Context, ip string) (*IPGeolocation, error) {
	var result IPGeolocation
	result.IP = ip
	err := s.db.QueryRowContext(ctx, `
		SELECT country, country_code, region, city, timezone, latitude, longitude, provider, resolved_at, expires_at
		FROM visitor_ip_geolocation_cache WHERE ip = $1 AND expires_at > NOW()
	`, ip).Scan(&result.Country, &result.CountryCode, &result.Region, &result.City, &result.Timezone,
		&result.Latitude, &result.Longitude, &result.Provider, &result.ResolvedAt, &result.ExpiresAt)
	return &result, err
}

func (s *VisitorAnalyticsService) cleanupExpiredEvents(ctx context.Context, retentionDays int) {
	s.cleanupMu.Lock()
	defer s.cleanupMu.Unlock()
	if time.Since(s.lastCleanupTime) < time.Hour {
		return
	}
	s.lastCleanupTime = time.Now()
	_, _ = s.db.ExecContext(ctx, `DELETE FROM visitor_events WHERE occurred_at < NOW() - ($1 * interval '1 day')`, retentionDays)
	_, _ = s.db.ExecContext(ctx, `DELETE FROM visitor_ip_geolocation_cache WHERE expires_at < NOW() - interval '7 days'`)
}

func visitorEventWhere(query VisitorEventQuery) (string, []any) {
	clauses := []string{"e.occurred_at >= $1", "e.occurred_at < $2"}
	args := []any{query.Start, query.End}
	if query.ChannelCode != "" {
		args = append(args, query.ChannelCode)
		clauses = append(clauses, fmt.Sprintf("e.channel_code = $%d", len(args)))
	}
	if query.IP != "" {
		args = append(args, query.IP)
		clauses = append(clauses, fmt.Sprintf("e.ip = $%d", len(args)))
	}
	if query.Search != "" {
		args = append(args, "%"+query.Search+"%")
		clauses = append(clauses, fmt.Sprintf("(e.path ILIKE $%d OR e.referrer ILIKE $%d OR e.landing_url ILIKE $%d)", len(args), len(args), len(args)))
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func normalizeVisitorChannelCode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || !visitorChannelCodePattern.MatchString(value) {
		return "direct"
	}
	return value
}

func cleanPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return "/home"
	}
	return clamp(value, 512)
}

func clamp(value string, max int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > max {
		return string(runes[:max])
	}
	return value
}

func nullableInt64(value sql.NullInt64) any {
	if value.Valid {
		return value.Int64
	}
	return nil
}

func nullablePositiveInt64(value int64) any {
	if value <= 0 {
		return nil
	}
	return value
}

func looksLikeBot(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	for _, marker := range []string{"bot", "crawler", "spider", "slurp", "headless", "preview"} {
		if strings.Contains(ua, marker) {
			return true
		}
	}
	return false
}

func isNonPublicIP(ip net.IP) bool {
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}
