# urlshrt
urlshrt is a fast and small URL shortener that can be deployed as a single binary. It's only dependency is a running MongoDB instance.

# Running urlshrt
Please do not forget to set up a working MongoDB instance for urlshrt.
## Environment variables
urlshrt requires the following environment variables to be set:
- SERVER_PORT: The port urlshrt should listen on
- MONGO_URI: The URI your MongoDB instance is running on
- MONGO_PORT: The port your MongoDB instance is running on
- MONGO_USER: A username that has access to the database "sh" and collection "redirects" (these default names can be changed in the source, see [here](https://github.com/niklasstich/urlshrt/blob/master/main.go#L25-L26))
- MONGO_PASSWORD: The password to the MONGO_USER
- HOSTNAME: The hostname of the server (optional, used to display a redirect hint on successful shorthand creation)
## Docker (recommended)
You can compile urlshrt as a very slim alpine-based image with the included Dockerfile. Then simply start the image with the appropriate environment variables you require.
## Manual deployment
Requires the golang toolchain and go-bindata. First, run `go-bindata -o data.go index.html` to make sure the binary data in that file is up to date. Then, build the project with `go build -o out/urlshrt .` or optionally install it with `go install .`
