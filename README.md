# GoSCIM #

SCIM 2.0 spec implementation in Go. Still very much work in progress

### Implementation summary so far ###

* /Users - GET, POST, PUT, DELETE

### How do I run this? ###

1. This use MongoDB as the persistent store. Install MongoDB and run it with the default port 127.0.0.1:27017. It'll use a db called "users" to
 store user info
2. Set $GOPATh and execute following
    * $ go get bitbucket.org/chintana/goscim
    * $ go install bitbucket.org/chintana/goscim
    * $ goscim

### TODO ###

* Make functionality spec complete
* Make hard coded entries into a config file
* Generate/document using Swagger
