package mturk_test

import (
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/exp/mturk"
	"launchpad.net/gocheck"
)

// Mechanical Turk REST authentication docs: http://goo.gl/wrzfn

var testAuth = aws.Auth{AccessKey: "user", SecretKey: "secret", Token: ""}

// == fIJy9wCApBNL2R4J2WjJGtIBFX4=
func (s *S) TestBasicSignature(c *gocheck.C) {
	params := map[string]string{}
	mturk.Sign(testAuth, "AWSMechanicalTurkRequester", "CreateHIT", "2012-02-16T20:30:47Z", params)
	expected := "b/TnvzrdeD/L/EyzdFrznPXhido="
	c.Assert(params["Signature"], gocheck.Equals, expected)
}
