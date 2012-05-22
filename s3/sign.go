package s3

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"launchpad.net/goamz/aws"
	"log"
	"sort"
	"strings"
)

var b64 = base64.StdEncoding

// ----------------------------------------------------------------------------
// S3 signing (http://goo.gl/G1LrK)

var s3ParamsToSign = map[string]bool{
	"acl":                          true,
	"location":                     true,
	"logging":                      true,
	"notification":                 true,
	"partNumber":                   true,
	"policy":                       true,
	"requestPayment":               true,
	"torrent":                      true,
	"uploadId":                     true,
	"uploads":                      true,
	"versionId":                    true,
	"versioning":                   true,
	"versions":                     true,
	"response-content-type":        true,
	"response-content-language":    true,
	"response-expires":             true,
	"response-cache-control":       true,
	"response-content-disposition": true,
	"response-content-encoding":    true,
}

func sign(auth aws.Auth, method, path string, params, headers map[string][]string) {
	var host, md5, ctype, date, xamz string
	var xamzDate bool
	var sarray []string
	for k, v := range headers {
		k = strings.ToLower(k)
		switch k {
		case "host":
			host = v[0]
		case "content-md5":
			md5 = v[0]
		case "content-type":
			ctype = v[0]
		case "date":
			if !xamzDate {
				date = v[0]
			}
		default:
			if strings.HasPrefix(k, "x-amz-") {
				vall := strings.Join(v, ",")
				sarray = append(sarray, k+":"+vall)
				if k == "x-amz-date" {
					xamzDate = true
					date = ""
				}
			}
		}
	}
	if len(sarray) > 0 {
		sort.StringSlice(sarray).Sort()
		xamz = strings.Join(sarray, "\n") + "\n"
	}

	colon := strings.Index(host, ":")
	if colon != -1 {
		host = host[:colon]
	}

	if strings.HasSuffix(host, ".amazonaws.com") {
		parts := strings.Split(host, ".")
		// -3 also strips out .s3. (or .s3-us-west-1., etc)
		bucket := strings.Join(parts[:len(parts)-3], ".")
		if bucket != "" {
			path = "/" + bucket + path
		}
	} else if strings.HasSuffix(host, ".dreamhost.com") {
		parts := strings.Split(host, ".")
		// -3 also strips out .objects.
		bucket := strings.Join(parts[:len(parts)-3], ".")
		if bucket != "" {
			path = "/" + bucket + path
		}
	} else {
		path = "/" + host + path
	}

	sarray = sarray[0:0]
	for k, v := range params {
		if s3ParamsToSign[k] {
			for _, vi := range v {
				if vi == "" {
					sarray = append(sarray, k)
				} else {
					// "When signing you do not encode these values."
					sarray = append(sarray, k+"="+vi)
				}
			}
		}
	}
	if len(sarray) > 0 {
		sort.StringSlice(sarray).Sort()
		path = path + "?" + strings.Join(sarray, "&")
	}

	payload := method + "\n" + md5 + "\n" + ctype + "\n" + date + "\n" + xamz + path
	if debug {
		log.Printf("Signature payload: %q\n", payload)
	}
	hash := hmac.New(sha1.New, []byte(auth.SecretKey))
	hash.Write([]byte(payload))
	signature := make([]byte, b64.EncodedLen(hash.Size()))
	b64.Encode(signature, hash.Sum(nil))

	headers["Authorization"] = []string{"AWS " + auth.AccessKey + ":" + string(signature)}
	if debug {
		log.Printf("Authorization header: %q", headers["Authorization"][0])
	}
}
