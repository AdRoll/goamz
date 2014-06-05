//
// gosqs - Go packages to interact with the Amazon SQS Web Services.
//
// depends on https://wiki.ubuntu.com/goamz
//
//
// Written by Prudhvi Krishna Surapaneni <me@prudhvi.net>
// Extended by Fabrizio Milo <mistobaan@gmail.com>
//
package sqs

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"
)

const debug = false

// The SQS type encapsulates operation with an SQS region.
type SQS struct {
	aws.Auth
	aws.Region
	private byte // Reserve the right of using private data.
}

// NewFrom Create A new SQS Client given an access and secret Key
// region must be one of "us.east, us.west, eu.west"
func NewFrom(accessKey, secretKey, region string) (*SQS, error) {

	auth := aws.Auth{AccessKey: accessKey, SecretKey: secretKey}
	aws_region := aws.USEast

	switch region {
	case "us.east":
		aws_region = aws.USEast
	case "us.west":
		aws_region = aws.USWest
	case "eu.west":
		aws_region = aws.EUWest
	default:
		return nil, errors.New(fmt.Sprintf("Unknow/Unsupported region %s", region))
	}

	aws_sqs := New(auth, aws_region)
	return aws_sqs, nil
}

// NewFrom Create A new SQS Client from an exisisting aws.Auth
func New(auth aws.Auth, region aws.Region) *SQS {
	return &SQS{auth, region, 0}
}

// Queue Reference to a Queue
type Queue struct {
	*SQS
	Url string
}

type CreateQueueResponse struct {
	QueueUrl         string `xml:"CreateQueueResult>QueueUrl"`
	ResponseMetadata ResponseMetadata
}

type GetQueueUrlResponse struct {
	QueueUrl         string `xml:"GetQueueUrlResult>QueueUrl"`
	ResponseMetadata ResponseMetadata
}

type ListQueuesResponse struct {
	QueueUrl         []string `xml:"ListQueuesResult>QueueUrl"`
	ResponseMetadata ResponseMetadata
}

type DeleteMessageResponse struct {
	ResponseMetadata ResponseMetadata
}

type DeleteQueueResponse struct {
	ResponseMetadata ResponseMetadata
}

type SendMessageResponse struct {
	AttributeMD5     string `xml:"SendMessageResult>MD5OfMessageAttributes"`
	MD5              string `xml:"SendMessageResult>MD5OfMessageBody"`
	Id               string `xml:"SendMessageResult>MessageId"`
	ResponseMetadata ResponseMetadata
}

type ReceiveMessageResponse struct {
	Messages         []Message `xml:"ReceiveMessageResult>Message"`
	ResponseMetadata ResponseMetadata
}

type Message struct {
	MessageId        string             `xml:"MessageId"`
	Body             string             `xml:"Body"`
	MD5OfBody        string             `xml:"MD5OfBody"`
	ReceiptHandle    string             `xml:"ReceiptHandle"`
	Attribute        []Attribute        `xml:"Attribute"`
	MessageAttribute []MessageAttribute `xml:"MessageAttribute"`
	DelaySeconds     int
}

type Attribute struct {
	Name  string `xml:"Name"`
	Value string `xml:"Value"`
}

type MessageAttribute struct {
	Name  string                `xml:"Name"`
	Value MessageAttributeValue `xml:"Value"`
}

type MessageAttributeValue struct {
	DataType    string `xml:"DataType"`
	StringValue string `xml:"StringValue"`
}

type ChangeMessageVisibilityResponse struct {
	ResponseMetadata ResponseMetadata
}

type GetQueueAttributesResponse struct {
	Attributes       []Attribute `xml:"GetQueueAttributesResult>Attribute"`
	ResponseMetadata ResponseMetadata
}

type ResponseMetadata struct {
	RequestId string
	BoxUsage  float64
}

type Error struct {
	StatusCode int
	Code       string
	Message    string
	RequestId  string
}

func (err *Error) Error() string {
	if err.Code == "" {
		return err.Message
	}
	return fmt.Sprintf("%s (%s)", err.Message, err.Code)
}

func (err *Error) String() string {
	return err.Message
}

