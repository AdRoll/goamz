package ses

import (
	"testing"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/testutil"
	"gopkg.in/check.v1"
)

const t_ERROR_RESPONSE = `
<?xml version="1.0"?>
<ErrorResponse xmlns="http://ses.amazonaws.com/doc/2010-12-01/">
	<Error>
		<Type>Sender</Type>
		<Code>MessageRejected</Code>
		<Message>Email address is not verified.</Message>
	</Error>
	<RequestId>21d1e58d-28b2-4d5f-a974-669c3c67674f</RequestId>
</ErrorResponse>
`

func Test(t *testing.T) {
	check.TestingT(t)
}

var _ = check.Suite(&S{})
var testServer = testutil.NewHTTPServer()

type S struct {
	sesService *SES
}

func (s *S) SetUpSuite(c *check.C) {
	testServer.Start()
	auth := aws.Auth{AccessKey: "abc", SecretKey: "123"}
	sesService := New(auth, aws.Region{SESEndpoint: testServer.URL})
	s.sesService = sesService
}

func (s *S) TearDownTest(c *check.C) {
	testServer.Flush()
}

func (s *S) TestBuildError(c *check.C) {
	testServer.Response(400, nil, TestSendEmailError)

	resp, err := s.sesService.SendEmail("foo@example.com",
		NewDestination([]string{"unauthorized@example.com"}, []string{}, []string{}),
		NewMessage("subject", "textBody", "htmlBody"))
	_ = testServer.WaitRequest()

	c.Assert(resp, check.IsNil)
	c.Assert(err.Error(), check.Equals, "Email address is not verified. (MessageRejected)")
}

func (s *S) TestSendEmail(c *check.C) {
	testServer.Response(200, nil, TestSendEmailOk)

	resp, err := s.sesService.SendEmail("foo@example.com",
		NewDestination([]string{"to1@example.com", "to2@example.com"},
			[]string{"cc1@example.com", "cc2@example.com"},
			[]string{"bcc1@example.com", "bcc2@example.com"}),
		NewMessage("subject", "textBody", "htmlBody"))
	req := testServer.WaitRequest()

	c.Assert(req.Method, check.Equals, "POST")
	c.Assert(req.URL.Path, check.Equals, "/")
	c.Assert(req.Header["Date"], check.Not(check.Equals), "")
	c.Assert(req.FormValue("Source"), check.Equals, "foo@example.com")
	c.Assert(req.FormValue("Destination.ToAddresses.member.1"), check.Equals, "to1@example.com")
	c.Assert(req.FormValue("Destination.ToAddresses.member.2"), check.Equals, "to2@example.com")
	c.Assert(req.FormValue("Destination.CcAddresses.member.1"), check.Equals, "cc1@example.com")
	c.Assert(req.FormValue("Destination.CcAddresses.member.2"), check.Equals, "cc2@example.com")
	c.Assert(req.FormValue("Destination.BccAddresses.member.1"), check.Equals, "bcc1@example.com")
	c.Assert(req.FormValue("Destination.BccAddresses.member.2"), check.Equals, "bcc2@example.com")

	c.Assert(req.FormValue("Message.Subject.Data"), check.Equals, "subject")
	c.Assert(req.FormValue("Message.Subject.Charset"), check.Equals, "utf-8")

	c.Assert(req.FormValue("Message.Body.Text.Data"), check.Equals, "textBody")
	c.Assert(req.FormValue("Message.Body.Text.Charset"), check.Equals, "utf-8")

	c.Assert(req.FormValue("Message.Body.Html.Data"), check.Equals, "htmlBody")
	c.Assert(req.FormValue("Message.Body.Html.Charset"), check.Equals, "utf-8")

	c.Assert(err, check.IsNil)
	c.Assert(resp.SendEmailResult, check.NotNil)
	c.Assert(resp.ResponseMetadata, check.NotNil)
}
