package sqs_test

import (
	"crypto/md5"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/sqs"
	"hash"
	"launchpad.net/gocheck"
)

var _ = gocheck.Suite(&S{})

type S struct {
	HTTPSuite
	sqs *sqs.SQS
}

func (s *S) SetUpSuite(c *gocheck.C) {
	s.HTTPSuite.SetUpSuite(c)
	auth := aws.Auth{AccessKey: "abc", SecretKey: "123"}
	s.sqs = sqs.New(auth, aws.Region{SQSEndpoint: testServer.URL})
}

func (s *S) TestCreateQueue(c *gocheck.C) {
	testServer.PrepareResponse(200, nil, TestCreateQueueXmlOK)

	resp, err := s.sqs.CreateQueue("testQueue")
	req := testServer.WaitRequest()

	c.Assert(req.Method, gocheck.Equals, "GET")
	c.Assert(req.URL.Path, gocheck.Equals, "/")
	c.Assert(req.Header["Date"], gocheck.Not(gocheck.Equals), "")

	c.Assert(resp.Url, gocheck.Equals, "http://sqs.us-east-1.amazonaws.com/123456789012/testQueue")
	c.Assert(err, gocheck.IsNil)
}

func (s *S) TestListQueues(c *gocheck.C) {
	testServer.PrepareResponse(200, nil, TestListQueuesXmlOK)

	resp, err := s.sqs.ListQueues("")
	req := testServer.WaitRequest()

	c.Assert(req.Method, gocheck.Equals, "GET")
	c.Assert(req.URL.Path, gocheck.Equals, "/")
	c.Assert(req.Header["Date"], gocheck.Not(gocheck.Equals), "")

	c.Assert(len(resp.QueueUrl), gocheck.Not(gocheck.Equals), 0)
	c.Assert(resp.QueueUrl[0], gocheck.Equals, "http://sqs.us-east-1.amazonaws.com/123456789012/testQueue")
	c.Assert(resp.ResponseMetadata.RequestId, gocheck.Equals, "725275ae-0b9b-4762-b238-436d7c65a1ac")
	c.Assert(err, gocheck.IsNil)
}

func (s *S) TestDeleteQueue(c *gocheck.C) {
	testServer.PrepareResponse(200, nil, TestDeleteQueueXmlOK)

	q := &sqs.Queue{s.sqs, testServer.URL + "/123456789012/testQueue/"}
	resp, err := q.Delete()
	req := testServer.WaitRequest()

	c.Assert(req.Method, gocheck.Equals, "GET")
	c.Assert(req.URL.Path, gocheck.Equals, "/123456789012/testQueue/")
	c.Assert(req.Header["Date"], gocheck.Not(gocheck.Equals), "")

	c.Assert(resp.ResponseMetadata.RequestId, gocheck.Equals, "6fde8d1e-52cd-4581-8cd9-c512f4c64223")
	c.Assert(err, gocheck.IsNil)
}

func (s *S) TestSendMessage(c *gocheck.C) {
	testServer.PrepareResponse(200, nil, TestSendMessageXmlOK)

	q := &sqs.Queue{s.sqs, testServer.URL + "/123456789012/testQueue/"}
	resp, err := q.SendMessage("This is a test message")
	req := testServer.WaitRequest()

	c.Assert(req.Method, gocheck.Equals, "GET")
	c.Assert(req.URL.Path, gocheck.Equals, "/123456789012/testQueue/")
	c.Assert(req.Header["Date"], gocheck.Not(gocheck.Equals), "")

	msg := "This is a test message"
	var h hash.Hash = md5.New()
	h.Write([]byte(msg))
	c.Assert(resp.MD5, gocheck.Equals, fmt.Sprintf("%x", h.Sum(nil)))
	c.Assert(resp.Id, gocheck.Equals, "5fea7756-0ea4-451a-a703-a558b933e274")
	c.Assert(err, gocheck.IsNil)
}

func (s *S) TestSendMessageBatch(c *gocheck.C) {
	testServer.PrepareResponse(200, nil, TestSendMessageBatchXmlOk)

	q := &sqs.Queue{s.sqs, testServer.URL + "/123456789012/testQueue/"}

	msgList := []string{"test message body 1", "test message body 2"}
	resp, err := q.SendMessageBatchString(msgList)
	req := testServer.WaitRequest()

	c.Assert(req.Method, gocheck.Equals, "GET")
	c.Assert(req.URL.Path, gocheck.Equals, "/123456789012/testQueue/")
	c.Assert(req.Header["Date"], gocheck.Not(gocheck.Equals), "")

	for idx, msg := range msgList {
		var h hash.Hash = md5.New()
		h.Write([]byte(msg))
		c.Assert(resp.SendMessageBatchResult[idx].MD5OfMessageBody, gocheck.Equals, fmt.Sprintf("%x", h.Sum(nil)))
		c.Assert(err, gocheck.IsNil)
	}
}

