package ses

import (
	"time"

	"github.com/crowdmob/goamz/aws"
)

//http://docs.aws.amazon.com/ses/latest/DeveloperGuide/mailbox-simulator.html
var (
	tEST_TO_ADDRESSES = []string{
		"success@simulator.amazonses.com",
		"bounce@simulator.amazonses.com",
		"ooto@simulator.amazonses.com",
		"complaint@simulator.amazonses.com",
		"suppressionlist@simulator.amazonses.com"}
	tEST_CC_ADDRESSES  = []string{}
	tEST_BCC_ADDRESSES = []string{}
)

const (
	tEST_EMAIL_SUBJECT = "goamz TestSESIntegration"
	tEST_TEXT_BODY     = "This is a test email send by goamz.TestSES_Integration."

	tEST_HTML_BODY = `
<html>
<body>
	<h1>This is a test email send by goamz.TestSES_Integration.</h1>
	<p>Foo bar baz</p>
</body>
</html>
`
)

// This is an helper function for integration tests between SES ans SNS.
// Use this method to send emails to the testing endpoint and read to SNS
// to process the bounces and complains.
//
// from: the source email address registered in your SES account
func SendSESIntegrationTestEmail(from string) (*SendEmailResponse, error) {
	awsAuth, err := aws.GetAuth("", "", "", time.Time{})
	if err != nil {
		return nil, err
	}
	server := New(awsAuth, aws.EUWest)

	destination := NewDestination(tEST_TO_ADDRESSES,
		tEST_CC_ADDRESSES, tEST_BCC_ADDRESSES)
	message := NewMessage(tEST_EMAIL_SUBJECT, tEST_TEXT_BODY, tEST_HTML_BODY)

	return server.SendEmail(from, destination, message)
}
