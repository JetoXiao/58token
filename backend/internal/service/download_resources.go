package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	settingKeyDownloadResourceS3Config = "download_resource_s3_config"
	downloadResourceURLTTL             = 2 * time.Minute
	downloadResourceUploadURLTTL       = 15 * time.Minute
	downloadResourceS3TestTimeout      = 10 * time.Second
	maxDownloadResourceUploadBytes     = int64(2 * 1024 * 1024 * 1024)
)

var (
	downloadResourceSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	ErrDownloadResourceNotFound = errors.New("download resource not found")
	ErrDownloadRateLimited      = errors.New("download request rate limited")
	ErrDownloadStorageNotReady  = errors.New("download storage is not configured")
)

// DownloadResourceS3Config is intentionally separate from backup storage. A
// download-only R2 token can therefore be scoped to this bucket and prefix.
type DownloadResourceS3Config struct {
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key,omitempty"`
	Prefix          string `json:"prefix"`
	ForcePathStyle  bool   `json:"force_path_style"`
}

func (c DownloadResourceS3Config) IsConfigured() bool {
	return strings.TrimSpace(c.Endpoint) != "" && strings.TrimSpace(c.Bucket) != "" &&
		strings.TrimSpace(c.AccessKeyID) != "" && strings.TrimSpace(c.SecretAccessKey) != ""
}

type DownloadResourceObjectMetadata struct {
	SizeBytes   int64
	ContentType string
	UploadedAt  time.Time
}

type DownloadResourceObjectStore interface {
	HeadBucket(ctx context.Context) error
	HeadObject(ctx context.Context, key string) (DownloadResourceObjectMetadata, error)
	PresignDownload(ctx context.Context, key, fileName, contentType string, expiry time.Duration) (string, error)
	PresignUpload(ctx context.Context, key, contentType string, expiry time.Duration) (string, error)
}

type DownloadResourceObjectStoreFactory func(ctx context.Context, cfg DownloadResourceS3Config) (DownloadResourceObjectStore, error)

