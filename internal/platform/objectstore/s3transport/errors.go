package s3transport

import (
	"errors"
	"github.com/aws/smithy-go"
)

func ErrorCode(err error) string {
	var apiError smithy.APIError
	if errors.As(err, &apiError) {
		return apiError.ErrorCode()
	}
	return ""
}
