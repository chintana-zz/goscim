package main

import (
	"encoding/json"
	"flag"
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
		return nil, err
	}
	newUser, err := s.GetUser(userId)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

func (s *SCIMServer) DeleteUser(userId string) error {
	c := s.db.C("users")
	err := c.Remove(bson.M{"id": userId})
	if err != nil {
		return err
	}
	return nil
}

func (s *SCIMServer) AddGroup(body io.ReadCloser) (*types.Group, error) {
	group, err := createGroup("", body)
	if err != nil {
		return nil, err
	}
	c := s.db.C("groups")
	err = c.Insert(group)
	return group, err
}

// Getting group related information
func (s *SCIMServer) GetGroup(id string) (*types.Group, error) {
	c := s.db.C("groups")
	group := types.Group{}
	err := c.Find(bson.M{"id": id}).One(&group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// Update group info
func (s *SCIMServer) UpdateGroup(groupId string, body io.ReadCloser) (*types.Group, error) {
	group, err := createGroup(groupId, body)
	if err != nil {
		return nil, err
	}
	c := s.db.C("groups")
	err = c.Update(bson.M{"id": groupId}, group)
	if err != nil {
		return nil, err
	}
	newGroup, err := s.GetGroup(groupId)
	if err != nil {
		return nil, err
	}
	return newGroup, nil
}

func (s *SCIMServer) DeleteGroup(groupId string) error {
	c := s.db.C("groups")
	err := c.Remove(bson.M{"id": groupId})
	if err != nil {
		return err
	}
	return nil
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
	user.Meta.Location =
		fmt.Sprintf("http://%s:%d/%s/Users/%s", hostFlag, portFlag, SCIM_VERSION, user.Id)

	return user, nil
}

func createGroup(groupId string, b io.Reader) (group *types.Group, err error) {
	group = new(types.Group)

	if err = json.NewDecoder(b).Decode(&group); err != nil {
		return nil, err
	}
	if groupId != "" {
		group.Id = groupId
	} else {
		group.Id = util.GetUUID()
	}
	created := time.Now()
	group.Meta.Created = created.Format(time.RFC3339)
	group.Meta.LastModified = group.Meta.Created
	group.Meta.Location =
		fmt.Sprintf("http://%s:%d/%s/Groups/%s", hostFlag, portFlag, SCIM_VERSION, group.Id)

	return group, nil
}

// Connect to the given MongoDB instance and return a pointer to the connection. DB will be
// accessed through this pointer
func NewMongoDS(url string) *mgo.Database {
	fmt.Println("Connecting to MongoDB -", url)
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

// Handle all requests coming to /Users and /Groups- GET, POST, PUT, PATCH, DELETE
func (s *SCIMServer) userHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Search for user ID, then return user details in JSON

		// Passing the UUID extracted from the URL path
		user, err := s.GetUser(r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var data []byte
		data, err = json.Marshal(user)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location",
			fmt.Sprintf("http://%s:%d/%s/Users/%s", hostFlag, portFlag, SCIM_VERSION, user.Id))
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	case "POST":
		// Create a new resource (user)
		// http://www.simplecloud.info/specs/draft-scim-api-01.html#create-resource
		user, err := s.AddUser(r.Body)
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
		w.Header().Set("Location",
			fmt.Sprintf("http://%s:%d/%s/Users/%s", hostFlag, portFlag, SCIM_VERSION, user.Id))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(m)
	case "PUT":
		// Update a user, passing user ID from URL and req body
		user, err := s.UpdateUser(r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:], r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m, err := json.Marshal(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location",
			fmt.Sprintf("http://%s:%d/%s/Users/%s", hostFlag, portFlag, SCIM_VERSION, user.Id))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(m)
	case "DELETE":
		err := s.DeleteUser(r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

// Handle all requests coming to /Groups - GET, POST, PUT, PATCH, DELETE
func (s *SCIMServer) groupHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		group, err := s.GetGroup(r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var data []byte
		data, err = json.Marshal(group)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location",
			fmt.Sprintf("http://%s:%d/%s/Groups/%s", hostFlag, portFlag, SCIM_VERSION, group.Id))
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	case "POST":
		// Create a new resource (user)
		// http://www.simplecloud.info/specs/draft-scim-api-01.html#create-resource
		group, err := s.AddGroup(r.Body)
		if err != nil {
			// TODO: check whether it's a duplicate username and throw a 409
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m, err := json.Marshal(group)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location",
			fmt.Sprintf("http://%s:%d/%s/Groups/%s", hostFlag, portFlag, SCIM_VERSION, group.Id))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(m)
	case "PUT":
		// Update a user, passing user ID from URL and req body
		group, err := s.UpdateUser(r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:], r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m, err := json.Marshal(group)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location",
			fmt.Sprintf("http://%s:%d/%s/Groups/%s", hostFlag, portFlag, SCIM_VERSION, group.Id))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(m)
	case "DELETE":
		err := s.DeleteGroup(r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
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

var (
	hostFlag   string // hostname to listen to
	portFlag   int    // port to be used for the server
	dbhostFlag string // MongoDB hostname
	dbportFlag int    // MongoDB port
)

const SCIM_VERSION = "v2"

func init() {
	flag.StringVar(&hostFlag, "host", "localhost", "SCIM server host - listen address")
	flag.IntVar(&portFlag, "port", 8080, "SCIM server port - listen port")

	// MongoDB IP address - giving hostname doesn't work for some reason
	flag.StringVar(&dbhostFlag, "dbhost", "127.0.0.1", "MongoDB instance host")

	// Default MongoDB port on localhost, defaults to 27017
	flag.IntVar(&dbportFlag, "dbport", 27017, "MongoDB instance port")
}

func main() {
	flag.Parse()

	// Initialize server environment
	log.Println(fmt.Sprintf("%s:%d", hostFlag, portFlag))
	s := &SCIMServer{
		// database instance
		db: NewMongoDS(fmt.Sprintf("%s:%d", dbhostFlag, dbportFlag)),
	}

	http.Handle("/", s)
	address := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	log.Println("Server started", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