type DownloadResource struct {
	ID             int64     `json:"id"`
	Slug           string    `json:"slug"`
	NameZh         string    `json:"name_zh"`
	NameEn         string    `json:"name_en"`
	DescriptionZh  string    `json:"description_zh"`
	DescriptionEn  string    `json:"description_en"`
	Version        string    `json:"version"`
	Platform       string    `json:"platform"`
	ObjectKey      string    `json:"object_key,omitempty"`
	FileName       string    `json:"file_name"`
	ContentType    string    `json:"content_type"`
	SizeBytes      int64     `json:"size_bytes"`
	ChecksumSHA256 string    `json:"checksum_sha256,omitempty"`
	Published      bool      `json:"published"`
	SortOrder      int       `json:"sort_order"`
	DownloadCount  int64     `json:"download_count"`
	UploadedAt     time.Time `json:"uploaded_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DownloadResourceInput struct {
	Slug           string `json:"slug"`
	NameZh         string `json:"name_zh"`
	NameEn         string `json:"name_en"`
	DescriptionZh  string `json:"description_zh"`
	DescriptionEn  string `json:"description_en"`
	Version        string `json:"version"`
	Platform       string `json:"platform"`
	ObjectKey      string `json:"object_key"`
	FileName       string `json:"file_name"`
	ContentType    string `json:"content_type"`
	ChecksumSHA256 string `json:"checksum_sha256"`
	Published      bool   `json:"published"`
	SortOrder      int    `json:"sort_order"`
}

type DownloadResourceUploadRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type DownloadResourceUploadURL struct {
	ObjectKey string    `json:"object_key"`
	UploadURL string    `json:"upload_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type DownloadResourceDownloadRecord struct {
	ID           int64     `json:"id"`
	UserID       *int64    `json:"user_id,omitempty"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	ResourceID   int64     `json:"resource_id"`
	ResourceName string    `json:"resource_name"`
	Version      string    `json:"version"`
	IP           string    `json:"ip"`
	UserAgent    string    `json:"user_agent"`
	Referrer     string    `json:"referrer"`
	RequestedAt  time.Time `json:"requested_at"`
	GeoCountry   string    `json:"geo_country"`
	GeoRegion    string    `json:"geo_region"`
	GeoCity      string    `json:"geo_city"`
}

type DownloadResourceService struct {
	db           *sql.DB
	settings     SettingRepository
	storeFactory DownloadResourceObjectStoreFactory
	redis        *redis.Client
	encryptor    SecretEncryptor
}

func NewDownloadResourceService(
	db *sql.DB,
	settings SettingRepository,
	storeFactory DownloadResourceObjectStoreFactory,
	redisClient *redis.Client,
	encryptor SecretEncryptor,
) *DownloadResourceService {
	return &DownloadResourceService{
		db: db, settings: settings, storeFactory: storeFactory, redis: redisClient, encryptor: encryptor,
	}
}

func (s *DownloadResourceService) GetS3Config(ctx context.Context) (*DownloadResourceS3Config, error) {
	cfg, err := s.loadS3Config(ctx)
	if err != nil || cfg == nil {
		return cfg, err
	}
	cfg.SecretAccessKey = ""
	return cfg, nil
}

func (s *DownloadResourceService) UpdateS3Config(ctx context.Context, cfg DownloadResourceS3Config) (*DownloadResourceS3Config, error) {
	cfg = normalizeDownloadResourceS3Config(cfg)
	if cfg.SecretAccessKey == "" {
		old, err := s.loadS3Config(ctx)
		if err != nil {
			return nil, err
		}
		if old != nil {
			cfg.SecretAccessKey = old.SecretAccessKey
		}
	}
	if !cfg.IsConfigured() {
		return nil, errors.New("endpoint, bucket, access key and secret are required")
	}
	if s.encryptor == nil {
		return nil, errors.New("download storage encryption is unavailable")
	}
	ciphertext, err := s.encryptor.Encrypt(cfg.SecretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt download storage secret: %w", err)
	}
	cfg.SecretAccessKey = ciphertext
	payload, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("encode download storage config: %w", err)
	}
	if err := s.settings.Set(ctx, settingKeyDownloadResourceS3Config, string(payload)); err != nil {
		return nil, fmt.Errorf("save download storage config: %w", err)
	}
	cfg.SecretAccessKey = ""
	return &cfg, nil
}

func (s *DownloadResourceService) TestS3Connection(ctx context.Context, cfg DownloadResourceS3Config) error {
	cfg = normalizeDownloadResourceS3Config(cfg)
	if cfg.SecretAccessKey == "" {
		old, err := s.loadS3Config(ctx)
		if err != nil {
			return err
		}
		if old != nil {
			cfg.SecretAccessKey = old.SecretAccessKey
		}
	}
	if !cfg.IsConfigured() {
		return ErrDownloadStorageNotReady
	}
	store, err := s.newStore(ctx, cfg)
	if err != nil {
		return err
	}

	// An unreachable R2 endpoint must not leave the administrator waiting for the
	// SDK's transport-level timeout, which can be substantially longer.
	testCtx, cancel := context.WithTimeout(ctx, downloadResourceS3TestTimeout)
	defer cancel()
	return store.HeadBucket(testCtx)
}

func (s *DownloadResourceService) ListPublished(ctx context.Context) ([]DownloadResource, error) {
	return s.listResources(ctx, true)
}

func (s *DownloadResourceService) ListAll(ctx context.Context) ([]DownloadResource, error) {
	return s.listResources(ctx, false)
}

func (s *DownloadResourceService) Create(ctx context.Context, input DownloadResourceInput) (*DownloadResource, error) {
	cfg, err := s.requireS3Config(ctx)
	if err != nil {
		return nil, err
	}
	input, err = normalizeDownloadResourceInput(input, cfg.Prefix)
	if err != nil {
		return nil, err
	}
	metadata, err := s.headObject(ctx, cfg, input.ObjectKey)
	if err != nil {
		return nil, err
	}
	var item DownloadResource
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO download_resources (
			slug, name_zh, name_en, description_zh, description_en, version, platform,
			object_key, file_name, content_type, size_bytes, checksum_sha256, published, sort_order, uploaded_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING `+downloadResourceColumns, input.Slug, input.NameZh, input.NameEn, input.DescriptionZh,
		input.DescriptionEn, input.Version, input.Platform, input.ObjectKey, input.FileName,
		firstDownloadResourceNonEmpty(input.ContentType, metadata.ContentType, "application/octet-stream"), metadata.SizeBytes,
		input.ChecksumSHA256, input.Published, input.SortOrder, metadata.UploadedAt).Scan(downloadResourceScanTargets(&item)...)
	return &item, err
}

