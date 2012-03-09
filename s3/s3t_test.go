package s3_test

import (
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"launchpad.net/goamz/s3/s3test"
	. "launchpad.net/gocheck"
)

// SuiteT defines tests to run against the s3test server
type SuiteT struct {
	i SI
}

var _ = Suite(&SuiteT{})

func (s *SuiteT) SetUpSuite(c *C) {
	srv, err := s3test.NewServer()
	c.Assert(err, IsNil)
	c.Assert(srv, NotNil)

	s.i.s3 = s3.New(
		aws.Auth{},
		aws.Region{S3Endpoint: srv.URL()},
	)
}
func (s *SuiteT) TestBasicFunctionality(c *C) {
	s.i.TestBasicFunctionality(c)
}
func (s *SuiteT) TestGetNotFound(c *C) {
	s.i.TestGetNotFound(c)
}
