package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

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
		return
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
		return
	}
	w.Write(m)
}

// Get user from DB
func getUser(w http.ResponseWriter, r *http.Request) {
	// get user ID
	userId := r.RequestURI[7:]

	fmt.Println("Getting user...")
	session, err := mgo.Dial("127.0.0.1:27017")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer session.Close()

	c := session.DB("test").C("users")

	user := User{}
	fmt.Println("Finding user...")
	err = c.Find(bson.M{"id": userId}).One(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", "http://localhost:8080/Users/"+user.Id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	m, err := json.Marshal(user)
	fmt.Println("Marshalling user...")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(m)
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