func (s *DownloadResourceService) Update(ctx context.Context, id int64, input DownloadResourceInput) (*DownloadResource, error) {
	cfg, err := s.requireS3Config(ctx)
	if err != nil {
		return nil, err
	}
	input, err = normalizeDownloadResourceInput(input, cfg.Prefix)
	if err != nil {
		return nil, err
	}
	metadata, err := s.headObject(ctx, cfg, input.ObjectKey)
	if err != nil {
		return nil, err
	}
	var item DownloadResource
	err = s.db.QueryRowContext(ctx, `
		UPDATE download_resources SET
			slug=$2, name_zh=$3, name_en=$4, description_zh=$5, description_en=$6, version=$7, platform=$8,
			object_key=$9, file_name=$10, content_type=$11, size_bytes=$12, checksum_sha256=$13,
			published=$14, sort_order=$15, uploaded_at=$16, updated_at=NOW()
		WHERE id=$1
		RETURNING `+downloadResourceColumns, id, input.Slug, input.NameZh, input.NameEn, input.DescriptionZh,
		input.DescriptionEn, input.Version, input.Platform, input.ObjectKey, input.FileName,
		firstDownloadResourceNonEmpty(input.ContentType, metadata.ContentType, "application/octet-stream"), metadata.SizeBytes,
		input.ChecksumSHA256, input.Published, input.SortOrder, metadata.UploadedAt).Scan(downloadResourceScanTargets(&item)...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDownloadResourceNotFound
	}
	return &item, err
}

func (s *DownloadResourceService) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM download_resources WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrDownloadResourceNotFound
	}
	return nil
}

func (s *DownloadResourceService) CreateUploadURL(ctx context.Context, request DownloadResourceUploadRequest) (*DownloadResourceUploadURL, error) {
	cfg, err := s.requireS3Config(ctx)
	if err != nil {
		return nil, err
	}
	fileName := sanitizeDownloadResourceFileName(request.FileName)
	if fileName == "" {
		return nil, errors.New("file name is required")
	}
	if request.SizeBytes <= 0 || request.SizeBytes > maxDownloadResourceUploadBytes {
		return nil, fmt.Errorf("file size must be between 1 byte and %d bytes", maxDownloadResourceUploadBytes)
	}
	contentType := normalizeDownloadResourceContentType(request.ContentType)
	objectKey := cfg.Prefix + time.Now().UTC().Format("2006/01/02/") + uuid.NewString() + "-" + fileName
	store, err := s.newStore(ctx, cfg)
	if err != nil {
		return nil, err
	}
	url, err := store.PresignUpload(ctx, objectKey, contentType, downloadResourceUploadURLTTL)
	if err != nil {
		return nil, err
	}
	return &DownloadResourceUploadURL{ObjectKey: objectKey, UploadURL: url, ExpiresAt: time.Now().Add(downloadResourceUploadURLTTL)}, nil
}

