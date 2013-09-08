// package ses

// import (
// 	"fmt"
// 	"github.com/alimoeeny/goamz/aws"
// 	"github.com/alimoeeny/goamz/dynamodb"
// 	"net/http"
// 	"net/url"
// )

// type Server struct {
// 	Auth aws.Auth
// }

// func (s *Server) SendHTMLEmail(from, to, cc, subj, body string) (err error) {
// 	endPoint := "https://email.us-east-1.amazonaws.com"
// 	sescli := &http.Client{}
// 	data := url.Values{"AWSAccessKeyId": {s.Auth.AccessKey},
// 		"Action":                  {"SendEmail"},
// 		"Destination.ToAddresses": {to},
// 		"Message.Body.Text.Data":  {body},
// 		"Message.Subject.Data":    {subj},
// 	}
// 	req, err := http.NewRequest("POST", endPoint, nil)
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// 	req.PostForm = data

// 	service := dynamodb.Service{
// 		"dynamodb",
// 		aws.USEast.Name,
// 	}

// 	err = service.Sign(&s.Auth, req)

// 	resp, err := sescli.Do(req)
// 	fmt.Printf("resp: %s\n", resp)
// 	if err != nil {
// 		fmt.Printf("ERROR SENDING EMAIL: %s\n", err)
// 	}
// 	return err
// }

// Copyright 2011 Numrotron Inc.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.
//
// Developed at www.stathat.com by Patrick Crosby
// Contact us on twitter with any questions:  twitter.com/stat_hat
//
// Modified for Google App Engine by
// Brandon Thomson <bt@brandonthomson.com>

// amzses is a Go package to send emails using Amazon's Simple Email Service.
package ses

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/alimoeeny/goamz/aws"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	endpoint = "https://email.us-east-1.amazonaws.com"
)

type Server struct {
	Auth aws.Auth
}

func (s *Server) SendHTMLEmail(from, to, cc, subject, body string) (string, error) {
	data := make(url.Values)
	data.Add("Action", "SendEmail")
	data.Add("Source", from)
	data.Add("Destination.ToAddresses.member.1", to)
	data.Add("Message.Subject.Data", subject)
	//data.Add("Message.Body.Text.Data", body)
	data.Add("Message.Body.Html.Data", body)
	data.Add("AWSAccessKeyId", s.Auth.AccessKey)

	return s.sesGet(data)
}

func (s *Server) authorizationHeader(date string) []string {
	h := hmac.New(sha256.New, []uint8(s.Auth.SecretKey))
	h.Write([]uint8(date))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	auth := fmt.Sprintf("AWS3-HTTPS AWSAccessKeyId=%s, Algorithm=HmacSHA256, Signature=%s", s.Auth.AccessKey, signature)
	return []string{auth}
}

func (s *Server) sesGet(data url.Values) (string, error) {
	headers := http.Header{}

	now := time.Now().UTC()
	// date format: "Tue, 25 May 2010 21:20:27 +0000"
	date := now.Format("Mon, 02 Jan 2006 15:04:05 -0700")
	headers.Set("Date", date)

	h := hmac.New(sha256.New, []uint8(s.Auth.SecretKey))
	h.Write([]uint8(date))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	auth := fmt.Sprintf("AWS3-HTTPS AWSAccessKeyId=%s, Algorithm=HmacSHA256, Signature=%s", s.Auth.AccessKey, signature)
	headers.Set("X-Amzn-Authorization", auth)

	headers.Set("Content-Type", "application/x-www-form-urlencoded")

	body := strings.NewReader(data.Encode())
	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		return "", err
	}
	req.Header = headers

	if s.Auth.Token != "" {
		req.Header.Set("X-Amz-Security-Token", s.Auth.Token)
		//fmt.Printf("Ali: SecToken = %s \n", s.Auth.Token)
	}

	//c.Debugf("%+v", req)

	//client := urlfetch.Client(c)
	client := &http.Client{}

	r, err := client.Do(req)
	if err != nil {
		return "", err
	}

	resultbody, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	if r.StatusCode != 200 {
		return "", fmt.Errorf("error, status = %d; response = %s", r.StatusCode, resultbody)
	}

	return string(resultbody), nil
}
