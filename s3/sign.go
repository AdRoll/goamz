package s3

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"github.com/crowdmob/goamz/aws"
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
	"website":                      true,
	"delete":                       true,
}


type Param struct {
	K string
	V string
}
type SortableParams []*Param
func (p SortableParams) Len() int           { return len(p) }
func (p SortableParams) Less(i, j int) bool { return p[i].K < p[j].K }
func (p SortableParams) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p SortableParams) Join(kvSep, paramSep string, discardEmpty bool) string {
	out := ""
	for _, param := range p {
		paramStr := ""
		if discardEmpty && param.V == "" {
			paramStr += param.K
		} else {
			paramStr += param.K+kvSep+param.V
		}
		if out == "" {
			out += paramStr
		} else {
			out += paramSep+paramStr
		}
	}
	return out
}


func sign(auth aws.Auth, method, canonicalPath string, params, headers map[string][]string) {
	var md5, ctype, date, xamz string
	var xamzDate bool
	var keys, sarray []string
	xheaders := make(map[string]string)
	for k, v := range headers {
		k = strings.ToLower(k)
		switch k {
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
				keys = append(keys, k)
				xheaders[k] = strings.Join(v, ",")
				if k == "x-amz-date" {
					xamzDate = true
					date = ""
				}
			}
		}
	}
	if len(keys) > 0 {
		sort.StringSlice(keys).Sort()
		for i := range keys {
			key := keys[i]
			value := xheaders[key]
			sarray = append(sarray, key+":"+value)
		}
		xamz = strings.Join(sarray, "\n") + "\n"
	}

	expires := false
	if v, ok := params["Expires"]; ok {
		// Query string request authentication alternative.
		expires = true
		date = v[0]
		params["AWSAccessKeyId"] = []string{auth.AccessKey}
	}

	sarray = sarray[0:0]
	for k, v := range params {
		if s3ParamsToSign[k] {
			for _, vi := range v {
				sarray = append(sarray, &Param{k, vi})
			}
		}
	}
	if len(sarray) > 0 {
		sort.Sort(sarray)
		canonicalPath = canonicalPath + "?" + sarray.Join("=", "&", true)
	}

	payload := method + "\n" + md5 + "\n" + ctype + "\n" + date + "\n" + xamz + canonicalPath
	hash := hmac.New(sha1.New, []byte(auth.SecretKey))
	hash.Write([]byte(payload))
	signature := make([]byte, b64.EncodedLen(hash.Size()))
	b64.Encode(signature, hash.Sum(nil))

	if expires {
		params["Signature"] = []string{string(signature)}
	} else {
		headers["Authorization"] = []string{"AWS " + auth.AccessKey + ":" + string(signature)}
	}
	if debug {
		log.Printf("Signature payload: %q", payload)
		log.Printf("Signature: %q", signature)
	}
}
