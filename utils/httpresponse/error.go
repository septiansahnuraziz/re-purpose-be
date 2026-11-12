package httpresponse

import (
	"errors"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/repurpose-ai/re-purpose-be/config"
	"github.com/repurpose-ai/re-purpose-be/utils"
)

const (
	ErrorLangEN = "EN"
	ErrorLangID = "ID"

	ErrorCodeNotRegister = "ERROR_CODE_NOT_REGISTERED"
)

type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type HTTPError struct {
	Code      int
	Message   string
	ErrorCode string
	Data      any `json:"data,omitempty"`
}

func NewHTTPError() *HTTPError { return new(HTTPError) }

func (e *HTTPError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *HTTPError) StatusCode() int {
	if e == nil {
		return http.StatusInternalServerError
	}
	if e.Code == 0 {
		return http.StatusInternalServerError
	}
	return e.Code
}

func (e *HTTPError) WithCode(code int) *HTTPError {
	e.Code = code
	return e
}

func (e *HTTPError) WithMessage(message string) *HTTPError {
	e.Message = message
	return e
}

func (e *HTTPError) WithErrorCode(errorCode string) *HTTPError {
	e.ErrorCode = errorCode
	return e
}

func (e *HTTPError) WithData(data any) *HTTPError {
	e.Data = data
	return e
}

type WrapperErrorResponseDTO struct {
	WrapperResponseDTO
	ErrorCode string `json:"errorCode"`
}

// Error writes the standard error envelope and aborts the handler chain.
func Error(c *gin.Context, httpError *HTTPError) {
	if httpError == nil {
		return
	}

	status := httpError.StatusCode()
	traceID := utils.GetTraceID(c)
	envelope := WrapperResponseDTO{
		ID:      traceID,
		TraceID: traceID,
		AppName: config.AppName(),
		Version: config.AppVersion(),
		Build:   config.AppBuild(),
		Success: false,
		HttpResponse: HttpResponse{
			Message: httpError.Message,
			Data:    httpError.Data,
		},
	}

	c.AbortWithStatusJSON(status, WrapperErrorResponseDTO{
		WrapperResponseDTO: envelope,
		ErrorCode:          httpError.ErrorCode,
	})
}

// Error catalog. Codes follow the go-partner convention: <PREFIX>-<http status><serial>.
var (
	ErrorBadRequest          = HTTPError{Code: http.StatusBadRequest, Message: "bad request", ErrorCode: "RPUR-400000"}
	ErrorInvalidRequest      = HTTPError{Code: http.StatusBadRequest, Message: "invalid request payload", ErrorCode: "RPUR-400001"}
	ErrorValidationFailed    = HTTPError{Code: http.StatusBadRequest, Message: "validation failed", ErrorCode: "RPUR-400002"}
	ErrorContentTooShort     = HTTPError{Code: http.StatusBadRequest, Message: "content is too short to repurpose", ErrorCode: "RPUR-400003"}
	ErrorInvalidURL          = HTTPError{Code: http.StatusBadRequest, Message: "invalid or unsupported url", ErrorCode: "RPUR-400004"}
	ErrorInvalidTone         = HTTPError{Code: http.StatusBadRequest, Message: "invalid tone", ErrorCode: "RPUR-400005"}
	ErrorInvalidSponsorEvent = HTTPError{Code: http.StatusBadRequest, Message: "invalid sponsor event", ErrorCode: "RPUR-400006"}
	ErrorInvalidPagination   = HTTPError{Code: http.StatusBadRequest, Message: "invalid pagination parameter", ErrorCode: "RPUR-400007"}
	ErrorUnauthorized        = HTTPError{Code: http.StatusUnauthorized, Message: "unauthorized", ErrorCode: "RPUR-401000"}
	ErrorInvalidToken        = HTTPError{Code: http.StatusUnauthorized, Message: "invalid or expired token", ErrorCode: "RPUR-401001"}
	ErrorForbidden           = HTTPError{Code: http.StatusForbidden, Message: "forbidden", ErrorCode: "RPUR-403000"}
	ErrorNotFound            = HTTPError{Code: http.StatusNotFound, Message: "data not found", ErrorCode: "RPUR-404000"}
	ErrorGenerationNotFound  = HTTPError{Code: http.StatusNotFound, Message: "generation is not found", ErrorCode: "RPUR-404001"}
	ErrorUserNotFound        = HTTPError{Code: http.StatusNotFound, Message: "user is not found", ErrorCode: "RPUR-404002"}
	ErrorInternalServer      = HTTPError{Code: http.StatusInternalServerError, Message: "internal server error", ErrorCode: "RPUR-500000"}
	ErrorNotImplemented      = HTTPError{Code: http.StatusNotImplemented, Message: "not implemented", ErrorCode: "RPUR-501000"}
	ErrorURLExtractionFailed = HTTPError{Code: http.StatusBadGateway, Message: "failed to extract content from url", ErrorCode: "RPUR-502001"}
	ErrorLLMUnavailable      = HTTPError{Code: http.StatusBadGateway, Message: "generation engine is unavailable", ErrorCode: "RPUR-502002"}
)

