package rds

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_AvailabilityZone.html
type AvailabilityZone struct {
	Name                   string
	ProvisionedIopsCapable bool
}

// DBInstance encapsulates an instance of a Database
//
// See http://goo.gl/rQFpAe for more details.
type DBInstance struct {
	AllocatedStorage                      int                          `xml:"AllocatedStorage"`                      // Specifies the allocated storage size specified in gigabytes.
	AutoMinorVersionUpgrade               bool                         `xml:"AutoMinorVersionUpgrade"`               // Indicates that minor version patches are applied automatically.
	AvailabilityZone                      string                       `xml:"AvailabilityZone"`                      // Specifies the name of the Availability Zone the DB instance is located in.
	BackupRetentionPeriod                 int                          `xml:"BackupRetentionPeriod"`                 // Specifies the number of days for which automatic DB snapshots are retained.
	CharacterSetName                      string                       `xml:"CharacterSetName"`                      // If present, specifies the name of the character set that this instance is associated with.
	DBInstanceClass                       string                       `xml:"DBInstanceClass"`                       // Contains the name of the compute and memory capacity class of the DB instance.
	DBInstanceIdentifier                  string                       `xml:"DBInstanceIdentifier"`                  // Contains a user-supplied database identifier. This is the unique key that identifies a DB instance.
	DBInstanceStatus                      string                       `xml:"DBInstanceStatus"`                      // Specifies the current state of this database.
	DBName                                string                       `xml:"DBName"`                                // The meaning of this parameter differs according to the database engine you use.
	DBParameterGroups                     []DBParameterGroupStatus     `xml:"DBParameterGroups"`                     // Provides the list of DB parameter groups applied to this DB instance.
	DBSecurityGroups                      []DBSecurityGroupMembership  `xml:"DBSecurityGroups"`                      // Provides List of DB security group elements containing only DBSecurityGroup.Name and DBSecurityGroup.Status subelements.
	DBSubnetGroup                         DBSubnetGroup                `xml:"DBSubnetGroup"`                         // Specifies information on the subnet group associated with the DB instance, including the name, description, and subnets in the subnet group.
	Endpoint                              Endpoint                     `xml:"Endpoint"`                              // Specifies the connection endpoint.
	Engine                                string                       `xml:"Engine"`                                // Provides the name of the database engine to be used for this DB instance.
	EngineVersion                         string                       `xml:"EngineVersion"`                         // Indicates the database engine version.
	InstanceCreateTime                    string                       `xml:"InstanceCreateTime"`                    // Provides the date and time the DB instance was created.
	Iops                                  int                          `xml:"Iops"`                                  // Specifies the Provisioned IOPS (I/O operations per second) value.
	LatestRestorableTime                  string                       `xml:"LatestRestorableTime"`                  // Specifies the latest time to which a database can be restored with point-in-time restore.
	LicenseModel                          string                       `xml:"LicenseModel"`                          // License model information for this DB instance.
	MasterUsername                        string                       `xml:"MasterUsername"`                        // Contains the master username for the DB instance.
	MultiAZ                               bool                         `xml:"MultiAZ"`                               // Specifies if the DB instance is a Multi-AZ deployment.
	OptionGroupMemberships                []OptionGroupMembership      `xml:"OptionGroupMemberships"`                // Provides the list of option group memberships for this DB instance.
	PendingModifiedValues                 PendingModifiedValues        `xml:"PendingModifiedValues"`                 // Specifies that changes to the DB instance are pending. This element is only included when changes are pending. Specific changes are identified by subelements.
	PreferredBackupWindow                 string                       `xml:"PreferredBackupWindow"`                 // Specifies the daily time range during which automated backups are created if automated backups are enabled, as determined by the BackupRetentionPeriod.
	PreferredMaintenanceWindow            string                       `xml:"PreferredMaintenanceWindow"`            // Specifies the weekly time range (in UTC) during which system maintenance can occur.
	PubliclyAccessible                    bool                         `xml:"PubliclyAccessible"`                    // Specifies the accessibility options for the DB instance. A value of true specifies an Internet-facing instance with a publicly resolvable DNS name, which resolves to a public IP address. A value of false specifies an internal instance with a DNS name that resolves to a private IP address.
	ReadReplicaDBInstanceIdentifiers      []string                     `xml:"ReadReplicaDBInstanceIdentifiers"`      // Contains one or more identifiers of the read replicas associated with this DB instance.
	ReadReplicaSourceDBInstanceIdentifier string                       `xml:"ReadReplicaSourceDBInstanceIdentifier"` // Contains the identifier of the source DB instance if this DB instance is a read replica.
	SecondaryAvailabilityZone             string                       `xml:"SecondaryAvailabilityZone"`             // If present, specifies the name of the secondary Availability Zone for a DB instance with multi-AZ support.
	StatusInfos                           []DBInstanceStatusInfo       `xml:"StatusInfos"`                           // The status of a read replica. If the instance is not a read replica, this will be blank.
	VpcSecurityGroups                     []VpcSecurityGroupMembership `xml:"VpcSecurityGroups"`                     // Provides List of VPC security group elements that the DB instance belongs to.
}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_DBParameterGroup.html
type DBParameterGroup struct {
	Name        string `xml:"DBParameterGroupName"`
	Description string `xml:"Description"`
	Family      string `xml:"DBParameterGroupFamily"`
}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_DBParameterGroupStatus.html
type DBParameterGroupStatus struct {
	Name   string `xml:"DBParameterGroupName"`
	Status string `xml:"ParameterApplyStatus"`
}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_DBSecurityGroup.html
type DBSecurityGroup struct {
	Name              string             `xml:"DBSecurityGroupName"`
	Description       string             `xml:"DBSecurityGroupDescription"`
	EC2SecurityGroups []EC2SecurityGroup `xml:"EC2SecurityGroups"`
	IPRanges          []IPRange          `xml:"IPRanges"`
	OwnerId           string             `xml:"OwnerId"`
	VpcId             string             `xml:"VpcId"`
}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_DBSecurityGroupMembership.html
type DBSecurityGroupMembership struct {
	Name   string `xml:"DBSecurityGroupName"`
	Status string `xml:"Status"`
}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_DBInstanceStatusInfo.html
type DBInstanceStatusInfo struct {
	Message    string `xml:"Message"`    // Details of the error if there is an error for the instance. If the instance is not in an error state, this value is blank.
	Normal     bool   `xml:"Normal"`     // Boolean value that is true if the instance is operating normally, or false if the instance is in an error state.
	Status     string `xml:"Status"`     // Status of the DB instance. For a StatusType of read replica, the values can be replicating, error, stopped, or terminated.
	StatusType string `xml:"StatusType"` // This value is currently "read replication."
}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_DBSubnetGroup.html
type DBSubnetGroup struct {
	Name        string   `xml:"DBSubnetGroupName"`
	Description string   `xml:"DBSubnetGroupDescription"`
	Status      string   `xml:"SubnetGroupStatus"`
	Subnets     []Subnet `xml:"Subnets"`
	VpcId       string   `xml:"VpcId"`
}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_EC2SecurityGroup.html
type EC2SecurityGroup struct {
	Id      string `xml:"EC2SecurityGroupId"`
	Name    string `xml:"EC2SecurityGroupName"`
	OwnerId string `xml:"EC2SecurityGroupOwnerId"`
	Status  string `xml:"Status"`
}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_Endpoint.html
type Endpoint struct {
	Address string `xml:"Address"`
	Port    int    `xml:"Port"`
}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_IPRange.html
type IPRange struct {
	CIDRIP string `xml:"CIDRIP"`
	Status string `xml:"Status"` // Specifies the status of the IP range. Status can be "authorizing", "authorized", "revoking", and "revoked".
}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_OptionGroup.html
type OptionGroup struct {
	Name                                  string   `xml:"OptionGroupName"`
	Description                           string   `xml:"OptionGroupDescription"`
	VpcId                                 string   `xml:"VpcId"`
	AllowsVpcAndNonVpcInstanceMemberships bool     `xml:"AllowsVpcAndNonVpcInstanceMemberships"`
	EngineName                            string   `xml:"EngineName"`
	MajorEngineVersion                    string   `xml:"MajorEngineVersion"`
	Options                               []Option `xml:"Options"`
}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_OptionGroupMembership.html
type OptionGroupMembership struct {
	Name   string `xml:"OptionGroupName"` // The name of the option group that the instance belongs to
	Status string `xml:"Status"`          // The status of the option group membership, e.g. in-sync, pending, pending-maintenance, applying
}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_PendingModifiedValues.html
type PendingModifiedValues struct{}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_Subnet.html
type Subnet struct {
	Id               string           `xml:"SubnetIdentifier"`
	Status           string           `xml:"SubnetStatus"`
	AvailabilityZone AvailabilityZone `xml:"SubnetAvailabilityZone"`
}

// http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_VpcSecurityGroupMembership.html
type VpcSecurityGroupMembership struct {
	Id     string `xml:"VpcSecurityGroupId"`
	Status string `xml:"Status"`
}
