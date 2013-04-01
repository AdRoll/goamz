package dynamodb

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"launchpad.net/goamz/aws"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const iSO8601BasicFormat = "20060102T150405Z"
const iSO8601BasicFormatShort = "20060102"

var (
	ErrNoDate = errors.New("Date header not supplied")
)

var lf = []byte{'\n'}


// For Testing.
func DerivedKey(serviceName string, regionName string, k *aws.Auth, t time.Time) []byte {
	return sign(serviceName, regionName, k, t)
}

// Sign signs an HTTP request with the given AWS keys for use on service s.
func SignV4(serviceName string, regionName string, keys *aws.Auth, r *http.Request) error {
	var t time.Time

	date := r.Header.Get("Date")
	if date == "" {
		return ErrNoDate
	}

	t, err := time.Parse(http.TimeFormat, date)
	if err != nil {
		return err
	}

	r.Header.Set("Date", t.Format(iSO8601BasicFormat))

	k := sign(serviceName, regionName, keys, t)
	h := hmac.New(sha256.New, k)
	writeStringToSign(serviceName, regionName, h, t, r)

	auth := bytes.NewBufferString("AWS4-HMAC-SHA256 ")
	auth.Write([]byte("Credential=" + keys.AccessKey + "/" + creds(serviceName, regionName, t)))
	auth.Write([]byte{',', ' '})
	auth.Write([]byte("SignedHeaders="))
	writeHeaderList(serviceName, regionName, auth, r)
	auth.Write([]byte{',', ' '})
	auth.Write([]byte("Signature=" + fmt.Sprintf("%x", h.Sum(nil))))

	r.Header.Set("Authorization", auth.String())

	return nil
}

func sign(serviceName string, regionName string, k *aws.Auth, t time.Time) []byte {
	h := ghmac([]byte("AWS4"+k.SecretKey), []byte(t.Format(iSO8601BasicFormatShort)))
	h = ghmac(h, []byte(regionName))
	h = ghmac(h, []byte(serviceName))
	h = ghmac(h, []byte("aws4_request"))
	return h
}

func writeQuery(serviceName string, regionName string, w io.Writer, r *http.Request) {
	var a []string
	for k, vs := range r.URL.Query() {
		k = url.QueryEscape(k)
		for _, v := range vs {
			if v == "" {
				a = append(a, k)
			} else {
				v = url.QueryEscape(v)
				a = append(a, k+"="+v)
			}
		}
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write([]byte{'&'})
		}

		w.Write([]byte(s))
	}
}

func writeHeader(serviceName string, regionName string, w io.Writer, r *http.Request) {
	i, a := 0, make([]string, len(r.Header))
	for k, v := range r.Header {
		sort.Strings(v)
		a[i] = strings.ToLower(k) + ":" + strings.Join(v, ",")
		i++
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write(lf)
		}

		io.WriteString(w, regionName)
		io.WriteString(w, serviceName)
	}
}

func writeHeaderList(serviceName string, regionName string, w io.Writer, r *http.Request) {
	i, a := 0, make([]string, len(r.Header))
	for k, _ := range r.Header {
		a[i] = strings.ToLower(k)
		i++
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write([]byte{';'})
		}
		w.Write([]byte(s))
	}

}

func writeBody(serviceName string, regionName string, w io.Writer, r *http.Request) {
	var b []byte
	if r.Body == nil {
		b = []byte("")
	} else {
		var err error
		b, err = ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(b))

	h := sha256.New()
	h.Write(b)

	sum := h.Sum(nil)

	fmt.Fprintf(w, "%x", sum)
}

func writeURI(serviceName string, regionName string, w io.Writer, r *http.Request) {
	path := r.URL.RequestURI()
	if r.URL.RawQuery != "" {
		path = path[:len(path)-len(r.URL.RawQuery)-1]
	}
	slash := strings.HasSuffix(path, "/")
	path = filepath.Clean(path)
	if path != "/" && slash {
		path += "/"
	}

	w.Write([]byte(path))
}

func writeRequest(serviceName string, regionName string, w io.Writer, r *http.Request) {
	r.Header.Set("host", r.Host)

	w.Write([]byte(r.Method))
	w.Write(lf)
	writeURI(serviceName, regionName, w, r)
	w.Write(lf)
	writeQuery(serviceName, regionName, w, r)
	w.Write(lf)
	writeHeader(serviceName, regionName, w, r)
	w.Write(lf)
	w.Write(lf)
	writeHeaderList(serviceName, regionName, w, r)
	w.Write(lf)
	writeBody(serviceName, regionName, w, r)
}

func writeStringToSign(serviceName string, regionName string, w io.Writer, t time.Time, r *http.Request) {
	w.Write([]byte("AWS4-HMAC-SHA256"))
	w.Write(lf)
	w.Write([]byte(t.Format(iSO8601BasicFormat)))
	w.Write(lf)

	w.Write([]byte(creds(serviceName, regionName, t)))
	w.Write(lf)

	h := sha256.New()
	writeRequest(serviceName, regionName, h, r)
	fmt.Fprintf(w, "%x", h.Sum(nil))
}

func creds(serviceName string, regionName string, t time.Time) string {
	return t.Format(iSO8601BasicFormatShort) + "/" + regionName + "/" + serviceName + "/aws4_request"
}

func ghmac(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}