package s3

import "encoding/xml"

const AllUsersUri = "http://acs.amazonaws.com/groups/global/AllUsers"

type Grantee struct {
	XMLName     xml.Name `xml:"Grantee"`
	Type        string   `xml:"type,attr"`
	URI         string   `xml:"URI"`
	ID          string   `xml:"ID"`
	DisplayName string   `xml:"DisplayName"`
}

type Grant struct {
	XMLName    xml.Name `xml:"Grant"`
	Grantee    []Grantee
	Permission string `xml:"Permission"`
}
type AccessControlListGrants struct {
	XMLName xml.Name `xml:"AccessControlList"`
	Grant   []Grant
}

type AclOwner struct {
	XMLName     xml.Name `xml:"Owner"`
	ID          string   `xml:"ID"`
	DisplayName string   `xml:"DisplayName"`
}

type AccessControlList struct {
	Owner  AclOwner
	Grants AccessControlListGrants
}

func ParseAclFromXml(aclxml string) (AccessControlList, error) {
	b := []byte(aclxml)

	var acl AccessControlList
	err := xml.Unmarshal(b, &acl)
	if err != nil {
		return acl, err
	}

	return acl, nil
}

func GetCannedPolicyByAcl(acl AccessControlList) ACL {
	for _, grant := range acl.Grants.Grant {
		//fmt.Printf("Permission: %q\n", grant.Permission)
		//fmt.Printf("Grantee Type: %q\n", grant.Grantee[0].URI)
		for _, gratee := range grant.Grantee {
			if gratee.URI == AllUsersUri {
				return PublicRead
			}
		}
	}

	return Private
}
