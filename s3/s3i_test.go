package s3_test

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	. "launchpad.net/gocheck"
	"time"
)

// AmazonServer represents an Amazon S3 server.
type AmazonServer struct {
	auth aws.Auth
}

func (s *AmazonServer) SetUp(c *C) {
	auth, err := aws.EnvAuth()
	if err != nil {
		c.Fatal(err.Error())
	}
	s.auth = auth
}

// us-east-1 classic tests
var _ = Suite(&AmazonClientSuite{Region: aws.USEast})
var _ = Suite(&AmazonDomainClientSuite{Region: aws.USEast})

// us-west-1 tests
var _ = Suite(&AmazonClientSuite{Region: aws.USWest})

// eu-west-1 tests
var _ = Suite(&AmazonClientSuite{Region: aws.EUWest})

// AmazonClientSuite tests the client against a live S3 server.
type AmazonClientSuite struct {
	aws.Region
	srv AmazonServer
	ClientTests
}

func (s *AmazonClientSuite) SetUpSuite(c *C) {
	if !*amazon {
		c.Skip("live tests against AWS disabled (no -amazon)")
	}
	s.srv.SetUp(c)
	s.s3 = s3.New(s.srv.auth, s.Region)
}

// AmazonDomainClientSuite tests the client against a live S3
// server using bucket names in the endpoint domain name rather
// than the request path.
type AmazonDomainClientSuite struct {
	aws.Region
	srv AmazonServer
	ClientTests
}

func (s *AmazonDomainClientSuite) SetUpSuite(c *C) {
	if !*amazon {
		c.Skip("live tests against AWS disabled (no -amazon)")
	}
	s.srv.SetUp(c)
	region := s.Region
	// TODO(dfc) this subsitution only works for us-east-1
	region.S3BucketEndpoint = "https://${bucket}.s3.amazonaws.com"
	s.s3 = s3.New(s.srv.auth, region)
}

// ClientTests defines integration tests designed to test the client.
// It is not used as a test suite in itself, but embedded within
// another type.
type ClientTests struct {
	s3           *s3.S3
	authIsBroken bool
}

func (s *ClientTests) Bucket(name string) *s3.Bucket {
	// Creating and deleting buckets across regions can fail due to namespace collision.
	// Use the region name to ensure the final bucket name does not clash.
	return s.s3.Bucket(fmt.Sprintf("%s-%s-%s", name, s.s3.Region.Name, s.s3.Auth.AccessKey))
}

const testBucket = "goamz-test-bucket"

func get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, err
}

func (s *ClientTests) TestBasicFunctionality(c *C) {
	b := s.Bucket(testBucket)
	err := b.PutBucket(s3.PublicRead)
	c.Assert(err, IsNil)

	err = b.Put("name", []byte("yo!"), "text/plain", s3.PublicRead)
	c.Assert(err, IsNil)

	data, err := b.Get("name")
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "yo!")

	data, err = get(b.URL("name"))
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "yo!")

	buf := bytes.NewBufferString("hey!")
	err = b.PutReader("name2", buf, int64(buf.Len()), "text/plain", s3.Private)
	c.Assert(err, IsNil)

	rc, err := b.GetReader("name2")
	c.Assert(err, IsNil)
	data, err = ioutil.ReadAll(rc)
	c.Check(err, IsNil)
	c.Check(string(data), Equals, "hey!")
	rc.Close()

	data, err = get(b.SignedURL("name2", time.Now().Add(time.Hour)))
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "hey!")

	if !s.authIsBroken {
		data, err = get(b.SignedURL("name2", time.Now().Add(-time.Hour)))
		c.Assert(err, IsNil)
		c.Assert(string(data), Matches, "(?s).*AccessDenied.*")
	}

	err = b.DelBucket()
	c.Assert(err, NotNil)

	s3err, ok := err.(*s3.Error)
	c.Assert(ok, Equals, true)
	c.Assert(s3err.Code, Equals, "BucketNotEmpty")
	c.Assert(s3err.BucketName, Equals, b.Name)
	c.Assert(s3err.Message, Equals, "The bucket you tried to delete is not empty")

	err = b.Del("name")
	c.Assert(err, IsNil)
	err = b.Del("name2")
	c.Assert(err, IsNil)

	err = b.DelBucket()
	c.Assert(err, IsNil)
}

func (s *ClientTests) TestGetNotFound(c *C) {
	b := s.Bucket("goamz-non-existent-bucket")
	data, err := b.Get("non-existent")

	s3err, _ := err.(*s3.Error)
	c.Assert(s3err, NotNil)
	c.Assert(s3err.StatusCode, Equals, 404)
	c.Assert(s3err.Code, Equals, "NoSuchBucket")
	c.Assert(s3err.Message, Equals, "The specified bucket does not exist")
	c.Assert(data, IsNil)
}

func (s *ClientTests) unique(name string) string {
	return name + "-" + s.s3.AccessKey
}

