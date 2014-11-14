package mturk_test

import (
	"github.com/rightscale/goamz/aws"
	"github.com/rightscale/goamz/exp/mturk"
	"gopkg.in/check.v1"
)

// Mechanical Turk REST authentication docs: http://goo.gl/wrzfn

var testAuth = aws.Auth{AccessKey: "user", SecretKey: "secret"}

// == fIJy9wCApBNL2R4J2WjJGtIBFX4=
func (s *S) TestBasicSignature(c *check.C) {
	params := map[string]string{}
	mturk.Sign(testAuth, "AWSMechanicalTurkRequester", "CreateHIT", "2012-02-16T20:30:47Z", params)
	expected := "b/TnvzrdeD/L/EyzdFrznPXhido="
	c.Assert(params["Signature"], check.Equals, expected)
}
