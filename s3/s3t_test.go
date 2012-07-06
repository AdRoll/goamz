package s3_test

import (
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"launchpad.net/goamz/s3/s3test"
	. "launchpad.net/gocheck"
)

type LocalServer struct {
	auth   aws.Auth
	region aws.Region
	srv    *s3test.Server
}

func (s *LocalServer) SetUp(c *C) {
	srv, err := s3test.NewServer()
	c.Assert(err, IsNil)
	c.Assert(srv, NotNil)

	s.srv = srv
	s.region = aws.Region{
		Name:                 "faux-region-1",
		S3Endpoint:           srv.URL(),
		S3LocationConstraint: true, // s3test server requires a LocationConstraint
	}
}

// LocalServerSuite defines tests that will run
// against the local s3test server. It includes
// selected tests from ClientTests;
// when the s3test functionality is sufficient, it should
// include all of them, and ClientTests can be simply embedded.
type LocalServerSuite struct {
	srv         LocalServer
	clientTests ClientTests
}

var _ = Suite(&LocalServerSuite{})

func (s *LocalServerSuite) SetUpSuite(c *C) {
	s.srv.SetUp(c)
	s.clientTests.s3 = s3.New(s.srv.auth, s.srv.region)

	// TODO Sadly the fake server ignores auth completely right now. :-(
	s.clientTests.authIsBroken = true
}

func (s *LocalServerSuite) TestBasicFunctionality(c *C) {
	s.clientTests.TestBasicFunctionality(c)
}

func (s *LocalServerSuite) TestGetNotFound(c *C) {
	s.clientTests.TestGetNotFound(c)
}

func (s *LocalServerSuite) TestBucketList(c *C) {
	s.clientTests.TestBucketList(c)
}
