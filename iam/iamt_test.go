package iam_test

import (
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/iam"
	"launchpad.net/goamz/iam/iamtest"
	. "launchpad.net/gocheck"
)

// LocalServer represents a local ec2test fake server.
type LocalServer struct {
	auth   aws.Auth
	region aws.Region
	srv    *iamtest.Server
}

func (s *LocalServer) SetUp(c *C) {
	srv, err := iamtest.NewServer()
	c.Assert(err, IsNil)
	c.Assert(srv, NotNil)

	s.srv = srv
	s.region = aws.Region{IAMEndpoint: srv.URL()}
}

// LocalServerSuite defines tests that will run
// against the local iamtest server. It includes
// selected tests from ClientTests;
// when the ec2test functionality is sufficient, it should
// include all of them, and ClientTests can be simply embedded.
type LocalServerSuite struct {
	srv LocalServer
	clientTests ClientTests
}

var _ = Suite(&LocalServerSuite{})

func (s *LocalServerSuite) SetUpSuite(c *C) {
	s.srv.SetUp(c)
	s.clientTests.iam = iam.New(s.srv.auth, s.srv.region)
}

func (s *LocalServerSuite) TestCreateAndDeleteUser(c *C) {
	s.clientTests.TestCreateAndDeleteUser(c)
}
