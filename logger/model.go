package logger

const (

	// Log field keys for structured logging
	KeyServiceModule = "module"
	KeyFunctionName  = "function_name"
	KeyError         = "error"
	KeyEnvironment   = "environment"
	KeyTimestamp     = "timestamp"

	KeyJwtString = "jwt"

	// Network related log field keys
	KeyNetRemoteAddr       = "remote_addr"
	KeyNetHttpMethod       = "http_method"
	KeyNetHttpPath         = "http_path"
	KeyNetHttpQuery        = "http_query"
	KeyNetStatus           = "status"
	KeyNetStatusCode       = "status_code"
	KeyNetDuration         = "duration"
	KeyNetClientID         = "client_id"
	KeyNetRequestID        = "request_id"
	KeyNetRequestPayload   = "request"
	KeyNetRequestHeaders   = "headers"
	KeyNetOrigin           = "origin"
	KeyNetUserAgent        = "user_agent"
	KeyNetHostname         = "hostname"
	KeyNetDescription      = "application_description"
	KeyNetDescriptionError = "application_description_error"
	KeyNetResponsePayload  = "response"
	KeyNetResponseSize     = "response_size"
)

const (
	// Environment variable keys
	EnvDeploymentKey = "DEPLOYMENT_ENVIRONMENT"
)
