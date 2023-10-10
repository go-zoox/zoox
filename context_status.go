package zoox

import (
	"net/http"
)

// StatusOK sets the response status code to 200.
func (ctx *Context) StatusOK() {
	ctx.Status(http.StatusOK)
}

// StatusCreated sets the response status code to 201.
func (ctx *Context) StatusCreated() {
	ctx.Status(http.StatusCreated)
}

// StatusAccepted sets the response status code to 202.
func (ctx *Context) StatusAccepted() {
	ctx.Status(http.StatusAccepted)
}

// StatusNonAuthoritativeInfo sets the response status code to 203.
func (ctx *Context) StatusNonAuthoritativeInfo() {
	ctx.Status(http.StatusNonAuthoritativeInfo)
}

// StatusNoContent sets the response status code to 204.
func (ctx *Context) StatusNoContent() {
	ctx.Status(http.StatusNoContent)
}

// StatusResetContent sets the response status code to 205.
func (ctx *Context) StatusResetContent() {
	ctx.Status(http.StatusResetContent)
}

// StatusPartialContent sets the response status code to 206.
func (ctx *Context) StatusPartialContent() {
	ctx.Status(http.StatusPartialContent)
}

// StatusMultiStatus sets the response status code to 207.
func (ctx *Context) StatusMultiStatus() {
	ctx.Status(http.StatusMultiStatus)
}

// StatusAlreadyReported sets the response status code to 208.
func (ctx *Context) StatusAlreadyReported() {
	ctx.Status(http.StatusAlreadyReported)
}

// StatusIMUsed sets the response status code to 226.
func (ctx *Context) StatusIMUsed() {
	ctx.Status(http.StatusIMUsed)
}

// StatusMultipleChoices sets the response status code to 300.
func (ctx *Context) StatusMultipleChoices() {
	ctx.Status(http.StatusMultipleChoices)
}

// StatusMovedPermanently sets the response status code to 301.
func (ctx *Context) StatusMovedPermanently() {
	ctx.Status(http.StatusMovedPermanently)
}

// StatusFound sets the response status code to 302.
func (ctx *Context) StatusFound() {
	ctx.Status(http.StatusFound)
}

// StatusSeeOther sets the response status code to 303.
func (ctx *Context) StatusSeeOther() {
	ctx.Status(http.StatusSeeOther)
}

// StatusNotModified sets the response status code to 304.
func (ctx *Context) StatusNotModified() {
	ctx.Status(http.StatusNotModified)
}

// StatusUseProxy sets the response status code to 305.
func (ctx *Context) StatusUseProxy() {
	ctx.Status(http.StatusUseProxy)
}

// StatusTemporaryRedirect sets the response status code to 307.
func (ctx *Context) StatusTemporaryRedirect() {
	ctx.Status(http.StatusTemporaryRedirect)
}

// StatusPermanentRedirect sets the response status code to 308.
func (ctx *Context) StatusPermanentRedirect() {
	ctx.Status(http.StatusPermanentRedirect)
}

// StatusBadRequest sets the response status code to 400.
func (ctx *Context) StatusBadRequest() {
	ctx.Status(http.StatusBadRequest)
}

// StatusUnauthorized sets the response status code to 401.
func (ctx *Context) StatusUnauthorized() {
	ctx.Status(http.StatusUnauthorized)
}

// StatusPaymentRequired sets the response status code to 402.
func (ctx *Context) StatusPaymentRequired() {
	ctx.Status(http.StatusPaymentRequired)
}

// StatusForbidden sets the response status code to 403.
func (ctx *Context) StatusForbidden() {
	ctx.Status(http.StatusForbidden)
}

// StatusNotFound sets the response status code to 404.
func (ctx *Context) StatusNotFound() {
	ctx.Status(http.StatusNotFound)
}

// StatusMethodNotAllowed sets the response status code to 405.
func (ctx *Context) StatusMethodNotAllowed() {
	ctx.Status(http.StatusMethodNotAllowed)
}

// StatusNotAcceptable sets the response status code to 406.
func (ctx *Context) StatusNotAcceptable() {
	ctx.Status(http.StatusNotAcceptable)
}

// StatusProxyAuthRequired sets the response status code to 407.
func (ctx *Context) StatusProxyAuthRequired() {
	ctx.Status(http.StatusProxyAuthRequired)
}

// StatusRequestTimeout sets the response status code to 408.
func (ctx *Context) StatusRequestTimeout() {
	ctx.Status(http.StatusRequestTimeout)
}

// StatusConflict sets the response status code to 409.
func (ctx *Context) StatusConflict() {
	ctx.Status(http.StatusConflict)
}

// StatusGone sets the response status code to 410.
func (ctx *Context) StatusGone() {
	ctx.Status(http.StatusGone)
}

// StatusLengthRequired sets the response status code to 411.
func (ctx *Context) StatusLengthRequired() {
	ctx.Status(http.StatusLengthRequired)
}

