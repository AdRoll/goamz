package route53

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"io"
	"net/http"
)

type Route53 struct {
	Auth     aws.Auth
	Endpoint string
	Signer   *aws.Route53Signer
	Service  *aws.Service
}

const route53_host = "https://route53.amazonaws.com"

// Factory for the route53 type
func NewRoute53(auth aws.Auth) (*Route53, error) {
	signer := aws.NewRoute53Signer(auth)

	return &Route53{
		Auth:     auth,
		Signer:   signer,
		Endpoint: route53_host + "/2012-12-12/hostedzone",
	}, nil
}

// General Structs used in all types of requests
type HostedZones struct {
	XMLName    xml.Name `xml:"HostedZones"`
	HostedZone []HostedZone
}

type HostedZone struct {
	XMLName                xml.Name `xml:"HostedZone"`
	Id                     string
	Name                   string
	CallerReference        string
	Config                 Config
	ResourceRecordSetCount int
}

type Config struct {
	XMLName xml.Name `xml:"Config"`
	Comment string
}

// Structs for getting the existing Hosted Zones
type ListHostedZonesResponse struct {
	XMLName     xml.Name `xml:"ListHostedZonesResponse"`
	HostedZones []HostedZones
	Marker      string
	IsTruncated bool
	NextMarker  string
	MaxItems    int
}

// Structs for Creating a New Host
type CreateHostedZoneRequest struct {
	XMLName          xml.Name `xml:"CreateHostedZoneRequest"`
	Xmlns            string   `xml:"xmlns,attr"`
	Name             string
	CallerReference  string
	HostedZoneConfig HostedZoneConfig
}

type HostedZoneConfig struct {
	XMLName xml.Name `xml:"HostedZoneConfig"`
	Comment string
}

type CreateHostedZoneResponse struct {
	XMLName       xml.Name `xml:"CreateHostedZoneResponse"`
	HostedZone    HostedZone
	ChangeInfo    ChangeInfo
	DelegationSet DelegationSet
}

type ChangeInfo struct {
	XMLName     xml.Name `xml:"ChangeInfo"`
	Id          string
	Status      string
	SubmittedAt string
}

type DelegationSet struct {
	XMLName     xml.Name `xml:"DelegationSet`
	NameServers NameServers
}

type NameServers struct {
	XMLName    xml.Name `xml:"NameServers`
	NameServer []string
}

type GetHostedZoneResponse struct {
	XMLName       xml.Name `xml:"GetHostedZoneResponse"`
	HostedZone    HostedZone
	DelegationSet DelegationSet
}

type DeleteHostedZoneResponse struct {
	XMLName    xml.Name `xml:"DeleteHostedZoneResponse"`
	Xmlns      string   `xml:"xmlns,attr"`
	ChangeInfo ChangeInfo
}

// query sends the specified HTTP request to the path and signs the request
// with the required authentication and headers based on the Auth.
//
// Automatically decodes the response into the the result interface
func (r *Route53) query(method string, path string, body io.Reader, result interface{}) error {
	var err error

	// Create the POST request and sign the headers
	req, err := http.NewRequest(method, path, body)
	r.Signer.Sign(req)

	// Send the request and capture the response
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if method == "POST" {
		defer req.Body.Close()
	}

	if res.StatusCode != 201 && res.StatusCode != 200 {
		err = r.Service.BuildError(res)
		return err
	}

	err = xml.NewDecoder(res.Body).Decode(result)

	return err
}

// CreateHostedZone send a creation request to the AWS Route53 API
func (r *Route53) CreateHostedZone(hostedZoneReq *CreateHostedZoneRequest) (*CreateHostedZoneResponse, error) {
	xmlBytes, err := xml.Marshal(hostedZoneReq)
	if err != nil {
		return nil, err
	}

	result := new(CreateHostedZoneResponse)
	err = r.query("POST", r.Endpoint, bytes.NewBuffer(xmlBytes), result)

	return result, err
}

// ListedHostedZones fetches a collection of HostedZones through the AWS Route53 API
func (r *Route53) ListHostedZones(marker string, maxItems int) (result *ListHostedZonesResponse, err error) {
	path := fmt.Sprintf("%s?marker=%v&maxitems=%d", r.Endpoint, marker, maxItems)

	result = new(ListHostedZonesResponse)
	err = r.query("GET", path, nil, result)

	return
}

// GetHostedZone fetches a particular hostedzones DelegationSet by id
func (r *Route53) GetHostedZone(id string) (result *GetHostedZoneResponse, err error) {
	result = new(GetHostedZoneResponse)
	err = r.query("GET", fmt.Sprintf("%s/%v", r.Endpoint, id), nil, result)

	return
}

// DeleteHostedZone deletes the hosted zone with the given id
func (r *Route53) DeleteHostedZone(id string) (result *DeleteHostedZoneResponse, err error) {
	path := fmt.Sprintf("%s/%s", r.Endpoint, id)

	result = new(DeleteHostedZoneResponse)
	err = r.query("DELETE", path, nil, result)

	return
}
