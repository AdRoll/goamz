package s3

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"io"
	"sort"
	"strconv"
)

// Multi represents an unfinished multipart upload.
//
// Multipart uploads allow sending big objects in smaller chunks.
// After all parts have been sent, the upload must be explicitly
// completed by calling Complete with the list of parts.
//
// See http://goo.gl/vJfTG for an overview of multipart uploads.
type Multi struct {
	Bucket   *Bucket
	Key      string
	UploadId string
}

var listMultiMax = 1000

type listMultiResp struct {
	NextKeyMarker      string
	NextUploadIdMarker string
	IsTruncated        bool
	Upload             []Multi
	CommonPrefixes     []string `xml:"CommonPrefixes>Prefix"`
}

// ListMulti returns the list of unfinished multipart uploads in b.
//
// The prefix parameter limits the response to keys that begin with the
// specified prefix. You can use prefixes to separate a bucket into different
// groupings of keys (to get the feeling of folders, for example).
//
// The delim parameter causes the response to group all of the keys that
// share a common prefix up to the next delimiter in a single entry within
// the CommonPrefixes field. You can use delimiters to separate a bucket
// into different groupings of keys, similar to how folders would work.
//
// See http://goo.gl/ePioY for details.
func (b *Bucket) ListMulti(prefix, delim string) (multis []*Multi, prefixes []string, err error) {
	params := map[string][]string{
		"uploads":     {""},
		"max-uploads": {strconv.FormatInt(int64(listMultiMax), 10)},
		"prefix":      {prefix},
		"delimiter":   {delim},
	}
	for attempt := attempts.Start(); attempt.Next(); {
		req := &request{
			method: "GET",
			bucket: b.Name,
			params: params,
		}
		var resp listMultiResp
		err := b.S3.query(req, &resp)
		if shouldRetry(err) && attempt.HasNext() {
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		for i := range resp.Upload {
			multi := &resp.Upload[i]
			multi.Bucket = b
			multis = append(multis, multi)
		}
		prefixes = append(prefixes, resp.CommonPrefixes...)
		if !resp.IsTruncated {
			return multis, prefixes, nil
		}
		params["key-marker"] = []string{resp.NextKeyMarker}
		params["upload-id-marker"] = []string{resp.NextUploadIdMarker}
		attempt = attempts.Start() // Last request worked.
	}
	panic("unreachable")
}

// InitMulti initializes a new multipart upload at the provided
// key inside b and returns a value for manipulating it.
//
// See http://goo.gl/XP8kL for details.
func (b *Bucket) InitMulti(key string, contType string, perm ACL) (*Multi, error) {
	headers := map[string][]string{
		"Content-Type":   {contType},
		"Content-Length": {"0"},
		"x-amz-acl":      {string(perm)},
	}
	params := map[string][]string{
		"uploads": {""},
	}
	req := &request{
		method:  "POST",
		bucket:  b.Name,
		path:    key,
		headers: headers,
		params:  params,
	}
	var err error
	var resp struct {
		UploadId string `xml:"UploadId"`
	}
	for attempt := attempts.Start(); attempt.Next(); {
		err = b.S3.query(req, &resp)
		if !shouldRetry(err) {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return &Multi{Bucket: b, Key: key, UploadId: resp.UploadId}, nil
}

// PutPart sends part n of the multipart upload, reading all the content from r.
// Each part, except for the last one, must be at least 5MB in size.
//
// See http://goo.gl/pqZer for details.
func (m *Multi) PutPart(n int, r io.ReadSeeker) (Part, error) {
	length, b64md5, err := seekerInfo(r)
	if err != nil {
		return Part{}, err
	}
	headers := map[string][]string{
		"Content-Length": {strconv.FormatInt(length, 10)},
		"Content-MD5":    {b64md5},
	}
	params := map[string][]string{
		"uploadId":   {m.UploadId},
		"partNumber": {strconv.FormatInt(int64(n), 10)},
	}
	req := &request{
		method:  "PUT",
		bucket:  m.Bucket.Name,
		path:    m.Key,
		headers: headers,
		params:  params,
		payload: r,
	}
	err = m.Bucket.S3.prepare(req)
	if err != nil {
		return Part{}, err
	}
	for attempt := attempts.Start(); attempt.Next(); {
		_, err := r.Seek(0, 0)
		if err != nil {
			return Part{}, err
		}
		resp, err := m.Bucket.S3.run(req, nil)
		if shouldRetry(err) && attempt.HasNext() {
			continue
		}
		if err != nil {
			return Part{}, err
		}
		etag := resp.Header.Get("ETag")
		if etag == "" {
			return Part{}, errors.New("part upload succeeded with no ETag")
		}
		return Part{n, etag, length}, nil
	}
	panic("unreachable")
}

func seekerInfo(r io.ReadSeeker) (length int64, b64md5 string, err error) {
	_, err = r.Seek(0, 0)
	if err != nil {
		return 0, "", err
	}
	digest := md5.New()
	length, err = io.Copy(digest, r)
	if err != nil {
		return 0, "", err
	}
	b64md5 = base64.StdEncoding.EncodeToString(digest.Sum(nil))
	return length, b64md5, nil
}

type Part struct {
	N    int `xml:"PartNumber"`
	ETag string
	Size int64
}

type listPartsResp struct {
	NextPartNumberMarker string
	IsTruncated          bool
	Part                 []Part
}

var listPartsMax = 1000

// ListParts returns the list of previously uploaded parts in m.
//
// See http://goo.gl/ePioY for details.
func (m *Multi) ListParts() ([]Part, error) {
	params := map[string][]string{
		"uploadId":  {m.UploadId},
		"max-parts": {strconv.FormatInt(int64(listPartsMax), 10)},
	}
	var parts []Part
	for attempt := attempts.Start(); attempt.Next(); {
		req := &request{
			method: "GET",
			bucket: m.Bucket.Name,
			path:   m.Key,
			params: params,
		}
		var resp listPartsResp
		err := m.Bucket.S3.query(req, &resp)
		if shouldRetry(err) && attempt.HasNext() {
			continue
		}
		if err != nil {
			return nil, err
		}
		parts = append(parts, resp.Part...)
		if !resp.IsTruncated {
			return parts, nil
		}
		params["part-number-marker"] = []string{resp.NextPartNumberMarker}
		attempt = attempts.Start() // Last request worked.
	}
	panic("unreachable")
}

type completeUpload struct {
	XMLName xml.Name      `xml:"CompleteMultipartUpload"`
	Parts   completeParts `xml:"Part"`
}

type completePart struct {
	PartNumber int
	ETag       string
}

type completeParts []completePart

func (p completeParts) Len() int           { return len(p) }
func (p completeParts) Less(i, j int) bool { return p[i].PartNumber < p[j].PartNumber }
func (p completeParts) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Complete assembles the given previously uploaded parts into the
// final object. This operation may take several minutes.
//
// See http://goo.gl/2Z7Tw for details.
func (m *Multi) Complete(parts []Part) error {
	params := map[string][]string{
		"uploadId": {m.UploadId},
	}
	c := completeUpload{}
	for _, p := range parts {
		c.Parts = append(c.Parts, completePart{p.N, p.ETag})
	}
	sort.Sort(c.Parts)
	data, err := xml.Marshal(&c)
	if err != nil {
		return err
	}
	for attempt := attempts.Start(); attempt.Next(); {
		req := &request{
			method:  "POST",
			bucket:  m.Bucket.Name,
			path:    m.Key,
			params:  params,
			payload: bytes.NewReader(data),
		}
		err := m.Bucket.S3.query(req, nil)
		if shouldRetry(err) && attempt.HasNext() {
			continue
		}
		return err
	}
	panic("unreachable")
}

// Abort deletes an unifinished multipart upload and any previously
// uploaded parts for it.
//
// After a multipart upload is aborted, no additional parts can be
// uploaded using it. However, if any part uploads are currently in
// progress, those part uploads might or might not succeed. As a result,
// it might be necessary to abort a given multipart upload multiple
// times in order to completely free all storage consumed by all parts.
//
// NOTE: If the described scenario happens to you, please report back to
// the goamz authors with details. In the future such retrying should be
// handled internally, but it's not clear what happens precisely (Is an
// error returned? Is the issue completely undetectable?).
//
// See http://goo.gl/dnyJw for details.
func (m *Multi) Abort() error {
	params := map[string][]string{
		"uploadId":   {m.UploadId},
	}
	for attempt := attempts.Start(); attempt.Next(); {
		req := &request{
			method:  "DELETE",
			bucket:  m.Bucket.Name,
			path:    m.Key,
			params:  params,
		}
		err := m.Bucket.S3.query(req, nil)
		if shouldRetry(err) && attempt.HasNext() {
			continue
		}
		return err
	}
	panic("unreachable")
}
