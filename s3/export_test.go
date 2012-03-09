package s3

import (
	"launchpad.net/goamz/aws"
)

func Sign(auth aws.Auth, method, path string, params, headers map[string][]string) {
	sign(auth, method, path, params, headers)
}
