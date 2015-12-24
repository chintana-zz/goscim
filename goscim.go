package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"bitbucket.org/chintana/goscim/types"
	"bitbucket.org/chintana/goscim/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Add user to DB and return error status
func (s *SCIMServer) AddUser(u *types.User) (err error) {
	c := s.db.C("users")
	err = c.Insert(u)
	return err
}

// Get user info
func (s *SCIMServer) GetUser(id string) (*types.User, error) {
	c := s.db.C("users")
	var user *types.User
	err := c.Find(bson.M{"id": id}).One(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Create a user data structure from incoming request body, then return a User struct that
// represent incoming user info in JSON payload
func createUser(b io.Reader) (user *types.User, err error) {
	user = new(types.User)

	if err = json.NewDecoder(b).Decode(&user); err != nil {
		return nil, err
	}
	// Generate metadata
	user.Id = util.GetUUID()
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
		log.Println("inside get")
		// Search for user ID, then return user details in JSON
		user, err := server.GetUser(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var data []byte
		err = json.Unmarshal(data, user)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "http://localhost:8181/Users/"+user.Id)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	case "POST":
		// Create a new resource (user)
		// http://www.simplecloud.info/specs/draft-scim-api-01.html#create-resource

		// Quote:
		//
		// Successful Resource creation is indicated with a 201 ("Created") response code. Upon
		// successful creation, the response body MUST contain the newly created Resource. Since
		// the server is free to alter and/or ignore POSTed content, returning the full
		// representation can be useful to the client, enabling it to correlate the client and
		// server views of the new Resource. When a Resource is created, its URI must be returned
		// in the response Location header.

		// If the Service Provider determines creation of the requested Resource conflicts with
		// existing resources; e.g., a User Resource with a duplicate userName, the Service Provider
		// MUST return a 409 error and SHOULD indicate the conflicting attribute(s) in the body of
		// the response.

		user, err := createUser(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = server.AddUser(user)
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
	log.Println("inside server")
	switch r.URL.Path {
	case "/Users/":
		log.Println("matched Users")
		s.userHandler(w, r)
	case "/Groups":
		s.groupHandler(w, r)
	case "/ServiceProviderConfigs":
		s.spConfigHandler(w, r)
	case "/Schemas":
		s.schemaHandler(w, r)
	case "/Bulk":
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
