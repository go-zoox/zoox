package zoox

import (
	"fmt"
	"time"

	"github.com/go-zoox/headers"
)

// SetContentType sets the request content-type header.
func (ctx *Context) SetContentType(contentType string) {
	ctx.SetHeader(headers.ContentType, contentType)
}

// SetCacheControl sets the request cache-control header.
func (ctx *Context) SetCacheControl(cacheControl string) {
	ctx.SetHeader(headers.CacheControl, cacheControl)
}

// SetCacheControlWithMaxAge sets the request cache-control header with max-age.
func (ctx *Context) SetCacheControlWithMaxAge(maxAge time.Duration) {
	ctx.SetCacheControl(fmt.Sprintf("max-age=%d", int(maxAge.Seconds())))
}

// SetCacheControlWithNoCache sets the request cache-control header with no-cache.
func (ctx *Context) SetCacheControlWithNoCache() {
	ctx.SetCacheControl("no-cache")
}

// SetCacheControlWithNoStore sets the request cache-control header with no-store.
func (ctx *Context) SetCacheControlWithNoStore() {
	ctx.SetCacheControl("no-store")
}

// SetHSTS sets the request strict-transport-security header.
func (ctx *Context) SetHSTS(maxAge time.Duration) {
	ctx.SetHeader(headers.StrictTransportSecurity, fmt.Sprintf("max-age=%d", int(maxAge.Seconds())))
}

// SetXFramOptions sets the request x-frame-options header.
func (ctx *Context) SetXFramOptions(value string) {
	ctx.SetHeader(headers.XFrameOptions, value)
}

// SetXFramOptionsDeny sets the request x-frame-options header with deny.
func (ctx *Context) SetXFramOptionsDeny() {
	ctx.SetXFramOptions("deny")
}

// SetXFramOptionsSameOrigin sets the request x-frame-options header with sameorigin.
func (ctx *Context) SetXFramOptionsSameOrigin() {
	ctx.SetXFramOptions("sameorigin")
}

// SetXFramOptionsAllowFrom sets the request x-frame-options header with allow-from.
func (ctx *Context) SetXFramOptionsAllowFrom(uri string) {
	ctx.SetXFramOptions(fmt.Sprintf("allow-from %s", uri))
}

// SetXDownloadOptions sets the request x-download-options header.
func (ctx *Context) SetXDownloadOptions(value string) {
	ctx.SetHeader(headers.XDownloadOptions, value)
}

// SetXDownloadOptionsNoOpen sets the request x-download-options header with noopen.
func (ctx *Context) SetXDownloadOptionsNoOpen() {
	ctx.SetXDownloadOptions("noopen")
}

// SetXDownloadOptionsNoSniff sets the request x-download-options header with nosniff.
func (ctx *Context) SetXDownloadOptionsNoSniff() {
	ctx.SetXDownloadOptions("nosniff")
}

// SetXXSSProtection sets the request x-xss-protection header.
func (ctx *Context) SetXXSSProtection(value string) {
	ctx.SetHeader(headers.XXSSProtection, value)
}

// SetXXSSProtectionEnable sets the request x-xss-protection header with 1; mode=block.
func (ctx *Context) SetXXSSProtectionEnable() {
	ctx.SetXXSSProtection("1; mode=block")
}

// SetXXSSProtectionDisable sets the request x-xss-protection header with 0.
func (ctx *Context) SetXXSSProtectionDisable() {
	ctx.SetXXSSProtection("0")
}

// SetXXSSProtectionReport sets the request x-xss-protection header with 1; report=<reporting-uri>.
func (ctx *Context) SetXXSSProtectionReport(reportingURI string) {
	ctx.SetXXSSProtection(fmt.Sprintf("1; report=%s", reportingURI))
}

// SetPoweredBy sets the request x-powered-by header.
func (ctx *Context) SetPoweredBy(value string) {
	ctx.SetHeader(headers.XPoweredBy, value)
}

// SetContentDisposition sets the request content-disposition header with attachment.
func (ctx *Context) SetContentDisposition(filename string) {
	ctx.SetHeader(headers.ContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
}

// SetContentDispositionInline sets the request content-disposition header with inline.
func (ctx *Context) SetContentDispositionInline(filename string) {
	ctx.SetHeader(headers.ContentDisposition, fmt.Sprintf("inline; filename=\"%s\"", filename))
}

// SetDownloadFilename sets the request content-disposition header with attachment and filename.
func (ctx *Context) SetDownloadFilename(filename string) {
	ctx.SetContentDisposition(filename)
}

// SetContentLocation sets the request content-location header.
func (ctx *Context) SetContentLocation(value string) {
	ctx.SetHeader(headers.ContentLocation, value)
}

// SetLocation sets the request location header.
func (ctx *Context) SetLocation(value string) {
	ctx.SetHeader(headers.Location, value)
}