func (s *S) TestDeleteMessageBatch(c *gocheck.C) {
	testServer.PrepareResponse(200, nil, TestDeleteMessageBatchXmlOK)

	q := &sqs.Queue{s.sqs, testServer.URL + "/123456789012/testQueue/"}

	msgList := []sqs.Message{*(&sqs.Message{ReceiptHandle: "gfk0T0R0waama4fVFffkjPQrrvzMrOg0fTFk2LxT33EuB8wR0ZCFgKWyXGWFoqqpCIiprQUEhir%2F5LeGPpYTLzjqLQxyQYaQALeSNHb0us3uE84uujxpBhsDkZUQkjFFkNqBXn48xlMcVhTcI3YLH%2Bd%2BIqetIOHgBCZAPx6r%2B09dWaBXei6nbK5Ygih21DCDdAwFV68Jo8DXhb3ErEfoDqx7vyvC5nCpdwqv%2BJhU%2FTNGjNN8t51v5c%2FAXvQsAzyZVNapxUrHIt4NxRhKJ72uICcxruyE8eRXlxIVNgeNP8ZEDcw7zZU1Zw%3D%3D"}),
		*(&sqs.Message{ReceiptHandle: "gfk0T0R0waama4fVFffkjKzmhMCymjQvfTFk2LxT33G4ms5subrE0deLKWSscPU1oD3J9zgeS4PQQ3U30qOumIE6AdAv3w%2F%2Fa1IXW6AqaWhGsEPaLm3Vf6IiWqdM8u5imB%2BNTwj3tQRzOWdTOePjOjPcTpRxBtXix%2BEvwJOZUma9wabv%2BSw6ZHjwmNcVDx8dZXJhVp16Bksiox%2FGrUvrVTCJRTWTLc59oHLLF8sEkKzRmGNzTDGTiV%2BYjHfQj60FD3rVaXmzTsoNxRhKJ72uIHVMGVQiAGgB%2BqAbSqfKHDQtVOmJJgkHug%3D%3D"}),
	}

	resp, err := q.DeleteMessageBatch(msgList)
	c.Assert(err, gocheck.IsNil)
	req := testServer.WaitRequest()

	c.Assert(req.Method, gocheck.Equals, "GET")
	c.Assert(req.URL.Path, gocheck.Equals, "/123456789012/testQueue/")
	c.Assert(req.Header["Date"], gocheck.Not(gocheck.Equals), "")

	for idx, _ := range msgList {
		c.Assert(resp.DeleteMessageBatchResult[idx].Id, gocheck.Equals, fmt.Sprintf("msg%d", idx+1))
	}
}

func (s *S) TestReceiveMessage(c *gocheck.C) {
	testServer.PrepareResponse(200, nil, TestReceiveMessageXmlOK)

	q := &sqs.Queue{s.sqs, testServer.URL + "/123456789012/testQueue/"}
	resp, err := q.ReceiveMessage(5)
	req := testServer.WaitRequest()

	c.Assert(req.Method, gocheck.Equals, "GET")
	c.Assert(req.URL.Path, gocheck.Equals, "/123456789012/testQueue/")
	c.Assert(req.Header["Date"], gocheck.Not(gocheck.Equals), "")

	c.Assert(len(resp.Messages), gocheck.Not(gocheck.Equals), 0)
	c.Assert(resp.Messages[0].MessageId, gocheck.Equals, "5fea7756-0ea4-451a-a703-a558b933e274")
	c.Assert(resp.Messages[0].MD5OfBody, gocheck.Equals, "fafb00f5732ab283681e124bf8747ed1")
	c.Assert(resp.Messages[0].ReceiptHandle, gocheck.Equals, "MbZj6wDWli+JvwwJaBV+3dcjk2YW2vA3+STFFljTM8tJJg6HRG6PYSasuWXPJB+CwLj1FjgXUv1uSj1gUPAWV66FU/WeR4mq2OKpEGYWbnLmpRCJVAyeMjeU5ZBdtcQ+QEauMZc8ZRv37sIW2iJKq3M9MFx1YvV11A2x/KSbkJ0=")
	c.Assert(resp.Messages[0].Body, gocheck.Equals, "This is a test message")
	c.Assert(len(resp.Messages[0].Attribute), gocheck.Not(gocheck.Equals), 0)
	c.Assert(err, gocheck.IsNil)
}

func (s *S) TestChangeMessageVisibility(c *gocheck.C) {
	testServer.PrepareResponse(200, nil, TestReceiveMessageXmlOK)

	q := &sqs.Queue{s.sqs, testServer.URL + "/123456789012/testQueue/"}

	resp1, err := q.ReceiveMessage(1)
	req := testServer.WaitRequest()

	testServer.PrepareResponse(200, nil, TestChangeMessageVisibilityXmlOK)

	resp, err := q.ChangeMessageVisibility(&resp1.Messages[0], 50)
	req = testServer.WaitRequest()

	c.Assert(req.Method, gocheck.Equals, "GET")
	c.Assert(req.URL.Path, gocheck.Equals, "/123456789012/testQueue/")
	c.Assert(req.Header["Date"], gocheck.Not(gocheck.Equals), "")

	c.Assert(resp.ResponseMetadata.RequestId, gocheck.Equals, "6a7a282a-d013-4a59-aba9-335b0fa48bed")
	c.Assert(err, gocheck.IsNil)
}
