package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"bitbucket.org/chintana/goscim/types"
	"bitbucket.org/chintana/goscim/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Add user to DB and return error status
func (s *SCIMServer) AddUser(body io.ReadCloser) (*types.User, error) {
	user, err := createUser("", body)
	if err != nil {
		return nil, err
	}
	c := s.db.C("users")
	err = c.Insert(user)
	return user, err
}

// Get user info
func (s *SCIMServer) GetUser(id string) (*types.User, error) {
	c := s.db.C("users")
	user := types.User{}
	err := c.Find(bson.M{"id": id}).One(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update user info
func (s *SCIMServer) UpdateUser(userId string, body io.ReadCloser) (*types.User, error) {
	user, err := createUser(userId, body)
	if err != nil {
		return nil, err
	}
	c := s.db.C("users")
	err = c.Update(bson.M{"id": userId}, user)
	if err != nil {
		fmt.Println("update call error", err)
		return nil, err
	}
	newUser, err := s.GetUser(userId)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

// Create a user data structure from incoming request body, then return a User struct that
// represent incoming user info in JSON payload
func createUser(userId string, b io.Reader) (user *types.User, err error) {
	user = new(types.User)

	if err = json.NewDecoder(b).Decode(&user); err != nil {
		return nil, err
	}
	// Generate metadata
	if userId != "" {
		user.Id = userId
	} else {
		user.Id = util.GetUUID()
	}
	created := time.Now()
	user.Meta.Created = created.Format(time.RFC3339)
	user.Meta.LastModified = user.Meta.Created
	user.Meta.Location = "http://localhost:8080/Users/" + user.Id

	return user, nil
}

// Connect to the given MongoDB instance and return a pointer to the connection. DB will be
// accessed through this pointer
func NewMongoDS(url string) *mgo.Database {
	fmt.Println("Connecting to MongoDB - ", url)
	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	return session.DB("test")
}

// Hold data for the server instance.
type SCIMServer struct {
	// Pointer used to connect to a MongoDB instance
	db *mgo.Database
}

// Handle all requests coming to /Users - GET, POST, PUT, PATCH, DELETE
func (server *SCIMServer) userHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Search for user ID, then return user details in JSON

		// Passing the UUID extracted from the URL path
		user, err := server.GetUser(r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var data []byte
		data, err = json.Marshal(user)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "http://localhost:8181/Users/"+user.Id)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	case "POST":
		// Create a new resource (user)
		// http://www.simplecloud.info/specs/draft-scim-api-01.html#create-resource
		user, err := server.AddUser(r.Body)
		if err != nil {
			// TODO: check whether it's a duplicate username and throw a 409
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m, err := json.Marshal(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", "http://localhost:8181/Users/"+user.Id)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(m)
	case "PUT":
		// Update a user, passing user ID from URL and req body
		user, err := server.UpdateUser(r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:], r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m, err := json.Marshal(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", "http://localhost:8181/Users/"+user.Id)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(m)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

// Handle all requests coming to /Groups - GET, POST, PUT, PATCH, DELETE
func (s *SCIMServer) groupHandler(w http.ResponseWriter, r *http.Request) {
}

// Handle all requests coming to /ServiceProviderConfigs - GET
func (s *SCIMServer) spConfigHandler(w http.ResponseWriter, r *http.Request) {
}

// Handle all requests coming to /Schemas - GET
func (s *SCIMServer) schemaHandler(w http.ResponseWriter, r *http.Request) {
}

// Handle all requests coming to /Bulk - POST
func (s *SCIMServer) bulkHandler(w http.ResponseWriter, r *http.Request) {
}

func (s *SCIMServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case regexp.MustCompile("^/Users").MatchString(r.URL.Path):
		s.userHandler(w, r)
	case regexp.MustCompile("^/Groups").MatchString(r.URL.Path):
		s.groupHandler(w, r)
	case regexp.MustCompile("^/ServiceProviderConfigs").MatchString(r.URL.Path):
		s.spConfigHandler(w, r)
	case regexp.MustCompile("^/Schemas").MatchString(r.URL.Path):
		s.schemaHandler(w, r)
	case regexp.MustCompile("^/Bulk").MatchString(r.URL.Path):
		s.bulkHandler(w, r)
	}

}

func main() {
	// Initialize server environment
	s := &SCIMServer{
		// database instance
		db: NewMongoDS("127.0.0.1:27017"),
	}

	http.Handle("/", s)
	log.Println("Server started http://localhost:8181")
	log.Fatal(http.ListenAndServe(":8181", nil))
}