type xmlErrors struct {
	RequestId string
	Errors    []Error `xml:"Errors>Error"`
	Error     Error
}

// CreateQueue create a queue with a specific name
func (s *SQS) CreateQueue(queueName string) (*Queue, error) {
	return s.CreateQueueWithTimeout(queueName, 30)
}

// CreateQueue create a queue with a specific name and a timeout
func (s *SQS) CreateQueueWithTimeout(queueName string, timeout int) (*Queue, error) {
	params := map[string]string{
		"VisibilityTimeout": strconv.Itoa(timeout),
	}
	return s.CreateQueueWithAttributes(queueName, params)
}

func (s *SQS) CreateQueueWithAttributes(queueName string, attrs map[string]string) (q *Queue, err error) {
	resp, err := s.newQueue(queueName, attrs)
	if err != nil {
		return nil, err
	}
	q = &Queue{s, resp.QueueUrl}
	return
}

// GetQueue get a reference to the given quename
func (s *SQS) GetQueue(queueName string) (*Queue, error) {
	var q *Queue
	resp, err := s.getQueueUrl(queueName)
	if err != nil {
		return q, err
	}
	q = &Queue{s, resp.QueueUrl}
	return q, nil
}

func (s *SQS) QueueFromArn(queueUrl string) (q *Queue) {
	q = &Queue{s, queueUrl}
	return
}

func (s *SQS) getQueueUrl(queueName string) (resp *GetQueueUrlResponse, err error) {
	resp = &GetQueueUrlResponse{}
	params := makeParams("GetQueueUrl")
	params["QueueName"] = queueName
	err = s.query("", params, resp)
	return resp, err
}

func (s *SQS) newQueue(queueName string, attrs map[string]string) (resp *CreateQueueResponse, err error) {
	resp = &CreateQueueResponse{}
	params := makeParams("CreateQueue")
	params["QueueName"] = queueName

	i := 1
	for k, v := range attrs {
		nameParam := fmt.Sprintf("Attribute.%d.Name", i)
		valParam := fmt.Sprintf("Attribute.%d.Value", 1)
		params[nameParam] = k
		params[valParam] = v
		i++
	}

	err = s.query("", params, resp)
	return
}

func (s *SQS) ListQueues(QueueNamePrefix string) (resp *ListQueuesResponse, err error) {
	resp = &ListQueuesResponse{}
	params := makeParams("ListQueues")

	if QueueNamePrefix != "" {
		params["QueueNamePrefix"] = QueueNamePrefix
	}

	err = s.query("", params, resp)
	return
}

func (q *Queue) Delete() (resp *DeleteQueueResponse, err error) {
	resp = &DeleteQueueResponse{}
	params := makeParams("DeleteQueue")

	err = q.SQS.query(q.Url, params, resp)
	return
}

func (q *Queue) SendMessageWithDelay(MessageBody string, DelaySeconds int64) (resp *SendMessageResponse, err error) {
	resp = &SendMessageResponse{}
	params := makeParams("SendMessage")

	params["MessageBody"] = MessageBody
	params["DelaySeconds"] = strconv.Itoa(int(DelaySeconds))

	err = q.SQS.query(q.Url, params, resp)
	return
}

func (q *Queue) SendMessageWithAttributes(MessageBody string, MessageAttributes map[string]string) (resp *SendMessageResponse, err error) {
	resp = &SendMessageResponse{}
	params := makeParams("SendMessage")

	params["MessageBody"] = MessageBody

	// Add attributes (currently only supports string values)
	i := 1
	for k, v := range MessageAttributes {
		params[fmt.Sprintf("MessageAttribute.%d.Name", i)] = k
		params[fmt.Sprintf("MessageAttribute.%d.Value.StringValue", i)] = v
		params[fmt.Sprintf("MessageAttribute.%d.Value.DataType", i)] = "String"
		i++
	}

	err = q.SQS.query(q.Url, params, resp)

	// Assert we have expected Attribute MD5 if we've passed any Message Attributes
	if len(MessageAttributes) > 0 {
		expectedAttributeMD5 := fmt.Sprintf("%x", calculateAttributeMD5(MessageAttributes))

		if expectedAttributeMD5 != resp.AttributeMD5 {
			return resp, errors.New(fmt.Sprintf("Attribute MD5 mismatch, expecting `%v`, found `%v`", expectedAttributeMD5, resp.AttributeMD5))
		}
	}

	return
}

