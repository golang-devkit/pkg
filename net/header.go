package net

const (
	corsAllowOriginHeader      string = "Access-Control-Allow-Origin"
	corsExposeHeadersHeader    string = "Access-Control-Expose-Headers"
	corsMaxAgeHeader           string = "Access-Control-Max-Age"
	corsAllowMethodsHeader     string = "Access-Control-Allow-Methods"
	corsAllowHeadersHeader     string = "Access-Control-Allow-Headers"
	corsAllowCredentialsHeader string = "Access-Control-Allow-Credentials"
	corsRequestMethodHeader    string = "Access-Control-Request-Method"
	corsRequestHeadersHeader   string = "Access-Control-Request-Headers"
	corsOriginHeader           string = "Origin"
	corsVaryHeader             string = "Vary"

	headerOrigin        string = "Origin"
	headerUserAgent     string = "User-Agent"
	headerContentType   string = "Content-Type"
	headerAuthorization string = "Authorization"
	headerContentLength string = "Content-Length"

	// Custom request headers
	xApiClientId       string = "X-Api-Client-Id"
	xApiRequestId      string = "X-Api-Request-Id"
	xApiServiceAccount string = "X-Api-Service-Account"

	// Custom response headers
	xDescription      string = "X-Description"
	xDescriptionError string = "X-Description-Error"
)
