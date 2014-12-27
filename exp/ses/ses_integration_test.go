package ses

import (
	"log"
	"testing"
	"time"

	"github.com/crowdmob/goamz/aws"
)

//http://docs.aws.amazon.com/ses/latest/DeveloperGuide/mailbox-simulator.html
var (
	tEST_SOURCE_ADDRESS = "Funky name <noreply@example.com>"
	tEST_TO_ADDRESSES   = []string{"success@simulator.amazonses.com"}
	tEST_CC_ADDRESSES   = []string{}
	tEST_BCC_ADDRESSES  = []string{}
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

// This is an end to end test which can be used to manually test your environment
// integration with SES. This tests is normally Suppressed because this sends a
// real email which costs money.
//
// You can enable this manually by removing the underscore in front of the name.
// Don't forget to replace the email addresses in the variables with the ones you
// want to test.
func _TestSES_Integration(t *testing.T) {
	awsAuth, err := aws.GetAuth("", "", "", time.Time{})
	if err != nil {
		log.Fatal(err)
	}
	server := New(awsAuth, aws.EUWest)

	destination := NewDestination(tEST_TO_ADDRESSES,
		tEST_CC_ADDRESSES, tEST_BCC_ADDRESSES)
	message := NewMessage(tEST_EMAIL_SUBJECT, tEST_TEXT_BODY, tEST_HTML_BODY)

	resp, err := server.SendEmail(tEST_SOURCE_ADDRESS, destination, message)
	if err != nil {
		t.Fatal("Message delivery failed with error: %v\n", err)
	}
	log.Printf("Response: %v", resp)
}
