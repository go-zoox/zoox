package middleware

// inspired by https://github.com/gofiber/fiber helmet

import (
	"fmt"

	"github.com/go-zoox/headers"
	"github.com/go-zoox/zoox"
)

// HelmetConfig defines the helmet config
type HelmetConfig struct {
	// Next defines a function to skip this middleware when returned true.
	// Optional. Default: nil
	Next func(*zoox.Context) bool

	// XSSProtection adds some protection against cross site scripting (XSS) attacks
	// Optional. Default value "0"
	XSSProtection string

	// ContentTypeNosniff prevents the browser from doing MIME-type sniffing
	// Optional. Default value "nosniff"
	ContentTypeNosniff string

	// XFrameOptions can be used to indicate whether or not a browser should be allowed to render a page in a <frame>, <iframe> or <object> .
	// Optional. Default value "SAMEORIGIN"
	// Possible values: "SAMEORIGIN", "DENY", "ALLOW-FROM uri"
	XFrameOptions string

	// HSTSMaxAge sets the Strict-Transport-Security header to indicate how long (in seconds) browsers should remember that this site is only to be accessed using HTTPS
	// Optional. Default value 0
	HSTSMaxAge int

	// HSTSExcludeSubdomains excludes subdomains from the Strict-Transport-Security header
	HSTSExcludeSubdomains bool

	// ContentSecurityPolicy sets the Content-Security-Policy header to help prevent cross-site scripting attacks and other cross-site injections
	// Optional. Default value ""
	ContentSecurityPolicy string

	// CSPReportOnly sets the Content-Security-Policy-Report-Only header
	// Optional. Default value false
	CSPReportOnly bool

	// HSTSPreloadEnabled adds the preload directive to the Strict-Transport-Security header
	// Optional. Default value false
	HSTSPreloadEnabled bool

	// ReferrerPolicy sets the Referrer-Policy header to indicate which referrer information should be included with requests made
	// Optional. Default value "ReferrerPolicy"
	ReferrerPolicy string

	// Permissions-Policy sets the Permissions-Policy header to indicate which features are allowed to be used
	// Optional. Default value ""
	PermissionsPolicy string

	// Cross-Origin-Embedder-Policy sets the Cross-Origin-Embedder-Policy header to indicate whether a resource should be loaded as part of a document
	// Optional. Default value ""require-corp"
	CrossOriginEmbedderPolicy string

	// Cross-Origin-Opener-Policy sets the Cross-Origin-Opener-Policy header to indicate whether a resource should be opened in the same browsing context as the document
	// Optional. Default value "same-origin"
	CrossOriginOpenerPolicy string

	// Cross-Origin-Resource-Policy sets the Cross-Origin-Resource-Policy header to indicate whether a resource should be shared cross-origin
	// Optional. Default value "same-origin"
	CrossOriginResourcePolicy string

	// Origin-Agent-Cluster sets the Origin-Agent-Cluster header to indicate whether a resource should be shared cross-origin
	// Optional. Default value "?1"
	OriginAgentCluster string

	// X-DNS-Prefetch-Control sets the X-DNS-Prefetch-Control header to control DNS prefetching
	// Optional. Default value "off"
	XDNSPrefetchControl string

	// X-Download-Options sets the X-Download-Options header to prevent Internet Explorer from executing downloads in your site’s context
	// Optional. Default value "noopen"
	XDownloadOptions string

	// X-Permitted-Cross-Domain-Policies sets the X-Permitted-Cross-Domain-Policies header to indicate whether a resource should be shared cross-origin
	// Optional. Default value "none"
	XPermittedCrossDomainPolicies string
}

// DefaultHelmetConfig is the default helmet config
var DefaultHelmetConfig = &HelmetConfig{
	XSSProtection:                 "0",
	ContentTypeNosniff:            "nosniff",
	XFrameOptions:                 "SAMEORIGIN",
	HSTSMaxAge:                    0,
	HSTSExcludeSubdomains:         false,
	ContentSecurityPolicy:         "",
	CSPReportOnly:                 false,
	HSTSPreloadEnabled:            false,
	ReferrerPolicy:                "ReferrerPolicy",
	PermissionsPolicy:             "",
	CrossOriginEmbedderPolicy:     "require-corp",
	CrossOriginOpenerPolicy:       "same-origin",
	CrossOriginResourcePolicy:     "same-origin",
	OriginAgentCluster:            "?1",
	XDNSPrefetchControl:           "off",
	XDownloadOptions:              "noopen",
	XPermittedCrossDomainPolicies: "none",
}

