package platform

import "errors"

func combineIgnoreErrors(ignoreErrors []error, ignoreErrorsFns []func(error) bool) func(error) bool {
	return func(err error) bool {
		for _, ignoreErr := range ignoreErrors {
			if errors.Is(err, ignoreErr) {
				return true
			}
		}

		for _, ignoreErrFn := range ignoreErrorsFns {
			if ignoreErrFn(err) {
				return true
			}
		}

		return false
	}
}

func isIgnoreError(err error, ignoreErrorsFn func(error) bool) bool {
	if multiErr, ok := err.(interface{ Unwrap() []error }); ok {
		for _, e := range multiErr.Unwrap() {
			if !isIgnoreError(e, ignoreErrorsFn) {
				return false
			}
		}
		return true
	}

	return ignoreErrorsFn(err)
}
