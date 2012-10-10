package iamtest

import (
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/iam"
	. "launchpad.net/gocheck"
	"testing"
)

type S struct {
	iam    *iam.IAM
	server *Server
}

var _ = Suite(&S{})

func Test(t *testing.T) {
	TestingT(t)
}

func (s *S) SetUpSuite(c *C) {
	var err error
	s.server, err = NewServer()
	c.Assert(err, IsNil)
	auth := aws.Auth{AccessKey: "access", SecretKey: "secret"}
	s.iam = iam.New(auth, aws.Region{IAMEndpoint: s.server.URL()})
}

func (s *S) TearDownSuite(c *C) {
	s.server.Quit()
}

func (s *S) SetUpTest(c *C) {
	s.server.users = []iam.User{}
}

func (s *S) TestCreateUser(c *C) {
	resp, err := s.iam.CreateUser("gopher", "/gopher/")
	c.Assert(err, IsNil)
	expected := []iam.User{resp.User}
	c.Assert(s.server.users, DeepEquals, expected)
}

func (s *S) TestDeleteUser(c *C) {
	_, err := s.iam.CreateUser("gopher", "/gopher/")
	c.Assert(err, IsNil)
	_, err = s.iam.DeleteUser("gopher")
	c.Assert(s.server.users, HasLen, 0)
}
