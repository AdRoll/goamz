package s3_test

import (
	"encoding/xml"
	"launchpad.net/goamz/s3"
	. "launchpad.net/gocheck"
	"strings"
)


func (s *S) TestInitMulti(c *C) {
	testServer.Response(200, nil, InitMultiResultDump)

	b := s.s3.Bucket("sample")

	multi, err := b.InitMulti("multi", "text/plain", s3.Private)
	c.Assert(err, IsNil)

	req := testServer.WaitRequest()
	c.Assert(req.Method, Equals, "POST")
	c.Assert(req.URL.Path, Equals, "/sample/multi")
	c.Assert(req.Header["Content-Type"], DeepEquals, []string{"text/plain"})
	c.Assert(req.Header["X-Amz-Acl"], DeepEquals, []string{"private"})
	c.Assert(req.Form["uploads"], DeepEquals, []string{""})

	c.Assert(multi.UploadId, Matches, "JNbR_[A-Za-z0-9.]+QQ--")
}

func (s *S) TestPutPart(c *C) {
	headers := map[string]string{
		"ETag": `"26f90efd10d614f100252ff56d88dad8"`,
	}
	testServer.Response(200, nil, InitMultiResultDump)
	testServer.Response(200, headers, "")

	b := s.s3.Bucket("sample")

	multi, err := b.InitMulti("multi", "text/plain", s3.Private)
	c.Assert(err, IsNil)

	part, err := multi.PutPart(1, strings.NewReader("<part 1>"))
	c.Assert(err, IsNil)
	c.Assert(part.N, Equals, 1)
	c.Assert(part.Size, Equals, int64(8))
	c.Assert(part.ETag, Equals, headers["ETag"])

	testServer.WaitRequest()
	req := testServer.WaitRequest()
	c.Assert(req.Method, Equals, "PUT")
	c.Assert(req.URL.Path, Equals, "/sample/multi")
	c.Assert(req.Form.Get("uploadId"), Matches, "JNbR_[A-Za-z0-9.]+QQ--")
	c.Assert(req.Form["partNumber"], DeepEquals, []string{"1"})
	c.Assert(req.Header["Content-Length"], DeepEquals, []string{"8"})
	c.Assert(req.Header["Content-Md5"], DeepEquals, []string{"JvkO/RDWFPEAJS/1bYja2A=="})
}

func (s *S) TestListParts(c *C) {
	testServer.Response(200, nil, InitMultiResultDump)
	testServer.Response(200, nil, ListPartsResultDump1)
	testServer.Response(404, nil, NoSuchUploadErrorDump) // :-(
	testServer.Response(200, nil, ListPartsResultDump2)

	b := s.s3.Bucket("sample")

	multi, err := b.InitMulti("multi", "text/plain", s3.Private)
	c.Assert(err, IsNil)

	parts, err := multi.ListParts()
	c.Assert(err, IsNil)
	c.Assert(parts, HasLen, 3)
	c.Assert(parts[0].N, Equals, 1)
	c.Assert(parts[0].Size, Equals, int64(8))
	c.Assert(parts[0].ETag, Equals, `"26f90efd10d614f100252ff56d88dad8"`)
	c.Assert(parts[1].N, Equals, 2)
	c.Assert(parts[1].Size, Equals, int64(8))
	c.Assert(parts[1].ETag, Equals, `"b572ef59bfa4a719fdf2b3c13c583af8"`)
	c.Assert(parts[2].N, Equals, 3)
	c.Assert(parts[2].Size, Equals, int64(4))
	c.Assert(parts[2].ETag, Equals, `"50dbe110509c7952ba70a96587bd7c40"`)

	testServer.WaitRequest()
	req := testServer.WaitRequest()
	c.Assert(req.Method, Equals, "GET")
	c.Assert(req.URL.Path, Equals, "/sample/multi")
	c.Assert(req.Form.Get("uploadId"), Matches, "JNbR_[A-Za-z0-9.]+QQ--")
	c.Assert(req.Form["max-parts"], DeepEquals, []string{"1000"})

	testServer.WaitRequest() // The internal error.
	req = testServer.WaitRequest()
	c.Assert(req.Method, Equals, "GET")
	c.Assert(req.URL.Path, Equals, "/sample/multi")
	c.Assert(req.Form.Get("uploadId"), Matches, "JNbR_[A-Za-z0-9.]+QQ--")
	c.Assert(req.Form["max-parts"], DeepEquals, []string{"1000"})
	c.Assert(req.Form["part-number-marker"], DeepEquals, []string{"2"})
}

