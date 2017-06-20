package s3_test

import (
	"errors"

	"github.com/AdRoll/goamz/s3"
	"gopkg.in/check.v1"
)

const xmlResponseSimple = `<?xml version="1.0" encoding="UTF-8"?>
<AccessControlPolicy xmlns="http://s3.amazonaws.com/doc/2006-03-01/">

<Owner>
	<ID>owner1</ID>
	<DisplayName>My service user</DisplayName>
</Owner>

<AccessControlList>
<Grant>
	<Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="Group">
	<URI>http://acs.amazonaws.com/groups/global/AllUsers</URI>
	</Grantee>
	<Permission>READ</Permission>
</Grant>

<Grant>
	<Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="CanonicalUser">
	<ID>owner1</ID>
	<DisplayName>My service user</DisplayName>
	</Grantee>
	<Permission>FULL_CONTROL</Permission>
</Grant>

</AccessControlList>
</AccessControlPolicy>
`

const xmlResponseExtended = `<?xml version="1.0" encoding="UTF-8"?>
<AccessControlPolicy xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
<Owner>
<ID>Owner-canonical-user-ID</ID>
<DisplayName>display-name</DisplayName>
</Owner>
<AccessControlList>
<Grant>
<Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="CanonicalUser">
<ID>Owner-canonical-user-ID</ID>
<DisplayName>display-name</DisplayName>
</Grantee>
<Permission>FULL_CONTROL</Permission>
</Grant>

<Grant>
<Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="CanonicalUser">
<ID>user1-canonical-user-ID</ID>
<DisplayName>display-name</DisplayName>
</Grantee>
<Permission>WRITE</Permission>
</Grant>

<Grant>
<Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="CanonicalUser">
<ID>user2-canonical-user-ID</ID>
<DisplayName>display-name</DisplayName>
</Grantee>
<Permission>READ</Permission>
</Grant>

<Grant>
<Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="Group">
<URI>http://acs.amazonaws.com/groups/global/AllUsers</URI>
</Grantee>
<Permission>READ</Permission>
</Grant>
<Grant>
<Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="Group">
<URI>http://acs.amazonaws.com/groups/s3/LogDelivery</URI>
</Grantee>
<Permission>WRITE</Permission>
</Grant>

</AccessControlList>
</AccessControlPolicy>`

func (s *S) TestShouldParseXmlSimpleACLResponse(c *check.C) {
	acl, err := s3.ParseAclFromXml(xmlResponseSimple)

	c.Assert(err, check.Equals, nil)
	c.Assert(acl.Owner.ID, check.Equals, "owner1")
	c.Assert(acl.Owner.DisplayName, check.Equals, "My service user")
	c.Assert(acl.Grants.Grant[0].Grantee[0].Type, check.Equals, "Group")
	c.Assert(acl.Grants.Grant[0].Permission, check.Equals, "READ")
}

func (s *S) TestShouldParseXmlExtendedACLResponse(c *check.C) {
	acl, err := s3.ParseAclFromXml(xmlResponseExtended)

	c.Assert(err, check.Equals, nil)
	c.Assert(acl.Owner.ID, check.Equals, "Owner-canonical-user-ID")
	c.Assert(acl.Owner.DisplayName, check.Equals, "display-name")
	c.Assert(acl.Grants.Grant[0].Grantee[0].Type, check.Equals, "CanonicalUser")
	c.Assert(acl.Grants.Grant[0].Permission, check.Equals, "FULL_CONTROL")
}

func (s *S) TestShouldNotGetCannedPolicyByAclWithEmptyXmlInput(c *check.C) {
	_, err := s3.ParseAclFromXml("")

	c.Assert(err, check.DeepEquals, errors.New("EOF"))
}

func (s *S) TestShouldGetCannedPolicyByAclFromXmlSimpleResponse(c *check.C) {
	acl, _ := s3.ParseAclFromXml(xmlResponseSimple)

	cannedPolicy := s3.GetCannedPolicyByAcl(acl)

	c.Assert(cannedPolicy, check.Equals, s3.ACL(s3.PublicRead))
}
