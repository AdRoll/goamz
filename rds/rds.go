package rds

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/hailocab/goamz/aws"
	"log"
	"net/http"
	"net/http/httputil"
	"sort"
	"strconv"
	"strings"
	"time"
)

const debug = false

// The RDS type encapsulates operations within a specific EC2 region.
type RDS struct {
	aws.Auth
	aws.Region
	private byte // Reserve the right of using private data.
}

// New creates a new RDS Client.
func New(auth aws.Auth, region aws.Region) *RDS {
	return &RDS{auth, region, 0}
}