func (s *DownloadResourceService) IssueDownload(ctx context.Context, id, userID int64, ip, userAgent, referrer string) (string, error) {
	item, err := s.getPublishedResource(ctx, id)
	if err != nil {
		return "", err
	}
	cfg, err := s.requireS3Config(ctx)
	if err != nil {
		return "", err
	}
	if err := s.reserveDownload(ctx, item.ID, ip); err != nil {
		return "", err
	}
	store, err := s.newStore(ctx, cfg)
	if err != nil {
		return "", err
	}
	url, err := store.PresignDownload(ctx, item.ObjectKey, item.FileName, item.ContentType, downloadResourceURLTTL)
	if err != nil {
		return "", err
	}
	if err := s.recordDownload(ctx, item.ID, userID, ip, userAgent, referrer); err != nil {
		return "", err
	}
	return url, nil
}

func (s *DownloadResourceService) ListDownloads(ctx context.Context, page, pageSize int) ([]DownloadResourceDownloadRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM download_resource_downloads`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT d.id, d.user_id, COALESCE(u.username, ''), COALESCE(u.email, ''),
		       d.resource_id, COALESCE(NULLIF(r.name_en, ''), r.name_zh), r.version,
		       d.ip, d.user_agent, d.referrer, d.requested_at,
		       COALESCE(g.country, ''), COALESCE(g.region, ''), COALESCE(g.city, '')
		FROM download_resource_downloads d
		JOIN download_resources r ON r.id = d.resource_id
		LEFT JOIN users u ON u.id = d.user_id
		LEFT JOIN visitor_ip_geolocation_cache g ON g.ip = d.ip AND g.expires_at > NOW()
		ORDER BY d.requested_at DESC
		LIMIT $1 OFFSET $2`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]DownloadResourceDownloadRecord, 0, pageSize)
	for rows.Next() {
		var item DownloadResourceDownloadRecord
		var userID sql.NullInt64
		if err := rows.Scan(&item.ID, &userID, &item.Username, &item.Email, &item.ResourceID, &item.ResourceName, &item.Version,
			&item.IP, &item.UserAgent, &item.Referrer, &item.RequestedAt, &item.GeoCountry, &item.GeoRegion, &item.GeoCity); err != nil {
			return nil, 0, err
		}
		if userID.Valid {
			value := userID.Int64
			item.UserID = &value
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

const downloadResourceColumns = `id, slug, name_zh, name_en, description_zh, description_en, version, platform,
	object_key, file_name, content_type, size_bytes, checksum_sha256, published, sort_order, download_count,
	uploaded_at, created_at, updated_at`

func (s *DownloadResourceService) listResources(ctx context.Context, publishedOnly bool) ([]DownloadResource, error) {
	query := `SELECT ` + downloadResourceColumns + ` FROM download_resources`
	if publishedOnly {
		query += ` WHERE published = TRUE`
	}
	query += ` ORDER BY sort_order DESC, uploaded_at DESC, id DESC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]DownloadResource, 0)
	for rows.Next() {
		var item DownloadResource
		if err := rows.Scan(downloadResourceScanTargets(&item)...); err != nil {
			return nil, err
		}
		if publishedOnly {
			item.ObjectKey = ""
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *DownloadResourceService) getPublishedResource(ctx context.Context, id int64) (*DownloadResource, error) {
	var item DownloadResource
	err := s.db.QueryRowContext(ctx, `SELECT `+downloadResourceColumns+` FROM download_resources WHERE id = $1 AND published = TRUE`, id).Scan(downloadResourceScanTargets(&item)...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDownloadResourceNotFound
	}
	return &item, err
}

func (s *DownloadResourceService) recordDownload(ctx context.Context, resourceID, userID int64, ip, userAgent, referrer string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `UPDATE download_resources SET download_count = download_count + 1, updated_at = NOW() WHERE id = $1`, resourceID); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO download_resource_downloads (resource_id, user_id, ip, user_agent, referrer)
		VALUES ($1, $2, $3, $4, $5)`, resourceID, nullableDownloadUserID(userID), clampDownloadResourceValue(ip, 45), clampDownloadResourceValue(userAgent, 512), clampDownloadResourceValue(referrer, 1024))
	if err != nil {
		return err
	}
	return tx.Commit()
}

func nullableDownloadUserID(userID int64) any {
	if userID <= 0 {
		return nil
	}
	return userID
}

func (s *DownloadResourceService) reserveDownload(ctx context.Context, resourceID int64, ip string) error {
	if s.redis == nil || strings.TrimSpace(ip) == "" {
		return errors.New("download protection is temporarily unavailable")
	}
	checks := []struct {
		key    string
		limit  int64
		window time.Duration
	}{
		{key: fmt.Sprintf("download_resource:resource:%d:ip:%s", resourceID, ip), limit: 3, window: 10 * time.Minute},
		{key: "download_resource:ip:" + ip + ":hour", limit: 12, window: time.Hour},
		{key: "download_resource:ip:" + ip + ":day", limit: 40, window: 24 * time.Hour},
	}
	for _, check := range checks {
		count, err := incrementDownloadRateCounter(ctx, s.redis, check.key, check.window)
		if err != nil {
			return errors.New("download protection is temporarily unavailable")
		}
		if count > check.limit {
			return ErrDownloadRateLimited
		}
	}
	return nil
}

var downloadRateCounterScript = redis.NewScript(`
local current = redis.call('INCR', KEYS[1])
local ttl = redis.call('PTTL', KEYS[1])
if current == 1 or ttl == -1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return current
`)

func incrementDownloadRateCounter(ctx context.Context, client *redis.Client, key string, window time.Duration) (int64, error) {
	value, err := downloadRateCounterScript.Run(ctx, client, []string{key}, window.Milliseconds()).Int64()
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (s *DownloadResourceService) requireS3Config(ctx context.Context) (DownloadResourceS3Config, error) {
	cfg, err := s.loadS3Config(ctx)
	if err != nil {
		return DownloadResourceS3Config{}, err
	}
	if cfg == nil || !cfg.IsConfigured() {
		return DownloadResourceS3Config{}, ErrDownloadStorageNotReady
	}
	return *cfg, nil
}

func (s *DownloadResourceService) loadS3Config(ctx context.Context) (*DownloadResourceS3Config, error) {
	if s.settings == nil {
		return nil, errors.New("download storage settings are unavailable")
	}
	raw, err := s.settings.GetValue(ctx, settingKeyDownloadResourceS3Config)
	if errors.Is(err, ErrSettingNotFound) || strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg DownloadResourceS3Config
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("decode download storage config: %w", err)
	}
	cfg = normalizeDownloadResourceS3Config(cfg)
	if cfg.SecretAccessKey != "" {
		if s.encryptor == nil {
			return nil, errors.New("download storage encryption is unavailable")
		}
		secret, err := s.encryptor.Decrypt(cfg.SecretAccessKey)
		if err != nil {
			return nil, fmt.Errorf("decrypt download storage secret: %w", err)
		}
		cfg.SecretAccessKey = secret
	}
	return &cfg, nil
}

func (s *DownloadResourceService) newStore(ctx context.Context, cfg DownloadResourceS3Config) (DownloadResourceObjectStore, error) {
	if s.storeFactory == nil {
		return nil, errors.New("download storage factory is unavailable")
	}
	return s.storeFactory(ctx, cfg)
}

func (s *DownloadResourceService) headObject(ctx context.Context, cfg DownloadResourceS3Config, key string) (DownloadResourceObjectMetadata, error) {
	store, err := s.newStore(ctx, cfg)
	if err != nil {
		return DownloadResourceObjectMetadata{}, err
	}
	metadata, err := store.HeadObject(ctx, key)
	if err != nil {
		return DownloadResourceObjectMetadata{}, fmt.Errorf("verify R2 object: %w", err)
	}
	if metadata.SizeBytes <= 0 {
		return DownloadResourceObjectMetadata{}, errors.New("R2 object is empty")
	}
	if metadata.SizeBytes > maxDownloadResourceUploadBytes {
		return DownloadResourceObjectMetadata{}, fmt.Errorf("R2 object exceeds the %d byte download resource limit", maxDownloadResourceUploadBytes)
	}
	if metadata.UploadedAt.IsZero() {
		metadata.UploadedAt = time.Now()
	}
	return metadata, nil
}

func normalizeDownloadResourceS3Config(cfg DownloadResourceS3Config) DownloadResourceS3Config {
	cfg.Endpoint = strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/")
	cfg.Region = strings.TrimSpace(cfg.Region)
	if cfg.Region == "" {
		cfg.Region = "auto"
	}
	cfg.Bucket = strings.TrimSpace(cfg.Bucket)
	cfg.AccessKeyID = strings.TrimSpace(cfg.AccessKeyID)
	cfg.SecretAccessKey = strings.TrimSpace(cfg.SecretAccessKey)
	cfg.Prefix = normalizeDownloadResourcePrefix(cfg.Prefix)
	return cfg
}

func normalizeDownloadResourcePrefix(prefix string) string {
	prefix = strings.Trim(strings.TrimSpace(prefix), "/")
	if prefix == "" {
		return "downloads/"
	}
	return prefix + "/"
}

func normalizeDownloadResourceInput(input DownloadResourceInput, prefix string) (DownloadResourceInput, error) {
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	if !downloadResourceSlugPattern.MatchString(input.Slug) {
		return input, errors.New("slug must use lowercase letters, numbers, and hyphens")
	}
	input.NameZh = clampDownloadResourceValue(input.NameZh, 160)
	input.NameEn = clampDownloadResourceValue(input.NameEn, 160)
	if input.NameZh == "" && input.NameEn == "" {
		return input, errors.New("at least one resource name is required")
	}
	input.DescriptionZh = clampDownloadResourceValue(input.DescriptionZh, 1000)
	input.DescriptionEn = clampDownloadResourceValue(input.DescriptionEn, 1000)
	input.Version = clampDownloadResourceValue(input.Version, 64)
	input.Platform = clampDownloadResourceValue(input.Platform, 64)
	input.FileName = sanitizeDownloadResourceFileName(input.FileName)
	if input.FileName == "" {
		return input, errors.New("file name is required")
	}
	input.ObjectKey = strings.TrimSpace(input.ObjectKey)
	if !isSafeDownloadResourceKey(input.ObjectKey, prefix) {
		return input, errors.New("object key must stay within the configured downloads prefix")
	}
	input.ContentType = normalizeDownloadResourceContentType(input.ContentType)
	input.ChecksumSHA256 = strings.ToLower(strings.TrimSpace(input.ChecksumSHA256))
	if input.ChecksumSHA256 != "" && !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(input.ChecksumSHA256) {
		return input, errors.New("checksum must be a SHA-256 hex value")
	}
	return input, nil
}

func isSafeDownloadResourceKey(key, prefix string) bool {
	if key == "" || strings.Contains(key, "\\") || strings.Contains(key, "..") || !strings.HasPrefix(key, prefix) {
		return false
	}
	return path.Clean(key) == key
}

func sanitizeDownloadResourceFileName(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = path.Base(value)
	value = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 || r == '"' || r == '\\' || r == '/' {
			return -1
		}
		return r
	}, value)
	return clampDownloadResourceValue(value, 180)
}

func normalizeDownloadResourceContentType(value string) string {
	value = clampDownloadResourceValue(value, 128)
	if value == "" || !strings.Contains(value, "/") {
		return "application/octet-stream"
	}
	return value
}

func clampDownloadResourceValue(value string, max int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > max {
		return string(runes[:max])
	}
	return value
}

func firstDownloadResourceNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func downloadResourceScanTargets(item *DownloadResource) []any {
	return []any{
		&item.ID, &item.Slug, &item.NameZh, &item.NameEn, &item.DescriptionZh, &item.DescriptionEn,
		&item.Version, &item.Platform, &item.ObjectKey, &item.FileName, &item.ContentType, &item.SizeBytes,
		&item.ChecksumSHA256, &item.Published, &item.SortOrder, &item.DownloadCount, &item.UploadedAt,
		&item.CreatedAt, &item.UpdatedAt,
	}
}
