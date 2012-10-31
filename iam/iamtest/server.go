// Package iamtest implements a fake IAM provider with the capability of
// inducing errors on any given operation, and retrospectively determining what
// operations have been carried out.
package iamtest

import (
	"encoding/xml"
	"fmt"
	"launchpad.net/goamz/iam"
	"net"
	"net/http"
	"sync"
)

type action struct {
	srv   *Server
	w     http.ResponseWriter
	req   *http.Request
	reqId string
}

// Server implements an IAM simulator for use in tests.
type Server struct {
	reqId    int
	url      string
	listener net.Listener
	users    []iam.User
	mutex    sync.Mutex
}

func NewServer() (*Server, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("cannot listen on localhost: %v", err)
	}
	srv := &Server{
		listener: l,
		url:      "http://" + l.Addr().String(),
	}
	go http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		srv.serveHTTP(w, req)
	}))
	return srv, nil
}

// Quit closes down the server.
func (srv *Server) Quit() error {
	return srv.listener.Close()
}

// URL returns a URL for the server.
func (srv *Server) URL() string {
	return srv.url
}

type xmlErrors struct {
	XMLName string `xml:"ErrorResponse"`
	Error   iam.Error
}

func (srv *Server) error(w http.ResponseWriter, err *iam.Error) {
	w.WriteHeader(err.StatusCode)
	xmlErr := xmlErrors{Error: *err}
	if e := xml.NewEncoder(w).Encode(xmlErr); e != nil {
		panic(e)
	}
}

func (srv *Server) serveHTTP(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	action := req.FormValue("Action")
	if action == "" {
		srv.error(w, &iam.Error{
			StatusCode: 400,
			Code:       "MissingAction",
			Message:    "Missing action",
		})
	}
	if a, ok := actions[action]; ok {
		reqId := fmt.Sprintf("req%0X", srv.reqId)
		srv.reqId++
		if resp, err := a(srv, w, req, reqId); err == nil {
			if err := xml.NewEncoder(w).Encode(resp); err != nil {
				panic(err)
			}
		} else {
			switch err.(type) {
			case *iam.Error:
				srv.error(w, err.(*iam.Error))
			default:
				panic(err)
			}
		}
	} else {
		srv.error(w, &iam.Error{
			StatusCode: 400,
			Code:       "InvalidAction",
			Message:    "Invalid action",
		})
	}
}

func (srv *Server) createUser(w http.ResponseWriter, req *http.Request, reqId string) (interface{}, error) {
	if err := srv.validate(req, []string{"UserName"}); err != nil {
		return nil, err
	}
	path := req.FormValue("Path")
	if path == "" {
		path = "/"
	}
	name := req.FormValue("UserName")
	for _, user := range srv.users {
		if user.Name == name {
			return nil, &iam.Error{
				StatusCode: 409,
				Code:       "EntityAlreadyExists",
				Message:    fmt.Sprintf("User with name %s already exists.", name),
			}
		}
	}
	user := iam.User{
		Id:   "USER" + reqId + "EXAMPLE",
		Arn:  fmt.Sprintf("arn:aws:iam:::123456789012:user%s%s", path, name),
		Name: name,
		Path: path,
	}
	srv.users = append(srv.users, user)
	return iam.CreateUserResp{
		RequestId: reqId,
		User:      user,
	}, nil
}

func (srv *Server) getUser(w http.ResponseWriter, req *http.Request, reqId string) (interface{}, error) {
	if err := srv.validate(req, []string{"UserName"}); err != nil {
		return nil, err
	}
	name := req.FormValue("UserName")
	index := -1
	for i, user := range srv.users {
		if user.Name == name {
			index = i
			break
		}
	}
	if index < 0 {
		return nil, &iam.Error{
			StatusCode: 404,
			Code:       "NoSuchEntity",
			Message:    fmt.Sprintf("The user with name %s cannot be found.", name),
		}
	}
	return iam.GetUserResp{RequestId: reqId, User: srv.users[index]}, nil
}

func (srv *Server) deleteUser(w http.ResponseWriter, req *http.Request, reqId string) (interface{}, error) {
	if err := srv.validate(req, []string{"UserName"}); err != nil {
		return nil, err
	}
	name := req.FormValue("UserName")
	index := -1
	for i, user := range srv.users {
		if user.Name == name {
			index = i
			break
		}
	}
	if index < 0 {
		return nil, &iam.Error{
			StatusCode: 404,
			Code:       "NoSuchEntity",
			Message:    fmt.Sprintf("The user with name %s cannot be found.", name),
		}
	}
	copy(srv.users[index:], srv.users[index+1:])
	srv.users = srv.users[:len(srv.users)-1]
	return iam.SimpleResp{RequestId: reqId}, nil
}

// Validates the presence of required request parameters.
func (srv *Server) validate(req *http.Request, required []string) error {
	for _, r := range required {
		if req.FormValue(r) == "" {
			return &iam.Error{
				StatusCode: 400,
				Code:       "InvalidParameterCombination",
				Message:    fmt.Sprintf("%s is required.", r),
			}
		}
	}
	return nil
}

var actions = map[string]func(*Server, http.ResponseWriter, *http.Request, string) (interface{}, error){
	"CreateUser": (*Server).createUser,
	"DeleteUser": (*Server).deleteUser,
	"GetUser":    (*Server).getUser,
}