func (q *Queue) SendMessage(MessageBody string) (resp *SendMessageResponse, err error) {
	return q.SendMessageWithAttributes(MessageBody, map[string]string{})
}

// ReceiveMessageWithVisibilityTimeout
func (q *Queue) ReceiveMessageWithVisibilityTimeout(MaxNumberOfMessages, VisibilityTimeoutSec int) (*ReceiveMessageResponse, error) {
	params := map[string]string{
		"MaxNumberOfMessages": strconv.Itoa(MaxNumberOfMessages),
		"VisibilityTimeout":   strconv.Itoa(VisibilityTimeoutSec),
	}
	return q.ReceiveMessageWithParameters(params)
}

// ReceiveMessage
func (q *Queue) ReceiveMessage(MaxNumberOfMessages int) (*ReceiveMessageResponse, error) {
	params := map[string]string{
		"MaxNumberOfMessages": strconv.Itoa(MaxNumberOfMessages),
	}
	return q.ReceiveMessageWithParameters(params)
}

func (q *Queue) ReceiveMessageWithParameters(p map[string]string) (resp *ReceiveMessageResponse, err error) {
	resp = &ReceiveMessageResponse{}
	params := makeParams("ReceiveMessage")
	params["AttributeName"] = "All"

	for k, v := range p {
		params[k] = v
	}

	err = q.SQS.query(q.Url, params, resp)
	return
}

func (q *Queue) ChangeMessageVisibility(M *Message, VisibilityTimeout int) (resp *ChangeMessageVisibilityResponse, err error) {
	resp = &ChangeMessageVisibilityResponse{}
	params := makeParams("ChangeMessageVisibility")
	params["VisibilityTimeout"] = strconv.Itoa(VisibilityTimeout)
	params["ReceiptHandle"] = M.ReceiptHandle

	err = q.SQS.query(q.Url, params, resp)
	return
}

func (q *Queue) GetQueueAttributes(A string) (resp *GetQueueAttributesResponse, err error) {
	resp = &GetQueueAttributesResponse{}
	params := makeParams("GetQueueAttributes")
	params["AttributeName"] = A

	err = q.SQS.query(q.Url, params, resp)
	return
}

func (q *Queue) DeleteMessage(M *Message) (resp *DeleteMessageResponse, err error) {
	resp = &DeleteMessageResponse{}
	params := makeParams("DeleteMessage")
	params["ReceiptHandle"] = M.ReceiptHandle

	err = q.SQS.query(q.Url, params, resp)
	return
}

type SendMessageBatchResultEntry struct {
	Id               string `xml:"Id"`
	MessageId        string `xml:"MessageId"`
	MD5OfMessageBody string `xml:"MD5OfMessageBody"`
}

type SendMessageBatchResponse struct {
	SendMessageBatchResult []SendMessageBatchResultEntry `xml:"SendMessageBatchResult>SendMessageBatchResultEntry"`
	ResponseMetadata       ResponseMetadata
}

/* SendMessageBatch
 */
func (q *Queue) SendMessageBatch(msgList []Message) (resp *SendMessageBatchResponse, err error) {
	resp = &SendMessageBatchResponse{}
	params := makeParams("SendMessageBatch")

	for idx, msg := range msgList {
		count := idx + 1
		params[fmt.Sprintf("SendMessageBatchRequestEntry.%d.Id", count)] = fmt.Sprintf("msg-%d", count)
		params[fmt.Sprintf("SendMessageBatchRequestEntry.%d.MessageBody", count)] = msg.Body

		if msg.DelaySeconds > 0 {
			params[fmt.Sprintf("SendMessageBatchRequestEntry.%d.DelaySeconds", count)] = strconv.Itoa(msg.DelaySeconds)
		}
	}

	err = q.SQS.query(q.Url, params, resp)
	return
}

/* SendMessageBatchString
 */
