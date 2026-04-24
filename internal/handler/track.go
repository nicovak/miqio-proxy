package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/nicovak/miqio-go/logger"
	"github.com/nicovak/miqio-proxy/internal/cache"
	"go.uber.org/zap"
)

const (
	trackMaxBodySize    = 16 * 1024 // 16 KB
	trackUpstreamTimout = 3 * time.Second
)

type TrackHandler struct {
	cacheStore      *cache.Store
	miqioPageDomain string
	appOrigin       string
	httpClient      *http.Client
}

func NewTrackHandler(store *cache.Store, miqioPageDomain, appOrigin string) *TrackHandler {
	return &TrackHandler{
		cacheStore:      store,
		miqioPageDomain: miqioPageDomain,
		appOrigin:       strings.TrimRight(appOrigin, "/"),
		httpClient: &http.Client{
			Timeout: trackUpstreamTimout,
		},
	}
}

func (h *TrackHandler) Handle(c *echo.Context) error {
	if c.Request().Method != http.MethodPost {
		return c.String(http.StatusMethodNotAllowed, "Method Not Allowed")
	}

	host := c.Request().Host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	if _, err := h.resolveTenant(host); err != nil {
		return c.String(http.StatusNotFound, "Site introuvable")
	}

	if !h.validateOrigin(c, host) {
		return c.String(http.StatusForbidden, "Origin mismatch")
	}

	body := http.MaxBytesReader(nil, c.Request().Body, trackMaxBodySize)
	defer body.Close()

	payload, err := io.ReadAll(body)
	if err != nil {
		return c.String(http.StatusRequestEntityTooLarge, "Payload too large")
	}

	upstreamURL := h.appOrigin + "/api/track"
	req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodPost, upstreamURL, strings.NewReader(string(payload)))
	if err != nil {
		logger.Get().Errorw("track: failed to build upstream request", zap.Error(err))
		return c.String(http.StatusBadGateway, "upstream error")
	}

	req.Header.Set("Content-Type", c.Request().Header.Get("Content-Type"))
	req.Header.Set("X-Forwarded-For", c.RealIP())
	req.Header.Set("X-Real-IP", c.RealIP())

	resp, err := h.httpClient.Do(req)
	if err != nil {
		logger.Get().Errorw("track: upstream request failed", zap.Error(err))
		return c.String(http.StatusBadGateway, "upstream error")
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	c.Response().Header().Set("Cache-Control", "no-store")
	return c.Blob(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

func (h *TrackHandler) resolveTenant(host string) (string, error) {
	if strings.HasSuffix(host, "."+h.miqioPageDomain) {
		slug := strings.TrimSuffix(host, "."+h.miqioPageDomain)
		if id, ok := h.cacheStore.Tenants.Resolve(slug); ok {
			return id, nil
		}
		return "", fmt.Errorf("tenant not found for slug: %s", slug)
	}

	if id, ok := h.cacheStore.Domains.Resolve(host); ok {
		return id, nil
	}

	return "", fmt.Errorf("unknown host: %s", host)
}

func (h *TrackHandler) validateOrigin(c *echo.Context, expectedHost string) bool {
	if origin := c.Request().Header.Get("Origin"); origin != "" {
		return matchesHost(origin, expectedHost)
	}

	if referer := c.Request().Header.Get("Referer"); referer != "" {
		return matchesHost(referer, expectedHost)
	}

	return false
}

func matchesHost(rawURL, expectedHost string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	hostname := parsed.Hostname()
	return strings.EqualFold(hostname, expectedHost)
}
