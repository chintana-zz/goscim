package types

// Main struct for holding a user according to SCIM spec 2.0
type User struct {
	Active            bool          `json:"active"`
	Addresses         []Address     `json:"addresses"`
	DisplayName       string        `json:"displayName"`
	Emails            []Email       `json:"emails"`
	ExternalId        string        `json:"externalId"`
	Groups            []Group       `json:"groups"`
	Id                string        `json:"id"`
	Ims               []Im          `json:"ims"`
	Locale            string        `json:"locale"`
	Meta              MetaT         `json:"meta"`
	Name              NameT         `json:"name"`
	NickName          string        `json:"nickName"`
	Password          string        `json:"password"`
	PhoneNumbers      []PhoneNumber `json:"phoneNumbers"`
	Photos            []Photo       `json:"photos"`
	PreferredLanguage string        `json:"preferredLanguage"`
	ProfileUrl        string        `json:"profileUrl"`
	Schemas           []string      `json:"schemas"`
	TimeZone          string        `json:"timezone"`
	Title             string        `json:"title"`
	UserName          string        `json:"userName"`
	UserType          string        `json:"userType"`
	X509Certificates  []Cert        `json:"x509certificates"`
}

type Cert struct {
	Value string `json:"value"`
}

type Photo struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type PhoneNumber struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type NameT struct {
	FamilyName      string `json:"familyName"`
	Formatted       string `json:"formatted"`
	GivenName       string `json:"givenName"`
	HonorificPrefix string `json:"honorificPrefix"`
	HonorificSuffix string `json:"honorificSuffix"`
	MiddleName      string `json:"middleName"`
}

type MetaT struct {
	Created      string `json:"created"`
	LastModified string `json:"lastModified"`
	Location     string `json:"location"`
	ResourceType string `json:"resourceType"`
	Version      string `json:"version"`
}

type Im struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Group struct {
	Ref     string `json:"ref"`
	Dispaly string `json:"display"`
	Value   string `json:"value"`
}

type Email struct {
	Primary bool   `json:"primary"`
	Type    string `json:"type"`
	Value   string `json:"value"`
}

type Address struct {
	Country       string `json:"country"`
	Formatted     string `json:"formatted"`
	Locality      string `json:"locality"`
	PostalCode    string `json:"postalCode"`
	Primary       bool   `json:"primary"`
	Region        string `json:"region"`
	StreetAddress string `json:"streetAddress"`
	Type          string `json:"type"`
}
