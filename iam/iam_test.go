package iam_test

import (
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/iam"
	. "launchpad.net/gocheck"
)

type S struct {
	HTTPSuite
	iam *iam.IAM
}

var _ = Suite(&S{})

func (s *S) SetUpSuite(c *C) {
	s.HTTPSuite.SetUpSuite(c)
	auth := aws.Auth{"abc", "123"}
	s.iam = iam.New(auth, aws.Region{IAMEndpoint: testServer.URL})
}

func (s *S) TestCreateUser(c *C) {
	testServer.PrepareResponse(200, nil, CreateUserExample)
	resp, err := s.iam.CreateUser("Bob", "/division_abc/subdivision_xyz/")
	values := testServer.WaitRequest().URL.Query()
	c.Assert(values.Get("Action"), Equals, "CreateUser")
	c.Assert(values.Get("UserName"), Equals, "Bob")
	c.Assert(values.Get("Path"), Equals, "/division_abc/subdivision_xyz/")
	c.Assert(err, IsNil)
	c.Assert(resp.RequestId, Equals, "7a62c49f-347e-4fc4-9331-6e8eEXAMPLE")
	expected := iam.User{
		Path: "/division_abc/subdivision_xyz/",
		Name: "Bob",
		Id:   "AIDACKCEVSQ6C2EXAMPLE",
		Arn:  "arn:aws:iam::123456789012:user/division_abc/subdivision_xyz/Bob",
	}
	c.Assert(resp.User, DeepEquals, expected)
}

func (s *S) TestCreateUserConflict(c *C) {
	testServer.PrepareResponse(409, nil, DuplicateUserExample)
	resp, err := s.iam.CreateUser("Bob", "/division_abc/subdivision_xyz/")
	testServer.WaitRequest()
	c.Assert(resp, IsNil)
	c.Assert(err, NotNil)
	e, ok := err.(*iam.Error)
	c.Assert(ok, Equals, true)
	c.Assert(e.Message, Equals, "User with name Bob already exists.")
	c.Assert(e.Code, Equals, "EntityAlreadyExists")
}

func (s *S) TestGetUser(c *C) {
	testServer.PrepareResponse(200, nil, GetUserExample)
	resp, err := s.iam.GetUser("Bob")
	values := testServer.WaitRequest().URL.Query()
	c.Assert(values.Get("Action"), Equals, "GetUser")
	c.Assert(values.Get("UserName"), Equals, "Bob")
	c.Assert(err, IsNil)
	c.Assert(resp.RequestId, Equals, "7a62c49f-347e-4fc4-9331-6e8eEXAMPLE")
	expected := iam.User{
		Path: "/division_abc/subdivision_xyz/",
		Name: "Bob",
		Id:   "AIDACKCEVSQ6C2EXAMPLE",
		Arn:  "arn:aws:iam::123456789012:user/division_abc/subdivision_xyz/Bob",
	}
	c.Assert(resp.User, DeepEquals, expected)
}

func (s *S) TestDeleteUser(c *C) {
	testServer.PrepareResponse(200, nil, RequestIdExample)
	resp, err := s.iam.DeleteUser("Bob")
	values := testServer.WaitRequest().URL.Query()
	c.Assert(values.Get("Action"), Equals, "DeleteUser")
	c.Assert(values.Get("UserName"), Equals, "Bob")
	c.Assert(err, IsNil)
	c.Assert(resp.RequestId, Equals, "7a62c49f-347e-4fc4-9331-6e8eEXAMPLE")
}

func (s *S) TestCreateAccessKey(c *C) {
	testServer.PrepareResponse(200, nil, CreateAccessKeyExample)
	resp, err := s.iam.CreateAccessKey("Bob")
	values := testServer.WaitRequest().URL.Query()
	c.Assert(values.Get("Action"), Equals, "CreateAccessKey")
	c.Assert(values.Get("UserName"), Equals, "Bob")
	c.Assert(err, IsNil)
	c.Assert(resp.AccessKey.UserName, Equals, "Bob")
	c.Assert(resp.AccessKey.Id, Equals, "AKIAIOSFODNN7EXAMPLE")
	c.Assert(resp.AccessKey.Secret, Equals, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYzEXAMPLEKEY")
	c.Assert(resp.AccessKey.Status, Equals, "Active")
}
