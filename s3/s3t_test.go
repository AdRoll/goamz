package s3_test

import (
	"crypto/md5"
	"encoding/hex"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"launchpad.net/goamz/s3/s3test"
	. "launchpad.net/gocheck"
	"strings"
)

// SuiteT defines tests to run against the s3test server
type SuiteT struct {
	i SI
}

var _ = Suite(&SuiteT{})

func (s *SuiteT) SetUpSuite(c *C) {
	srv, err := s3test.NewServer()
	c.Assert(err, IsNil)
	c.Assert(srv, NotNil)

	s.i.s3 = s3.New(
		aws.Auth{},
		aws.Region{S3Endpoint: srv.URL()},
	)
}

func (s *SuiteT) TestBasicFunctionality(c *C) {
	s.i.TestBasicFunctionality(c)
}

func (s *SuiteT) TestGetNotFound(c *C) {
	s.i.TestGetNotFound(c)
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
		Marker: objectNames[0],
		Contents: keys(objectNames[1:]...),
	},
	{
		Marker: objectNames[0]+"a",
		Contents: keys(objectNames[1:]...),
	},
	{
		Marker: "z",
	},

	// limited results.
	{
		MaxKeys: 2,
		Contents: keys(objectNames[0:2]...),
		IsTruncated: true,
	},
	{
		MaxKeys: 2,
		Marker: objectNames[0],
		Contents: keys(objectNames[1:3]...),
		IsTruncated: true,
	},
	{
		MaxKeys: 2,
		Marker: objectNames[len(objectNames)-2],
		Contents: keys(objectNames[len(objectNames)-1:]...),
	},

	// with delimiter
	{
		Delimiter: "/",
		CommonPrefixes: []string{"photos/", "test/"},
		Contents: keys("index.html", "index2.html"),
	},
	{
		Delimiter: "/",
		Prefix: "photos/2006/",
		CommonPrefixes: []string{"photos/2006/February/", "photos/2006/January/"},
	},
	{
		Delimiter: "/",
		MaxKeys: 1,
		Contents: keys("index.html"),
		IsTruncated: true,
	},
	{
		Delimiter: "/",
		MaxKeys: 1,
		Marker: "index2.html",
		CommonPrefixes: []string{"photos/"},
		IsTruncated: true,
	},
	{
		Delimiter: "/",
		MaxKeys: 1,
		Marker: "photos/",
		CommonPrefixes: []string{"test/"},
		IsTruncated: false,
	},
	{
		Delimiter: "Feb",
		CommonPrefixes: []string{"photos/2006/Feb"},
		Contents: keys("index.html", "index2.html", "photos/2006/January/sample.jpg", "test/bar", "test/foo"),
	},
}

func (s *SuiteT) TestBucketList(c *C) {
	b := s.i.Bucket(testBucket)
	err := b.PutBucket(s3.Private)
	c.Assert(err, IsNil)

	objData := make(map[string][]byte)
	for i, path := range objectNames {
		data := []byte(strings.Repeat("a", i))
		err := b.Put(path, data, "test/plain", s3.Private)
		c.Assert(err, IsNil)
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

func checksum(data []byte) string {
	sum := md5.New()
	sum.Write(data)
	return hex.EncodeToString(sum.Sum(nil))
}

func checkContents(c *C, contents []s3.Key, data map[string][]byte, expected []s3.Key) {
	c.Assert(contents, HasLen, len(expected))
	for i, k := range contents {
		c.Check(k.Key, Equals, expected[i].Key)
		// TODO mtime
		c.Check(k.Size, Equals, int64(len(data[k.Key])))
		c.Check(k.ETag, Equals, checksum(data[k.Key]))
	}
}