// Communicate with all endpoints to see if they are alive.
func (s *ClientTests) TestRegions(c *C) {
	name := s.unique("goamz-region-test")
	errs := make(chan error, len(aws.Regions))
	for _, region := range aws.Regions {
		go func(r aws.Region) {
			s := s3.New(s.s3.Auth, r)
			_, err := s.Bucket(name).Get("non-existent")
			errs <- err
		}(region)
	}
	for _ = range aws.Regions {
		err := <-errs
		if err != nil {
			s3_err, ok := err.(*s3.Error)
			if ok {
				c.Check(s3_err.Code, Matches, "NoSuchBucket")
			} else {
				c.Errorf("Non-S3 error: %s", err)
			}
		} else {
			c.Errorf("Test should have errored but it seems to have succeeded")
		}
	}
}

var objectNames = []string{
	"index.html",
	"index2.html",
	"photos/2006/February/sample2.jpg",
	"photos/2006/February/sample3.jpg",
	"photos/2006/February/sample4.jpg",
	"photos/2006/January/sample.jpg",
	"test/bar",
	"test/foo",
}

func keys(names ...string) []s3.Key {
	ks := make([]s3.Key, len(names))
	for i, name := range names {
		ks[i].Key = name
	}
	return ks
}

// As the ListResp specifies all the parameters to the
// request too, we use it to specify request parameters
// and expected results. The Contents field is
// used only for the key names inside it.
var listTests = []s3.ListResp{
	// normal list.
	{
		Contents: keys(objectNames...),
	},
	{
		Marker:   objectNames[0],
		Contents: keys(objectNames[1:]...),
	},
	{
		Marker:   objectNames[0] + "a",
		Contents: keys(objectNames[1:]...),
	},
	{
		Marker: "z",
	},

	// limited results.
	{
		MaxKeys:     2,
		Contents:    keys(objectNames[0:2]...),
		IsTruncated: true,
	},
	{
		MaxKeys:     2,
		Marker:      objectNames[0],
		Contents:    keys(objectNames[1:3]...),
		IsTruncated: true,
	},
	{
		MaxKeys:  2,
		Marker:   objectNames[len(objectNames)-2],
		Contents: keys(objectNames[len(objectNames)-1:]...),
	},

	// with delimiter
	{
		Delimiter:      "/",
		CommonPrefixes: []string{"photos/", "test/"},
		Contents:       keys("index.html", "index2.html"),
	},
	{
		Delimiter:      "/",
		Prefix:         "photos/2006/",
		CommonPrefixes: []string{"photos/2006/February/", "photos/2006/January/"},
	},
	{
		Delimiter:      "/",
		Prefix:         "t",
		CommonPrefixes: []string{"test/"},
	},
	{
		Delimiter:   "/",
		MaxKeys:     1,
		Contents:    keys("index.html"),
		IsTruncated: true,
	},
	{
		Delimiter:      "/",
		MaxKeys:        1,
		Marker:         "index2.html",
		CommonPrefixes: []string{"photos/"},
		IsTruncated:    true,
	},
	{
		Delimiter:      "/",
		MaxKeys:        1,
		Marker:         "photos/",
		CommonPrefixes: []string{"test/"},
		IsTruncated:    false,
	},
	{
		Delimiter:      "Feb",
		CommonPrefixes: []string{"photos/2006/Feb"},
		Contents:       keys("index.html", "index2.html", "photos/2006/January/sample.jpg", "test/bar", "test/foo"),
	},
}

func (s *ClientTests) TestBucketList(c *C) {
	b := s.Bucket(testBucket)
	err := b.PutBucket(s3.Private)
	c.Assert(err, IsNil)
	defer b.DelBucket()

	objData := make(map[string][]byte)
	for i, path := range objectNames {
		data := []byte(strings.Repeat("a", i))
		err := b.Put(path, data, "test/plain", s3.Private)
		c.Assert(err, IsNil)
		defer b.Del(path)
		objData[path] = data
	}

	for i, t := range listTests {
		c.Logf("test %d", i)
		resp, err := b.List(t.Prefix, t.Delimiter, t.Marker, t.MaxKeys)
		c.Assert(err, IsNil)
		c.Check(resp.Name, Equals, b.Name)
		c.Check(resp.Delimiter, Equals, t.Delimiter)
		c.Check(resp.IsTruncated, Equals, t.IsTruncated)
		c.Check(resp.CommonPrefixes, DeepEquals, t.CommonPrefixes)
		checkContents(c, resp.Contents, objData, t.Contents)
	}
}

func etag(data []byte) string {
	sum := md5.New()
	sum.Write(data)
	return fmt.Sprintf(`"%x"`, sum.Sum(nil))
}

func checkContents(c *C, contents []s3.Key, data map[string][]byte, expected []s3.Key) {
	c.Assert(contents, HasLen, len(expected))
	for i, k := range contents {
		c.Check(k.Key, Equals, expected[i].Key)
		// TODO mtime
		c.Check(k.Size, Equals, int64(len(data[k.Key])))
		c.Check(k.ETag, Equals, etag(data[k.Key]))
	}
}