func (q *Queue) SendMessageBatchString(msgList []string) (resp *SendMessageBatchResponse, err error) {
	resp = &SendMessageBatchResponse{}
	params := makeParams("SendMessageBatch")

	for idx, msg := range msgList {
		count := idx + 1
		params[fmt.Sprintf("SendMessageBatchRequestEntry.%d.Id", count)] = fmt.Sprintf("msg-%d", count)
		params[fmt.Sprintf("SendMessageBatchRequestEntry.%d.MessageBody", count)] = msg
	}

	err = q.SQS.query(q.Url, params, resp)
	return
}

type DeleteMessageBatchResponse struct {
	DeleteMessageBatchResult []struct {
		Id          string
		SenderFault bool
		Code        string
		Message     string
	} `xml:"DeleteMessageBatchResult>DeleteMessageBatchResultEntry"`
	ResponseMetadata ResponseMetadata
}

/* DeleteMessageBatch */
func (q *Queue) DeleteMessageBatch(msgList []Message) (resp *DeleteMessageBatchResponse, err error) {
	resp = &DeleteMessageBatchResponse{}
	params := makeParams("DeleteMessageBatch")

	lutMsg := make(map[string]Message)

	for idx := range msgList {
		params[fmt.Sprintf("DeleteMessageBatchRequestEntry.%d.Id", idx+1)] = msgList[idx].MessageId
		params[fmt.Sprintf("DeleteMessageBatchRequestEntry.%d.ReceiptHandle", idx+1)] = msgList[idx].ReceiptHandle

		lutMsg[string(msgList[idx].MessageId)] = msgList[idx]
	}

	err = q.SQS.query(q.Url, params, resp)

	messageWithErrors := make([]Message, 0, len(msgList))

	for idx := range resp.DeleteMessageBatchResult {
		if resp.DeleteMessageBatchResult[idx].SenderFault {
			msg, ok := lutMsg[resp.DeleteMessageBatchResult[idx].Id]
			if ok {
				messageWithErrors = append(messageWithErrors, msg)
			}
		}
	}

	if len(messageWithErrors) > 0 {
		log.Printf("%d Message have not been sent", len(messageWithErrors))
	}

	return
}

func (s *SQS) query(queueUrl string, params map[string]string, resp interface{}) (err error) {
	params["Version"] = "2011-10-01"
	params["Timestamp"] = time.Now().In(time.UTC).Format(time.RFC3339)
	var url_ *url.URL

	var path string
	if queueUrl != "" && len(queueUrl) > len(s.Region.SQSEndpoint) {
		url_, err = url.Parse(queueUrl)
		path = queueUrl[len(s.Region.SQSEndpoint):]
	} else {
		url_, err = url.Parse(s.Region.SQSEndpoint)
		path = "/"
	}
	if err != nil {
		return err
	}

	//url_, err := url.Parse(s.Region.SQSEndpoint)
	//if err != nil {
	//	return err
	//}

	if s.Auth.Token() != "" {
		params["SecurityToken"] = s.Auth.Token()
	}
	sign(s.Auth, "GET", path, params, url_.Host)

	url_.RawQuery = multimap(params).Encode()

	if debug {
		log.Printf("GET ", url_.String())
	}

	r, err := http.Get(url_.String())
	if err != nil {
		return err
	}

	defer r.Body.Close()

	if debug {
		dump, _ := httputil.DumpResponse(r, true)
		log.Printf("DUMP:\n", string(dump))
	}

	if r.StatusCode != 200 {
		return buildError(r)
	}
	err = xml.NewDecoder(r.Body).Decode(resp)
	io.Copy(ioutil.Discard, r.Body)

	return err
}

func buildError(r *http.Response) error {
	errors := xmlErrors{}
	xml.NewDecoder(r.Body).Decode(&errors)
	var err Error
	if len(errors.Errors) > 0 {
		err = errors.Errors[0]
	} else {
		err = errors.Error
	}
	err.RequestId = errors.RequestId
	err.StatusCode = r.StatusCode
	if err.Message == "" {
		err.Message = r.Status
	}
	return &err
}

func makeParams(action string) map[string]string {
	params := make(map[string]string)
	params["Action"] = action
	return params
}

func multimap(p map[string]string) url.Values {
	q := make(url.Values, len(p))
	for k, v := range p {
		q[k] = []string{v}
	}
	return q
}
