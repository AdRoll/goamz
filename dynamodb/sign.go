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
func SignV4(serviceName string, regionName string, keys *aws.Auth, method, canonicalPath string, params, headers map[string][]string, payload io.Reader) error {
	var t time.Time

	date := headers["Date"][0]
	if date == "" {
		return ErrNoDate
	}

	t, err := time.Parse(http.TimeFormat, date)
	if err != nil {
		return err
	}

  // r.Header.Set("Date", t.Format(iSO8601BasicFormat)) // assume this is already done for us

	k := sign(serviceName, regionName, keys, t)
	h := hmac.New(sha256.New, k)
	writeStringToSign(serviceName, regionName, h, t, method, canonicalPath, params, headers, payload)

	auth := bytes.NewBufferString("AWS4-HMAC-SHA256 ")
	auth.Write([]byte("Credential=" + keys.AccessKey + "/" + creds(serviceName, regionName, t)))
	auth.Write([]byte{',', ' '})
	auth.Write([]byte("SignedHeaders="))
	writeHeaderList(serviceName, regionName, auth, method, canonicalPath, params, headers)
	auth.Write([]byte{',', ' '})
	auth.Write([]byte("Signature=" + fmt.Sprintf("%x", h.Sum(nil))))

	headers["Authorization"] = []string{auth.String()}

	return nil
}

func sign(serviceName string, regionName string, k *aws.Auth, t time.Time) []byte {
	h := ghmac([]byte("AWS4"+k.SecretKey), []byte(t.Format(iSO8601BasicFormatShort)))
	h = ghmac(h, []byte(regionName))
	h = ghmac(h, []byte(serviceName))
	h = ghmac(h, []byte("aws4_request"))
	return h
}

func writeQuery(serviceName string, regionName string, w io.Writer, method, canonicalPath string, params, headers map[string][]string) {
	var a []string
	for k, vs := range params {
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

func writeHeader(serviceName string, regionName string, w io.Writer, method, canonicalPath string, params, headers map[string][]string) {
	i, a := 0, make([]string, len(headers))
	for k, v := range headers {
		sort.Strings(v)
		a[i] = strings.ToLower(k) + ":" + strings.Join(v, ",")
		i++
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write(lf)
		}

		io.WriteString(w, s)
	}
}

func writeHeaderList(serviceName string, regionName string, w io.Writer, method, canonicalPath string, params, headers map[string][]string) {
	i, a := 0, make([]string, len(headers))
	for k, _ := range headers {
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

func writeBody(serviceName string, regionName string, w io.Writer, method, canonicalPath string, params, headers map[string][]string, payload io.Reader) {
	var b []byte
	if payload == nil {
		b = []byte("")
	} else {
		var err error
		b, err = ioutil.ReadAll(payload)
		if err != nil {
			panic(err)
		}
	}

  // rewindPayload, err := payload.Seek(0, 0)
  //   if err != nil {
  //     panic(err);
  //   }

	h := sha256.New()
	h.Write(b)

	sum := h.Sum(nil)

	fmt.Fprintf(w, "%x", sum)
}

func writeURI(serviceName string, regionName string, w io.Writer, method, canonicalPath string, params, headers map[string][]string) {
	path := canonicalPath
	slash := strings.HasSuffix(path, "/")
	path = filepath.Clean(path)
	if path != "/" && slash {
		path += "/"
	}

	w.Write([]byte(path))
}

func writeRequest(serviceName string, regionName string, w io.Writer, method, canonicalPath string, params, headers map[string][]string, payload io.Reader) {
	//headers["host"] = []string{r.Host} assume this is already done for us

	w.Write([]byte(method))
	w.Write(lf)
	writeURI(serviceName, regionName, w, method, canonicalPath, params, headers)
	w.Write(lf)
	writeQuery(serviceName, regionName, w, method, canonicalPath, params, headers)
	w.Write(lf)
	writeHeader(serviceName, regionName, w, method, canonicalPath, params, headers)
	w.Write(lf)
	w.Write(lf)
	writeHeaderList(serviceName, regionName, w, method, canonicalPath, params, headers)
	w.Write(lf)
	writeBody(serviceName, regionName, w, method, canonicalPath, params, headers, payload)
}

func writeStringToSign(serviceName string, regionName string, w io.Writer, t time.Time, method, canonicalPath string, params, headers map[string][]string, payload io.Reader) {
	w.Write([]byte("AWS4-HMAC-SHA256"))
	w.Write(lf)
	w.Write([]byte(t.Format(iSO8601BasicFormat)))
	w.Write(lf)

	w.Write([]byte(creds(serviceName, regionName, t)))
	w.Write(lf)

	h := sha256.New()
	writeRequest(serviceName, regionName, h, method, canonicalPath, params, headers, payload)
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
