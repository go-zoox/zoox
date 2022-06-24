package zoox

// defaultGlobalMiddleware is the default global middleware
var defaultGlobalMiddleware = []HandlerFunc{
	Logger(),
}