// StatusPreconditionFailed sets the response status code to 412.
func (ctx *Context) StatusPreconditionFailed() {
	ctx.Status(http.StatusPreconditionFailed)
}

// StatusRequestEntityTooLarge sets the response status code to 413.
func (ctx *Context) StatusRequestEntityTooLarge() {
	ctx.Status(http.StatusRequestEntityTooLarge)
}

// StatusRequestURITooLong sets the response status code to 414.
func (ctx *Context) StatusRequestURITooLong() {
	ctx.Status(http.StatusRequestURITooLong)
}

// StatusUnsupportedMediaType sets the response status code to 415.
func (ctx *Context) StatusUnsupportedMediaType() {
	ctx.Status(http.StatusUnsupportedMediaType)
}

// StatusRequestedRangeNotSatisfiable sets the response status code to 416.
func (ctx *Context) StatusRequestedRangeNotSatisfiable() {
	ctx.Status(http.StatusRequestedRangeNotSatisfiable)
}

// StatusExpectationFailed sets the response status code to 417.
func (ctx *Context) StatusExpectationFailed() {
	ctx.Status(http.StatusExpectationFailed)
}

// StatusTeapot sets the response status code to 418.
func (ctx *Context) StatusTeapot() {
	ctx.Status(http.StatusTeapot)
}

// StatusMisdirectedRequest sets the response status code to 421.
func (ctx *Context) StatusMisdirectedRequest() {
	ctx.Status(http.StatusMisdirectedRequest)
}

// StatusUnprocessableEntity sets the response status code to 422.
func (ctx *Context) StatusUnprocessableEntity() {
	ctx.Status(http.StatusUnprocessableEntity)
}

// StatusLocked sets the response status code to 423.
func (ctx *Context) StatusLocked() {
	ctx.Status(http.StatusLocked)
}

// StatusFailedDependency sets the response status code to 424.
func (ctx *Context) StatusFailedDependency() {
	ctx.Status(http.StatusFailedDependency)
}

// StatusTooEarly sets the response status code to 425.
func (ctx *Context) StatusTooEarly() {
	ctx.Status(http.StatusTooEarly)
}

// StatusUpgradeRequired sets the response status code to 426.
func (ctx *Context) StatusUpgradeRequired() {
	ctx.Status(http.StatusUpgradeRequired)
}

// StatusPreconditionRequired sets the response status code to 428.
func (ctx *Context) StatusPreconditionRequired() {
	ctx.Status(http.StatusPreconditionRequired)
}

// StatusTooManyRequests sets the response status code to 429.
func (ctx *Context) StatusTooManyRequests() {
	ctx.Status(http.StatusTooManyRequests)
}

// StatusRequestHeaderFieldsTooLarge sets the response status code to 431.
func (ctx *Context) StatusRequestHeaderFieldsTooLarge() {
	ctx.Status(http.StatusRequestHeaderFieldsTooLarge)
}

// StatusUnavailableForLegalReasons sets the response status code to 451.
func (ctx *Context) StatusUnavailableForLegalReasons() {
	ctx.Status(http.StatusUnavailableForLegalReasons)
}

// StatusInternalServerError sets the response status code to 500.
func (ctx *Context) StatusInternalServerError() {
	ctx.Status(http.StatusInternalServerError)
}

// StatusNotImplemented sets the response status code to 501.
func (ctx *Context) StatusNotImplemented() {
	ctx.Status(http.StatusNotImplemented)
}

// StatusBadGateway sets the response status code to 502.
func (ctx *Context) StatusBadGateway() {
	ctx.Status(http.StatusBadGateway)
}

// StatusServiceUnavailable sets the response status code to 503.
func (ctx *Context) StatusServiceUnavailable() {
	ctx.Status(http.StatusServiceUnavailable)
}

// StatusGatewayTimeout sets the response status code to 504.
func (ctx *Context) StatusGatewayTimeout() {
	ctx.Status(http.StatusGatewayTimeout)
}

// StatusHTTPVersionNotSupported sets the response status code to 505.
func (ctx *Context) StatusHTTPVersionNotSupported() {
	ctx.Status(http.StatusHTTPVersionNotSupported)
}

// StatusVariantAlsoNegotiates sets the response status code to 506.
func (ctx *Context) StatusVariantAlsoNegotiates() {
	ctx.Status(http.StatusVariantAlsoNegotiates)
}

// StatusInsufficientStorage sets the response status code to 507.
func (ctx *Context) StatusInsufficientStorage() {
	ctx.Status(http.StatusInsufficientStorage)
}

// StatusLoopDetected sets the response status code to 508.
func (ctx *Context) StatusLoopDetected() {
	ctx.Status(http.StatusLoopDetected)
}

// StatusNotExtended sets the response status code to 510.
func (ctx *Context) StatusNotExtended() {
	ctx.Status(http.StatusNotExtended)
}

// StatusNetworkAuthenticationRequired sets the response status code to 511.
func (ctx *Context) StatusNetworkAuthenticationRequired() {
	ctx.Status(http.StatusNetworkAuthenticationRequired)
}