func (s *S) TestMultiComplete(c *C) {
	testServer.Response(200, nil, InitMultiResultDump)
	// Note the 200 response. Completing will hold the connection on some
	// kind of long poll, and may return a late error even after a 200.
	testServer.Response(200, nil, InternalErrorDump)
	testServer.Response(200, nil, "")

	b := s.s3.Bucket("sample")

	multi, err := b.InitMulti("multi", "text/plain", s3.Private)
	c.Assert(err, IsNil)

	err = multi.Complete([]s3.Part{{2, `"ETag2"`, 32}, {1, `"ETag1"`, 64}})
	c.Assert(err, IsNil)

	testServer.WaitRequest()
	req := testServer.WaitRequest()
	c.Assert(req.Method, Equals, "POST")
	c.Assert(req.URL.Path, Equals, "/sample/multi")
	c.Assert(req.Form.Get("uploadId"), Matches, "JNbR_[A-Za-z0-9.]+QQ--")

	var payload struct{
		XMLName xml.Name
		Part []struct{
			PartNumber int
			ETag string
		}
	}

	dec := xml.NewDecoder(req.Body)
	err = dec.Decode(&payload)
	c.Assert(err, IsNil)

	c.Assert(payload.XMLName.Local, Equals, "CompleteMultipartUpload")
	c.Assert(len(payload.Part), Equals, 2)
	c.Assert(payload.Part[0].PartNumber, Equals, 1)
	c.Assert(payload.Part[0].ETag, Equals, `"ETag1"`)
	c.Assert(payload.Part[1].PartNumber, Equals, 2)
	c.Assert(payload.Part[1].ETag, Equals, `"ETag2"`)
}

func (s *S) TestMultiAbort(c *C) {
	testServer.Response(200, nil, InitMultiResultDump)
	testServer.Response(200, nil, "")

	b := s.s3.Bucket("sample")

	multi, err := b.InitMulti("multi", "text/plain", s3.Private)
	c.Assert(err, IsNil)

	err = multi.Abort()
	c.Assert(err, IsNil)

	testServer.WaitRequest()
	req := testServer.WaitRequest()
	c.Assert(req.Method, Equals, "DELETE")
	c.Assert(req.URL.Path, Equals, "/sample/multi")
	c.Assert(req.Form.Get("uploadId"), Matches, "JNbR_[A-Za-z0-9.]+QQ--")
}

func (s *S) TestListMulti(c *C) {
	testServer.Response(200, nil, ListMultiResultDump)

	b := s.s3.Bucket("sample")

	multis, prefixes, err := b.ListMulti("", "/")
	c.Assert(err, IsNil)
	c.Assert(prefixes, DeepEquals, []string{"a/", "b/"})
	c.Assert(multis, HasLen, 2)
	c.Assert(multis[0].Key, Equals, "multi1")
	c.Assert(multis[0].UploadId, Equals, "iUVug89pPvSswrikD")
	c.Assert(multis[1].Key, Equals, "multi2")
	c.Assert(multis[1].UploadId, Equals, "DkirwsSvPp98guVUi")

	req := testServer.WaitRequest()
	c.Assert(req.Method, Equals, "GET")
	c.Assert(req.URL.Path, Equals, "/sample/")
	c.Assert(req.Form["uploads"], DeepEquals, []string{""})
	c.Assert(req.Form["prefix"], DeepEquals, []string{""})
	c.Assert(req.Form["delimiter"], DeepEquals, []string{"/"})
	c.Assert(req.Form["max-uploads"], DeepEquals, []string{"1000"})
}
