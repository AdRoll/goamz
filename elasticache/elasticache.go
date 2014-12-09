package elasticache

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/crowdmob/goamz/aws"
)

var (
	ErrCacheClusterNotFound = errors.New("Cache cluster not found")
)

type ElastiCache struct {
	aws.Auth
	aws.Region
}

// DescribeCacheClustersResult represents the response from a
// DescribeCacheClusters ElastiCache API call
type DescribeCacheClustersResult struct {
	CacheClusters []*CacheCluster `xml:"DescribeCacheClustersResult>CacheClusters"`
}

// CacheCluster represents a cache cluster
type CacheCluster struct {
	CacheClusterId string       `xml:"CacheCluster>CacheClusterId"`
	CacheNodes     []*CacheNode `xml:"CacheCluster>CacheNodes"`
}

// CacheNode represents a cache node
type CacheNode struct {
	Endpoint *Endpoint `xml:"CacheNode>Endpoint"`
}

// Endpoint represents a cache node endpoint
type Endpoint struct {
	Host string `xml:"Address"`
	Port int    `xml:"Port"`
}

// New creates a new ElastiCache instance
func New(auth aws.Auth, region aws.Region) *ElastiCache {
	return &ElastiCache{auth, region}
}

// Describe returns information about a cache cluster
func (ec *ElastiCache) Describe(cluster string) (*CacheCluster, error) {
	var resp DescribeCacheClustersResult
	err := ec.query("Action=DescribeCacheClusters&CacheClusterId="+cluster+"&ShowCacheNodeInfo=true", &resp)

	if err != nil {
		return nil, err
	}

	if len(resp.CacheClusters) == 0 {
		return nil, ErrCacheClusterNotFound
	}

	return resp.CacheClusters[0], nil
}

func (ec *ElastiCache) query(query string, response interface{}) error {
	url := ec.Region.ElastiCacheEndpoint + "/?" + query

	hreq, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	hreq.Header.Set("Content-Type", "application/x-amz-json-1.0")
	hreq.Header.Set("X-Amz-Date", time.Now().UTC().Format(aws.ISO8601BasicFormat))

	token := ec.Auth.Token()
	if token != "" {
		hreq.Header.Set("X-Amz-Security-Token", token)
	}

	signer := aws.NewV4Signer(ec.Auth, "elasticache", ec.Region)
	signer.Sign(hreq)

	resp, err := http.DefaultClient.Do(hreq)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return buildError(resp)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return xml.Unmarshal(b, &response)
}

/* Copied from elb/elb.go - might not be entirely accurate */

// Error encapsulates an error returned by EC.
type Error struct {
	// HTTP status code
	StatusCode int
	// AWS error code
	Code string
	// The human-oriented error message
	Message string
}

func (err *Error) Error() string {
	if err.Code == "" {
		return err.Message
	}

	return fmt.Sprintf("%s (%s)", err.Message, err.Code)
}

type xmlErrors struct {
	Errors []Error `xml:"Error"`
}

func buildError(r *http.Response) error {
	var (
		err    Error
		errors xmlErrors
	)
	xml.NewDecoder(r.Body).Decode(&errors)
	if len(errors.Errors) > 0 {
		err = errors.Errors[0]
	}
	err.StatusCode = r.StatusCode
	if err.Message == "" {
		err.Message = r.Status
	}
	return &err
}
