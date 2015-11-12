package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"net/http"
	"os"
	"time"
)

type User struct {
	Active            bool
	Addresses         []Address
	DisplayName       string
	Emails            []Email
	ExternalId        string
	Groups            []Group
	Id                string
	Ims               []Im
	Locale            string
	Meta              MetaT
	Name              NameT
	NickName          string
	Password          string
	PhoneNumbers      []PhoneNumber
	Photos            []Photo
	PreferredLanguage string
	ProfileUrl        string
	Schemas           []string
	TimeZone          string
	Title             string
	UserName          string
	UserType          string
	X509Certificates  []Cert
}

type Cert struct {
	Value string
}

type Photo struct {
	Type  string
	Value string
}

type PhoneNumber struct {
	Type  string
	Value string
}

type NameT struct {
	FamilyName      string
	Formatted       string
	GivenName       string
	HonorificPrefix string
	HonorificSuffix string
	MiddleName      string
}

type MetaT struct {
	Created      string
	LastModified string
	Location     string
	ResourceType string
	Version      string
}

type Im struct {
	Type  string
	Value string
}

type Group struct {
	Ref     string
	Dispaly string
	Value   string
}

type Email struct {
	Primary bool
	Type    string
	value   string
}

type Address struct {
	Country       string
	Formatted     string
	Locality      string
	PostalCode    string
	Primary       bool
	Region        string
	StreetAddress string
	Type          string
}

// Save the user to DB
func (user *User) save() (bool, error) {
	fmt.Println("Saving user...")
	session, err := mgo.Dial("127.0.0.1:27017")
	if err != nil {
		return false, err
	}
	defer session.Close()

	c := session.DB("test").C("users")
	err = c.Insert(user)
	if err != nil {
		return false, err
	}
	return true, nil
}

func getUUID() string {
	f, _ := os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
	b := make([]byte, 16)
	f.Read(b)
	f.Close()
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// Add a user to the system
func addUser(w http.ResponseWriter, r *http.Request) {
	ud := json.NewDecoder(r.Body)
	var user = new(User)
	err := ud.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Generate metadata before saving
	user.Id = getUUID()
	created := time.Now()
	user.Meta.Created = created.Format(time.RFC3339)
	user.Meta.LastModified = user.Meta.Created
	user.Meta.Location = "http://localhost:8080/Users/" + user.Id

	go user.save()

	w.Header().Set("Location", "http://localhost:8080/Users/"+user.Id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	m, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(m)
}

// Get user from DB
func getUser(w http.ResponseWriter, r *http.Request) {
	// get user ID
	userId := r.RequestURI[7:]
	fmt.Println(userId)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		getUser(w, r)
	case r.Method == "POST":
		addUser(w, r)
	case r.Method == "PUT":
	case r.Method == "PATCH":
	case r.Method == "DELETE":
	default:
		// Any other method should result in an error
		http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/Users/", userHandler)
	http.ListenAndServe(":8181", nil)
}
