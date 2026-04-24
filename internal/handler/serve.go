package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/nicovak/miqio-go/logger"
	"github.com/nicovak/miqio-proxy/internal/cache"
	"github.com/nicovak/miqio-proxy/internal/r2"
	"go.uber.org/zap"
)

type ServeHandler struct {
	cacheStore      *cache.Store
	r2Client        *r2.Client
	miqioPageDomain string
}

func NewServeHandler(store *cache.Store, r2Client *r2.Client, miqioPageDomain string) *ServeHandler {
	return &ServeHandler{
		cacheStore:      store,
		r2Client:        r2Client,
		miqioPageDomain: miqioPageDomain,
	}
}

func (h *ServeHandler) Handle(c *echo.Context) error {
	host := c.Request().Host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	reqPath := c.Request().URL.Path
	rawQuery := c.Request().URL.RawQuery

	if redirect := h.normalizeURL(reqPath, rawQuery); redirect != "" {
		return c.Redirect(http.StatusMovedPermanently, redirect)
	}

	tenantID, err := h.resolveTenant(host)
	if err != nil {
		return c.HTML(http.StatusNotFound, notFoundSitePage)
	}

	slug := strings.TrimPrefix(reqPath, "/")

	if entry, ok := h.cacheStore.Redirects.Lookup(tenantID, slug); ok {
		targetURL := mergeQueryParams(entry.TargetURL, rawQuery)
		cacheHeader := "no-cache"
		if entry.StatusCode == http.StatusMovedPermanently {
			cacheHeader = "public, max-age=3600"
		}
		c.Response().Header().Set("Cache-Control", cacheHeader)
		return c.Redirect(entry.StatusCode, targetURL)
	}

	r2Key := h.buildR2Key(tenantID, slug)
	obj, err := h.r2Client.GetObject(c.Request().Context(), r2Key)
	if err != nil {
		if r2.IsNotFound(err) {
			return c.HTML(http.StatusNotFound, notFoundPagePage)
		}
		logger.Get().Errorw("r2 fetch failed", zap.String("key", r2Key), zap.Error(err))
		return c.String(http.StatusBadGateway, "upstream error")
	}
	defer obj.Body.Close()

	body, err := io.ReadAll(obj.Body)
	if err != nil {
		logger.Get().Errorw("r2 read failed", zap.String("key", r2Key), zap.Error(err))
		return c.String(http.StatusBadGateway, "upstream error")
	}

	c.Response().Header().Set("Cache-Control", "public, max-age=300")
	if obj.ETag != "" {
		c.Response().Header().Set("ETag", obj.ETag)
	}

	return c.HTMLBlob(http.StatusOK, body)
}

func (h *ServeHandler) resolveTenant(host string) (string, error) {
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

func (h *ServeHandler) normalizeURL(path, rawQuery string) string {
	if path != "/" && strings.HasSuffix(path, "/") {
		target := strings.TrimSuffix(path, "/")
		return appendQuery(target, rawQuery)
	}

	if strings.HasSuffix(path, ".html") {
		target := strings.TrimSuffix(path, ".html")
		if strings.HasSuffix(target, "/index") {
			target = strings.TrimSuffix(target, "/index")
			if target == "" {
				target = "/"
			}
		}
		return appendQuery(target, rawQuery)
	}

	if strings.HasSuffix(path, "/index") {
		target := strings.TrimSuffix(path, "/index")
		if target == "" {
			target = "/"
		}
		return appendQuery(target, rawQuery)
	}

	return ""
}

func (h *ServeHandler) buildR2Key(tenantUUID, slug string) string {
	if slug == "" {
		return fmt.Sprintf("%s/index.html", tenantUUID)
	}
	return fmt.Sprintf("%s/%s/index.html", tenantUUID, slug)
}

func appendQuery(path, rawQuery string) string {
	if rawQuery != "" {
		return path + "?" + rawQuery
	}
	return path
}

func mergeQueryParams(targetURL, rawQuery string) string {
	if rawQuery == "" {
		return targetURL
	}

	parsed, err := url.Parse(targetURL)
	if err != nil {
		return targetURL
	}

	incoming, _ := url.ParseQuery(rawQuery)
	existing := parsed.Query()

	for k, v := range incoming {
		existing[k] = v
	}

	parsed.RawQuery = existing.Encode()
	return parsed.String()
}

const notFoundSitePage = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>Site introuvable</title>
<style>body{font-family:system-ui,sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#f8f9fa;color:#333}
.c{text-align:center}h1{font-size:2rem;margin-bottom:.5rem}p{color:#666}</style></head>
<body><div class="c"><h1>404</h1><p>Site introuvable</p></div></body></html>`

const notFoundPagePage = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>Page introuvable</title>
<style>body{font-family:system-ui,sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#f8f9fa;color:#333}
.c{text-align:center}h1{font-size:2rem;margin-bottom:.5rem}p{color:#666}</style></head>
<body><div class="c"><h1>404</h1><p>Page introuvable</p></div></body></html>`
