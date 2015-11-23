package ses_test

import (
	"encoding/json"
	"time"

	"gopkg.in/check.v1"

	"github.com/AdRoll/goamz/exp/ses"
)

func (s *S) TestSNSBounceNotificationUnmarshalling(c *check.C) {
	notification := ses.SNSNotification{}
	err := json.Unmarshal([]byte(SNSBounceNotification), &notification)
	c.Assert(err, check.IsNil)

	c.Assert(notification.NotificationType, check.Equals,
		ses.NOTIFICATION_TYPE_BOUNCE)

	c.Assert(notification.Mail, check.NotNil)
	c.Assert(notification.Mail.Timestamp, check.DeepEquals,
		parseJsonTime("2012-06-19T01:05:45.000Z"))
	c.Assert(notification.Mail.Source, check.Equals, "sender@example.com")
	c.Assert(notification.Mail.MessageId, check.Equals,
		"00000138111222aa-33322211-cccc-cccc-cccc-ddddaaaa0680-000000")
	c.Assert(notification.Mail.Destination, check.DeepEquals,
		[]string{"username@example.com"})

	c.Assert(notification.Bounce, check.NotNil)
	c.Assert(notification.Bounce.BounceType, check.Equals,
		ses.BOUNCE_TYPE_PERMANENT)
	c.Assert(notification.Bounce.BounceSubType, check.Equals,
		ses.BOUNCE_SUBTYPE_GENERAL)
	c.Assert(notification.Bounce.FeedbackId, check.Equals,
		"00000138111222aa-33322211-cccc-cccc-cccc-ddddaaaa068a-000000")
	c.Assert(notification.Bounce.Timestamp, check.DeepEquals,
		parseJsonTime("2012-06-19T01:07:52.000Z"))
	c.Assert(notification.Bounce.ReportingMTA, check.Equals,
		"dns; email.example.com")
	c.Assert(notification.Bounce.BouncedRecipients, check.DeepEquals,
		[]*ses.BouncedRecipient{
			&ses.BouncedRecipient{
				EmailAddress:   "username@example.com",
				Status:         "5.1.1",
				Action:         "failed",
				DiagnosticCode: "smtp; 550 5.1.1 <username@example.com>... User",
			},
		})

	c.Assert(notification.Complaint, check.IsNil)

	c.Assert(notification.Delivery, check.IsNil)
}

func (s *S) TestSNSComplaintNotificationUnmarshalling(c *check.C) {
	notification := ses.SNSNotification{}
	err := json.Unmarshal([]byte(SNSComplaintNotification), &notification)
	c.Assert(err, check.IsNil)
	c.Assert(notification.Complaint, check.NotNil)

	c.Assert(notification.NotificationType, check.Equals,
		ses.NOTIFICATION_TYPE_COMPLAINT)

	c.Assert(notification.Mail, check.NotNil)
	c.Assert(notification.Mail.Timestamp, check.DeepEquals,
		parseJsonTime("2012-05-25T14:59:38.623-07:00"))
	c.Assert(notification.Mail.Source, check.Equals,
		"email_1337983178623@amazon.com")
	c.Assert(notification.Mail.MessageId, check.Equals,
		"000001378603177f-7a5433e7-8edb-42ae-af10-f0181f34d6ee-000000")
	c.Assert(notification.Mail.Destination, check.DeepEquals,
		[]string{"recipient1@example.com", "recipient2@example.com",
			"recipient3@example.com", "recipient4@example.com"})

	c.Assert(notification.Bounce, check.IsNil)

	c.Assert(notification.Complaint, check.NotNil)
	c.Assert(notification.Complaint.FeedbackId, check.Equals,
		"000001378603177f-18c07c78-fa81-4a58-9dd1-fedc3cb8f49a-000000")
	c.Assert(notification.Complaint.Timestamp, check.DeepEquals,
		parseJsonTime("2012-05-25T14:59:38.623-07:00"))
	c.Assert(notification.Complaint.ArrivalDate, check.DeepEquals,
		parseJsonTime("2009-12-03T04:24:21.000-05:00"))
	c.Assert(notification.Complaint.UserAgent, check.Equals,
		"Comcast Feedback Loop (V0.01)")
	c.Assert(notification.Complaint.ComplaintFeedbackType, check.Equals,
		ses.COMPLAINT_FEEDBACK_TYPE_ABUSE)
	c.Assert(notification.Complaint.ComplainedRecipients, check.DeepEquals,
		[]*ses.ComplainedRecipient{
			&ses.ComplainedRecipient{
				EmailAddress: "recipient1@example.com",
			},
		})

	c.Assert(notification.Delivery, check.IsNil)
}

func (s *S) TestSNSDeliveryNotificationUnmarshalling(c *check.C) {
	notification := ses.SNSNotification{}
	err := json.Unmarshal([]byte(SNSDeliveryNotification), &notification)
	c.Assert(err, check.IsNil)

	c.Assert(notification.NotificationType, check.Equals,
		ses.NOTIFICATION_TYPE_DELIVERY)

	c.Assert(notification.Mail, check.NotNil)
	c.Assert(notification.Mail.Timestamp, check.DeepEquals,
		parseJsonTime("2014-05-28T22:40:59.638Z"))
	c.Assert(notification.Mail.Source, check.Equals,
		"test@ses-example.com")
	c.Assert(notification.Mail.MessageId, check.Equals,
		"0000014644fe5ef6-9a483358-9170-4cb4-a269-f5dcdf415321-000000")
	c.Assert(notification.Mail.Destination, check.DeepEquals,
		[]string{"success@simulator.amazonses.com",
			"recipient@ses-example.com"})

	c.Assert(notification.Bounce, check.IsNil)

	c.Assert(notification.Complaint, check.IsNil)

	c.Assert(notification.Delivery, check.NotNil)
	c.Assert(notification.Delivery.Timestamp, check.DeepEquals,
		parseJsonTime("2014-05-28T22:41:01.184Z"))
	c.Assert(notification.Delivery.ReportingMTA, check.Equals,
		"a8-70.smtp-out.amazonses.com")
	c.Assert(notification.Delivery.ProcessingTimeMillis, check.Equals,
		int64(546))
	c.Assert(notification.Delivery.SmtpResponse, check.Equals,
		"250 ok:  Message 64111812 accepted")
	c.Assert(notification.Delivery.Recipients, check.DeepEquals,
		[]string{"success@simulator.amazonses.com"})
}

func parseJsonTime(str string) time.Time {
	t, _ := time.Parse(time.RFC3339, str)
	return t
}
