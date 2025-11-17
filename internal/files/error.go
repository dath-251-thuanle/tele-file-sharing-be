package files

import "fmt"

// ServiceError represents a service-level error with a code and message.
type ServiceError struct {
    Code    string
    Message string
}

func (e *ServiceError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Helper constructors for common service errors used by the files service.
func NewFileNotFoundError() error {
    return &ServiceError{Code: "file_not_found", Message: "file not found or access denied"}
}

func NewUploadReportNotFoundError() error {
    return &ServiceError{Code: "upload_report_not_found", Message: "upload report not found"}
}

func NewDatabaseError() error {
    return &ServiceError{Code: "database_error", Message: "internal database error"}
}
func InvalidPasswordError() error{
    return &ServiceError{Code: "ERR_INVALID_PASSWORD", Message: "the provided password is incorrect"}
}
