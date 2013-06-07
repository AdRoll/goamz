/***** BEGIN LICENSE BLOCK *****
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this file,
# You can obtain one at http://mozilla.org/MPL/2.0/.
#
# The Initial Developer of the Original Code is the Mozilla Foundation.
# Portions created by the Initial Developer are Copyright (C) 2012
# the Initial Developer. All Rights Reserved.
#
# Contributor(s):
#   Ben Bangert (bbangert@mozilla.com)
#
# ***** END LICENSE BLOCK *****/

package cloudwatch

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/crowdmob/goamz/aws"
	"sort"
	"strings"
)

var b64 = base64.StdEncoding

func v2Sign(auth aws.Auth, method, path string, rp RequestParams, host string) (params Params) {
	params = make(map[string]string)
	for k, v := range rp.Params() {
		params[k] = v
	}
	params["AWSAccessKeyId"] = auth.AccessKey
	params["SignatureVersion"] = "2"
	params["SignatureMethod"] = "HmacSHA256"

	var sarray []string
	for k, v := range rp.Params() {
		sarray = append(sarray, aws.Encode(k)+"="+aws.Encode(v))
	}
	sort.StringSlice(sarray).Sort()
	joined := strings.Join(sarray, "&")
	payload := method + "\n" + host + "\n" + path + "\n" + joined
	hash := hmac.New(sha256.New, []byte(auth.SecretKey))
	hash.Write([]byte(payload))
	signature := make([]byte, b64.EncodedLen(hash.Size()))
	b64.Encode(signature, hash.Sum(nil))

	params["Signature"] = string(signature)
	return
}
