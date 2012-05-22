//
// goamz - Go packages to interact with the Amazon Web Services.
//
//   https://wiki.ubuntu.com/goamz
//
// Copyright (c) 2011 Canonical Ltd.
//
// Written by Gustavo Niemeyer <gustavo.niemeyer@canonical.com>
//
package s3

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"launchpad.net/goamz/aws"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const debug = false

// The S3 type encapsulates operations with an S3 region.
type S3 struct {
	aws.Auth
	aws.Region
	private byte // Reserve the right of using private data.
}

// The Bucket type encapsulates operations with an S3 bucket.
type Bucket struct {
	*S3
	Name string
}

// The Owner type represents the owner of the object in an S3 bucket.
type Owner struct {
	ID          string
	DisplayName string
}

// New creates a new S3.
func New(auth aws.Auth, region aws.Region) *S3 {
	return &S3{auth, region, 0}
}

// Bucket returns a Bucket with the given name.
func (s3 *S3) Bucket(name string) *Bucket {
	if s3.Region.S3BucketEndpoint != "" {
		// If passing bucket name via hostname, it is necessarily lowercased.
		name = strings.ToLower(name)
	}
	return &Bucket{s3, name}
}

// ----------------------------------------------------------------------------
// Bucket-level operations.

type ACL string

const (
	Private           = ACL("private")
	PublicRead        = ACL("public-read")
	PublicReadWrite   = ACL("public-read-write")
	AuthenticatedRead = ACL("authenticated-read")
	BucketOwnerRead   = ACL("bucket-owner-read")
	BucketOwnerFull   = ACL("bucket-owner-full-control")
)

// PutBucket creates a new bucket.
//
// See http://goo.gl/ndjnR for more details.
func (b *Bucket) PutBucket(perm ACL) error {
	headers := map[string][]string{
		"x-amz-acl": {string(perm)},
	}
	_, err := b.S3.query("PUT", b.Name, "/", nil, headers, nil, nil)
	return err
}

// DelBucket removes an existing S3 bucket. All objects in the bucket must
// be removed before the bucket itself can be removed.
//
// See http://goo.gl/GoBrY for more details.
func (b *Bucket) DelBucket() error {
	_, err := b.S3.query("DELETE", b.Name, "/", nil, nil, nil, nil)
	return err
}

// ----------------------------------------------------------------------------
// Operations for bucket objects.

// Get retrieves an object from an S3 bucket.
//
// See http://goo.gl/isCO7 for more details.
func (b *Bucket) Get(path string) (data []byte, err error) {
	body, err := b.GetReader(path)
	if err != nil {
		return nil, err
	}
	data, err = ioutil.ReadAll(body)
	body.Close()
	return data, err
}

