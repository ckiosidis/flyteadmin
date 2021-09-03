// Convenience methods for shared errors.
package shared

import (
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/errors"

	"google.golang.org/grpc/codes"
)

const missingFieldFormat = "missing %s"
const invalidArgFormat = "invalid value for %s"
const notSupportedFormat = "not supported for %s"

func GetMissingArgumentError(field string) error {
	return errors.NewFlyteAdminError(codes.InvalidArgument, fmt.Sprintf(missingFieldFormat, field))
}

func GetInvalidArgumentError(field string) error {
	return errors.NewFlyteAdminError(codes.InvalidArgument, fmt.Sprintf(invalidArgFormat, field))
}

func GetNotSupportedError(field string) error {
	return errors.NewFlyteAdminError(codes.Unimplemented, fmt.Sprintf(notSupportedFormat, field))
}
