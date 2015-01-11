package sns

import (
	"github.com/crowdmob/goamz/aws"
	"time"
)

type Topic struct {
	TopicArn string
}

type Subscription struct {
	Endpoint        string
	Owner           string
	Protocol        string
	SubscriptionArn string
	TopicArn        string
}

type Attribute struct {
	Key   string `xml:"key"`
	Value string `xml:"value"`
}

type Permission struct {
	ActionName string
	AccountId  string
}

type PlatformApplication struct {
	Attributes             []Attribute `xml:"Attributes>entry"`
	PlatformApplicationArn string
}

type Endpoint struct {
	EndpointArn string      `xml:"EndpointArn"`
	Attributes  []Attribute `xml:"Attributes>entry"`
}

// ============ Request ============

type PublishOptions struct {
	Message          string
	MessageStructure string
	Subject          string
	TopicArn         string
	TargetArn        string
}

type PlatformEndpointOptions struct {
	Attributes             []Attribute
	PlatformApplicationArn string
	CustomUserData         string
	Token                  string
}

// ============ Response ============

type ListTopicsResponse struct {
	NextToken        string  `xml:"ListTopicsResult>NextToken"`
	Topics           []Topic `xml:"ListTopicsResult>Topics>member"`
	ResponseMetadata aws.ResponseMetadata
	Error            aws.Error
}

type CreateTopicResponse struct {
	Topic            Topic `xml:"CreateTopicResult"`
	ResponseMetadata aws.ResponseMetadata
}

type DeleteTopicResponse struct {
	ResponseMetadata aws.ResponseMetadata
}

type ListSubscriptionsResponse struct {
	NextToken        string         `xml:"ListSubscriptionsResult>NextToken"`
	Subscriptions    []Subscription `xml:"ListSubscriptionsResult>Subscriptions>member"`
	ResponseMetadata aws.ResponseMetadata
}

type GetTopicAttributesResponse struct {
	Attributes       []Attribute `xml:"GetTopicAttributesResult>Attributes>entry"`
	ResponseMetadata aws.ResponseMetadata
}

type SetTopicAttributesResponse struct {
	ResponseMetadata aws.ResponseMetadata
}

type PublishResponse struct {
	MessageId        string `xml:"PublishResult>MessageId"`
	ResponseMetadata aws.ResponseMetadata
}

type SubscribeResponse struct {
	SubscriptionArn  string `xml:"SubscribeResult>SubscriptionArn"`
	ResponseMetadata aws.ResponseMetadata
}

type UnsubscribeResponse struct {
	ResponseMetadata aws.ResponseMetadata
}

type GetSubscriptionAttributesResponse struct {
	Attributes       []Attribute `xml:"GetSubscriptionAttributesResult>Attributes>entry"`
	ResponseMetadata aws.ResponseMetadata
}

type SetSubscriptionAttributesResponse struct {
	ResponseMetadata aws.ResponseMetadata
}

type ConfirmSubscriptionResponse struct {
	SubscriptionArn  string `xml:"ConfirmSubscriptionResult>SubscriptionArn"`
	ResponseMetadata aws.ResponseMetadata
}

type AddPermissionResponse struct {
	ResponseMetadata aws.ResponseMetadata
}

type RemovePermissionResponse struct {
	ResponseMetadata aws.ResponseMetadata
}

type ListSubscriptionByTopicResponse struct {
	NextToken        string         `xml:"ListSubscriptionsByTopicResult>NextToken"`
	Subscriptions    []Subscription `xml:"ListSubscriptionsByTopicResult>Subscriptions>member"`
	ResponseMetadata aws.ResponseMetadata
}

type CreatePlatformApplicationResponse struct {
	PlatformApplicationArn string `xml:"CreatePlatformApplicationResult>PlatformApplicationArn"`
	ResponseMetadata       aws.ResponseMetadata
}

type CreatePlatformEndpointResponse struct {
	EndpointArn      string `xml:"CreatePlatformEndpointResult>EndpointArn"`
	ResponseMetadata aws.ResponseMetadata
}

type DeleteEndpointResponse struct {
	ResponseMetadata aws.ResponseMetadata
}

type DeletePlatformApplicationResponse struct {
	ResponseMetadata aws.ResponseMetadata
}

type GetEndpointAttributesResponse struct {
	Attributes       []Attribute `xml:"GetEndpointAttributesResult>Attributes>entry"`
	ResponseMetadata aws.ResponseMetadata
}

type GetPlatformApplicationAttributesResponse struct {
	Attributes       []Attribute `xml:"GetPlatformApplicationAttributesResult>Attributes>entry"`
	ResponseMetadata aws.ResponseMetadata
}

type ListEndpointsByPlatformApplicationResponse struct {
	NextToken        string     `xml:"ListEndpointsByPlatformApplicationResult>NextToken"`
	Endpoints        []Endpoint `xml:"ListEndpointsByPlatformApplicationResult>Endpoints>member"`
	ResponseMetadata aws.ResponseMetadata
}

type ListPlatformApplicationsResponse struct {
	NextToken            string                `xml:"ListPlatformApplicationsResult>NextToken"`
	PlatformApplications []PlatformApplication `xml:"ListPlatformApplicationsResult>PlatformApplications>member"`
	ResponseMetadata     aws.ResponseMetadata
}

type SetEndpointAttributesResponse struct {
	ResponseMetadata aws.ResponseMetadata
}

type SetPlatformApplicationAttributesResponse struct {
	ResponseMetadata aws.ResponseMetadata
}

//===== Http Json messages ======

const (
	MESSAGE_TYPE_SUBSCRIPTION_CONFIRMATION = "SubscriptionConfirmation"
	MESSAGE_TYPE_UNSUBSCRIBE_CONFIRMATION  = "UnsubscribeConfirmation"
	MESSAGE_TYPE_NOTIFICATION              = "Notification"
)

// Json http notifications
// http://docs.aws.amazon.com/sns/latest/dg/json-formats.html#http-subscription-confirmation-json
// http://docs.aws.amazon.com/sns/latest/dg/json-formats.html#http-notification-json
// http://docs.aws.amazon.com/sns/latest/dg/json-formats.html#http-unsubscribe-confirmation-json
type HttpNotification struct {
	Type             string    `json:"Type"`
	MessageId        string    `json:"MessageId"`
	Token            string    `json:"Token" optional` // Only for subscribe and unsubscribe
	TopicArn         string    `json:"TopicArn"`
	Subject          string    `json:"Subject" optional` // Only for Notification
	Message          string    `json:"Message"`
	SubscribeURL     string    `json:"SubscribeURL" optional` // Only for subscribe and unsubscribe
	Timestamp        time.Time `json:"Timestamp"`
	SignatureVersion string    `json:"SignatureVersion"`
	Signature        string    `json:"Signature"`
	SigningCertURL   string    `json:"SigningCertURL"`
	UnsubscribeURL   string    `json:"UnsubscribeURL" optional` // Only for notifications
}