func defaultHelmetConfig(cfg *HelmetConfig) *HelmetConfig {
	if cfg == nil {
		return DefaultHelmetConfig
	}

	if cfg.XSSProtection == "" {
		cfg.XSSProtection = DefaultHelmetConfig.XSSProtection
	}

	if cfg.ContentTypeNosniff == "" {
		cfg.ContentTypeNosniff = DefaultHelmetConfig.ContentTypeNosniff
	}

	if cfg.XFrameOptions == "" {
		cfg.XFrameOptions = DefaultHelmetConfig.XFrameOptions
	}

	if cfg.HSTSMaxAge == 0 {
		cfg.HSTSMaxAge = DefaultHelmetConfig.HSTSMaxAge
	}

	if cfg.HSTSExcludeSubdomains == false {
		cfg.HSTSExcludeSubdomains = DefaultHelmetConfig.HSTSExcludeSubdomains
	}

	if cfg.ContentSecurityPolicy == "" {
		cfg.ContentSecurityPolicy = DefaultHelmetConfig.ContentSecurityPolicy
	}

	if cfg.CSPReportOnly == false {
		cfg.CSPReportOnly = DefaultHelmetConfig.CSPReportOnly
	}

	if cfg.HSTSPreloadEnabled == false {
		cfg.HSTSPreloadEnabled = DefaultHelmetConfig.HSTSPreloadEnabled
	}

	if cfg.ReferrerPolicy == "" {
		cfg.ReferrerPolicy = DefaultHelmetConfig.ReferrerPolicy
	}

	if cfg.PermissionsPolicy == "" {
		cfg.PermissionsPolicy = DefaultHelmetConfig.PermissionsPolicy
	}

	if cfg.CrossOriginEmbedderPolicy == "" {
		cfg.CrossOriginEmbedderPolicy = DefaultHelmetConfig.CrossOriginEmbedderPolicy
	}

	if cfg.CrossOriginOpenerPolicy == "" {
		cfg.CrossOriginOpenerPolicy = DefaultHelmetConfig.CrossOriginOpenerPolicy
	}

	if cfg.CrossOriginResourcePolicy == "" {
		cfg.CrossOriginResourcePolicy = DefaultHelmetConfig.CrossOriginResourcePolicy
	}

	if cfg.OriginAgentCluster == "" {
		cfg.OriginAgentCluster = DefaultHelmetConfig.OriginAgentCluster
	}

	if cfg.XDNSPrefetchControl == "" {
		cfg.XDNSPrefetchControl = DefaultHelmetConfig.XDNSPrefetchControl
	}

	if cfg.XDownloadOptions == "" {
		cfg.XDownloadOptions = DefaultHelmetConfig.XDownloadOptions
	}

	if cfg.XPermittedCrossDomainPolicies == "" {
		cfg.XPermittedCrossDomainPolicies = DefaultHelmetConfig.XPermittedCrossDomainPolicies
	}

	return cfg
}

// Helmet is a middleware that adds some security response headers.
func Helmet(cfg *HelmetConfig) zoox.Middleware {
	cfgX := defaultHelmetConfig(cfg)
	return func(ctx *zoox.Context) {
		if cfgX.XSSProtection != "" {
			ctx.SetHeader(headers.XXSSProtection, cfgX.XSSProtection)
		}

		if cfgX.ContentTypeNosniff != "" {
			ctx.SetHeader(headers.XContentTypeOptions, cfgX.ContentTypeNosniff)
		}

		if cfgX.XFrameOptions != "" {
			ctx.SetHeader(headers.XFrameOptions, cfgX.XFrameOptions)
		}

		if ctx.Protocol() == "https" {
			subdomains := ""
			if !cfgX.HSTSExcludeSubdomains {
				subdomains = "; includeSubDomains"
			}
			if cfgX.HSTSPreloadEnabled {
				subdomains = fmt.Sprintf("%s; preload", subdomains)
			}

			ctx.SetHeader(headers.StrictTransportSecurity, fmt.Sprintf("max-age=%d%s", cfgX.HSTSMaxAge, subdomains))
		}

		if cfgX.ContentSecurityPolicy != "" {
			if cfgX.CSPReportOnly {
				ctx.SetHeader(headers.ContentSecurityPolicyReportOnly, cfgX.ContentSecurityPolicy)
			} else {
				ctx.SetHeader(headers.ContentSecurityPolicy, cfgX.ContentSecurityPolicy)
			}
		}

		if cfgX.ReferrerPolicy != "" {
			ctx.SetHeader(headers.ReferrerPolicy, cfgX.ReferrerPolicy)
		}

		if cfgX.PermissionsPolicy != "" {
			ctx.SetHeader(headers.PermissionsPolicy, cfgX.PermissionsPolicy)
		}

		if cfgX.CrossOriginEmbedderPolicy != "" {
			ctx.SetHeader(headers.CrossOriginEmbedderPolicy, cfgX.CrossOriginEmbedderPolicy)
		}

		if cfgX.CrossOriginOpenerPolicy != "" {
			ctx.SetHeader(headers.CrossOriginOpenerPolicy, cfgX.CrossOriginOpenerPolicy)
		}

		if cfgX.CrossOriginResourcePolicy != "" {
			ctx.SetHeader(headers.CrossOriginResourcePolicy, cfgX.CrossOriginResourcePolicy)
		}

		if cfgX.OriginAgentCluster != "" {
			ctx.SetHeader(headers.OriginAgentCluster, cfgX.OriginAgentCluster)
		}

		if cfgX.XDNSPrefetchControl != "" {
			ctx.SetHeader(headers.XDNSPrefetchControl, cfgX.XDNSPrefetchControl)
		}

		if cfgX.XDownloadOptions != "" {
			ctx.SetHeader(headers.XDownloadOptions, cfgX.XDownloadOptions)
		}

		if cfgX.XPermittedCrossDomainPolicies != "" {
			ctx.SetHeader(headers.XPermittedCrossDomainPolicies, cfgX.XPermittedCrossDomainPolicies)
		}

		ctx.Next()
	}
}