var localizedMessages = map[string]map[string]string{
	ErrorLangEN: {
		"RPUR-400000": "bad request",
		"RPUR-400001": "invalid request payload",
		"RPUR-400002": "validation failed",
		"RPUR-400003": "content is too short to repurpose",
		"RPUR-400004": "invalid or unsupported url",
		"RPUR-400005": "invalid tone",
		"RPUR-400006": "invalid sponsor event",
		"RPUR-400007": "invalid pagination parameter",
		"RPUR-401000": "unauthorized",
		"RPUR-401001": "invalid or expired token",
		"RPUR-403000": "forbidden",
		"RPUR-404000": "data not found",
		"RPUR-404001": "generation is not found",
		"RPUR-404002": "user is not found",
		"RPUR-500000": "internal server error",
		"RPUR-501000": "not implemented",
		"RPUR-502001": "failed to extract content from url",
		"RPUR-502002": "generation engine is unavailable",
	},
	ErrorLangID: {
		"RPUR-400000": "Permintaan tidak valid.",
		"RPUR-400001": "Format permintaan tidak valid.",
		"RPUR-400002": "Validasi permintaan gagal.",
		"RPUR-400003": "Teks terlalu pendek untuk diproses.",
		"RPUR-400004": "URL tidak valid atau tidak didukung.",
		"RPUR-400005": "Gaya nada tidak valid.",
		"RPUR-400006": "Event sponsor tidak valid.",
		"RPUR-400007": "Parameter pagination tidak valid.",
		"RPUR-401000": "Tidak terotorisasi.",
		"RPUR-401001": "Token tidak valid atau sudah kedaluwarsa.",
		"RPUR-403000": "Akses ditolak.",
		"RPUR-404000": "Data tidak ditemukan.",
		"RPUR-404001": "Riwayat generasi tidak ditemukan.",
		"RPUR-404002": "Pengguna tidak ditemukan.",
		"RPUR-500000": "Terjadi kesalahan pada server.",
		"RPUR-501000": "Belum diimplementasikan.",
		"RPUR-502001": "Gagal mengekstrak konten dari URL.",
		"RPUR-502002": "Mesin generasi sedang tidak tersedia.",
	},
}

func GetErrorMessage(lang, code string) error {
	if messages, ok := localizedMessages[lang]; ok {
		if message, ok := messages[code]; ok {
			return errors.New(message)
		}
	}
	return errors.New(ErrorCodeNotRegister)
}

// GetErrorList returns the error catalog sorted by error code for stable output.
func GetErrorList(lang string) []ErrorMessage {
	messages, ok := localizedMessages[lang]
	if !ok {
		messages = localizedMessages[ErrorLangEN]
	}

	codes := make([]string, 0, len(messages))
	for code := range messages {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	result := make([]ErrorMessage, 0, len(codes))
	for _, code := range codes {
		result = append(result, ErrorMessage{Code: code, Message: messages[code]})
	}
	return result
}
