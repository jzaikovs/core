package core

const (
	ContentType_JSON = "application/json"
)

const (
	MIME_JSON = `application/json; charset=UTF-8`
	MIME_HTML = `text/html; charset=UTF-8`
)

const (
	Response_Ok                     = 200
	Response_Created                = 201
	Response_Bad_Request            = 400
	Response_Unauthorized           = 401
	Response_Forbidden              = 403
	Response_Not_Found              = 404
	Response_Unsupported_Media_Type = 415
	Response_Unprocessable_Entity   = 422
	Response_Too_Many_Requests      = 429
	Response_Internal_Server_Error  = 500
)

const (
	Header_X_Rate_Limit_Limit     = `X-Rate-Limit-Limit`
	Header_X_Rate_Limit_Remaining = `X-Rate-Limit-Remaining`
)
