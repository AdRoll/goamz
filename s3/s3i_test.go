package s3_test

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	. "launchpad.net/gocheck"
)

var _ = Suite(&SI{})

type SI struct {
	SuiteI
	s3 *s3.S3
}

func (s *SI) SetUpSuite(c *C) {
	s.SuiteI.SetUpSuite(c)
	s.s3 = s3.New(s.auth, aws.USEast)
}

func (s *SI) Bucket(name string) *s3.Bucket {
	return s.s3.Bucket(name + "-" + s.s3.Auth.AccessKey)
}

const testBucket = "goamz-test-bucket"

func (s *SI) TestBasicFunctionality(c *C) {
	b := s.Bucket(testBucket)
	err := b.PutBucket(s3.PublicRead)
	c.Assert(err, IsNil)

	err = b.Put("name", []byte("yo!"), "text/plain", s3.PublicRead)
	c.Assert(err, IsNil)

	data, err := b.Get("name")
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "yo!")

	resp, err := http.Get(b.URL("name"))
	c.Assert(err, IsNil)
	data, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "yo!")

	buf := bytes.NewBufferString("hey!")
	err = b.PutReader("name2", buf, int64(buf.Len()), "text/plain", s3.PublicRead)
	c.Assert(err, IsNil)

	rc, err := b.GetReader("name2")
	c.Assert(err, IsNil)
	data, err = ioutil.ReadAll(rc)
	c.Check(err, IsNil)
	c.Check(string(data), Equals, "hey!")
	rc.Close()

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

func (s *SI) TestGetNotFound(c *C) {
	b := s.Bucket("goamz-non-existent-bucket")
	data, err := b.Get("non-existent")

	s3err, _ := err.(*s3.Error)
	c.Assert(s3err, NotNil)
	c.Assert(s3err.StatusCode, Equals, 404)
	c.Assert(s3err.Code, Equals, "NoSuchBucket")
	c.Assert(s3err.Message, Equals, "The specified bucket does not exist")
	c.Assert(data, IsNil)
}

func (s *SI) unique(name string) string {
	return name + "-" + s.auth.AccessKey
}

var allRegions = []aws.Region{
	aws.USEast,
	aws.USWest,
	aws.EUWest,
	aws.APSoutheast,
	aws.APNortheast,
}

// Communicate with all endpoints to see if they are alive.
func (s *SI) TestRegions(c *C) {
	name := s.unique("goamz-region-test")
	errs := make(chan error, len(allRegions))
	for _, region := range allRegions {
		go func(r aws.Region) {
			s := s3.New(s.auth, r)
			_, err := s.Bucket(name).Get("non-existent")
			errs <- err
		}(region)
	}
	for _ = range allRegions {
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