// GetReader retrieves an object from an S3 bucket.
// It is the caller's responsibility to call Close on rc when
// finished reading.
func (b *Bucket) GetReader(path string) (rc io.ReadCloser, err error) {
	params := map[string][]string{}
	headers := map[string][]string{}

	resp, err := b.S3.query("GET", b.Name, path, params, headers, nil, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// Put inserts an object into the S3 bucket.
//
// See http://goo.gl/FEBPD for more details.
func (b *Bucket) Put(path string, data []byte, contType string, perm ACL) error {
	body := bytes.NewBuffer(data)
	return b.PutReader(path, body, int64(len(data)), contType, perm)
}

// PutReader inserts an object into the S3 bucket by consuming data
// from r until EOF.
func (b *Bucket) PutReader(path string, r io.Reader, length int64, contType string, perm ACL) error {
	headers := map[string][]string{
		"Content-Length": {strconv.FormatInt(length, 10)},
		"Content-Type":   {contType},
		"x-amz-acl":      {string(perm)},
	}
	_, err := b.S3.query("PUT", b.Name, path, nil, headers, r, nil)
	return err
}

// Del removes an object from the S3 bucket.
//
// See http://goo.gl/APeTt for more details.
func (b *Bucket) Del(path string) error {
	_, err := b.S3.query("DELETE", b.Name, path, nil, nil, nil, nil)
	return err
}

// The ListResp type holds the results of a List bucket operation.
type ListResp struct {
	Name      string
	Prefix    string
	Delimiter string
	Marker    string
	MaxKeys   int
	// IsTruncated is true if the results have been truncated because
	// there are more keys and prefixes than can fit in MaxKeys.
	// N.B. this is the opposite sense to that documented (incorrectly) in
	// http://goo.gl/YjQTc
	IsTruncated    bool
	Contents       []Key
	CommonPrefixes []string `xml:">Prefix"`
}

// The Key type represents an item stored in an S3 bucket.
type Key struct {
	Key          string
	LastModified string
	Size         int64
	// ETag gives the hex-encoded MD5 sum of the contents,
	// surrounded with double-quotes.
	ETag         string
	StorageClass string
	Owner        Owner
}

// List returns a information about objects in an S3 bucket.
//
// The prefix parameter limits the response to keys that begin with the
// specified prefix. You can use prefixes to separate a bucket into different
// groupings of keys (e.g. to get a feeling of folders).
//
// The delimited parameter causes the response to group all of the keys that
// share a common prefix up to the next delimiter to be grouped in a single
// entry within the CommonPrefixes field.
//
// The marker parameter specifies the key to start with when listing objects
// in a bucket. Amazon S3 lists objects in alphabetical order and
// will return keys alphabetically greater than the marker.
//
// The max parameter specifies how many keys + common prefixes to return in
// the response. The default is 1000.
//
// For example, given these keys in a bucket:
//
//     index.html
//     index2.html
//     photos/2006/January/sample.jpg
//     photos/2006/February/sample2.jpg
//     photos/2006/February/sample3.jpg
//     photos/2006/February/sample4.jpg
//
// Listing this bucket with delimiter set to "/" would yield the
// following result:
//
//     &ListResp{
//         Name:      "sample-bucket",
//         MaxKeys:   1000,
//         Delimiter: "/",
//         Contents:  []Key{
//             {Key: "index.html", "index2.html"},
//         },
//         CommonPrefixes: []string{
//             "photos/",
//         },
//     }
//
// Listing the same bucket with delimiter set to "/" and prefix set to
// "photos/2006/" would yield the following result:
//
//     &ListResp{
//         Name:      "sample-bucket",
//         MaxKeys:   1000,
//         Delimiter: "/",
//         Prefix:    "photos/2006/",
//         CommonPrefixes: []string{
//             "photos/2006/February/",
//             "photos/2006/January/",
//         },
//     }
// 
// See http://goo.gl/YjQTc for more details.
func (b *Bucket) List(prefix, delim, marker string, max int) (result *ListResp, err error) {
	params := map[string][]string{}
	params["prefix"] = []string{prefix}
	params["delimiter"] = []string{delim}
	params["marker"] = []string{marker}
	if max != 0 {
		params["max-keys"] = []string{strconv.FormatInt(int64(max), 10)}
	}
	result = &ListResp{}
	_, err = b.S3.query("GET", b.Name, "", params, nil, nil, result)
	return
}

// URL returns a URL for the given path. It is not signed,
// so any operations accessed this way must be available
// to anyone.
func (b *Bucket) URL(path string) string {
	if strings.HasPrefix(path, "/") {
		path = "/" + b.Name + path
	} else {
		path = "/" + b.Name + "/" + path
	}
	return b.Region.S3Endpoint + path
}

// ----------------------------------------------------------------------------
// Request dispatching logic.

func (s3 *S3) query(method, bucket, path string, params url.Values, headers http.Header, body io.Reader, resp interface{}) (hresp *http.Response, err error) {
	if debug {
		log.Printf("s3 request: method=%q; bucket=%q; path=%q resp=%T{", method, bucket, path, resp)
	}
	var endpointLocation string
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if bucket != "" {
		endpointLocation = s3.Region.S3BucketEndpoint
		if endpointLocation == "" {
			// Use the path method to address the bucket.
			endpointLocation = s3.Region.S3Endpoint
			path = "/" + bucket + path
		} else {
			for _, c := range bucket {
				if c == '/' || c == ':' || c == '@' {
					// Just in case.
					return nil, fmt.Errorf("bad S3 bucket: %q", bucket)
				}
			}
			endpointLocation = strings.Replace(endpointLocation, "${bucket}", bucket, -1)
		}
	}
	if debug {
		log.Printf("s3 endpoint: %q", endpointLocation)
	}
	endpoint, err := url.Parse(endpointLocation)
	if err != nil {
		return nil, fmt.Errorf("bad S3 endpoint URL %q: %v", endpointLocation, err)
	}
	if headers == nil {
		headers = map[string][]string{}
	}
	headers["Host"] = []string{endpoint.Host}
	headers["Date"] = []string{time.Now().In(time.UTC).Format(time.RFC1123)}
	sign(s3.Auth, method, path, params, headers)

	endpoint.Path = path
	if len(params) > 0 {
		endpoint.RawQuery = params.Encode()
	}

	req := http.Request{
		URL:        endpoint,
		Method:     method,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Close:      true,
		Header:     headers,
	}

	if body != nil {
		req.Body = ioutil.NopCloser(body)
	}

	if v, ok := headers["Content-Length"]; ok {
		req.ContentLength, _ = strconv.ParseInt(v[0], 10, 64)
		delete(headers, "Content-Length")
	}

	r, err := http.DefaultClient.Do(&req)
	if err != nil {
		return nil, err
	}
	if debug {
		dump, _ := httputil.DumpResponse(r, true)
		log.Printf("} -> %s\n", dump)
	}
	if r.StatusCode != 200 && r.StatusCode != 204 {
		return nil, buildError(r)
	}
	if resp != nil {
		err = xml.NewDecoder(r.Body).Decode(resp)
		r.Body.Close()
	}
	return r, err
}

// Error represents an error in an operation with S3.
type Error struct {
	StatusCode int    // HTTP status code (200, 403, ...)
	Code       string // EC2 error code ("UnsupportedOperation", ...)
	Message    string // The human-oriented error message
	BucketName string
	RequestId  string
	HostId     string
}

func (e *Error) Error() string {
	return e.Message
}

func buildError(r *http.Response) error {
	if debug {
		log.Printf("got error (status code %v)", r.StatusCode)
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("\tread error: %v", err)
		} else {
			log.Printf("\tdata:\n%s\n\n", data)
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	}

	err := Error{}
	// TODO return error if Unmarshal fails?
	xml.NewDecoder(r.Body).Decode(&err)
	r.Body.Close()
	err.StatusCode = r.StatusCode
	if err.Message == "" {
		err.Message = r.Status
	}
	if debug {
		log.Printf("err: %#v\n", err)
	}
	return &err
}
