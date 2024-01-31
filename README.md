# Secret File Provider

[![semantic-release: angular](https://img.shields.io/badge/semantic--release-angular-e10079?logo=semantic-release)](https://github.com/semantic-release/semantic-release)

Sidecar implementation which is used to copy K8s secret content into local filesystem.

## Configuration

* port - change port configuration for this service
  * healthcheck - healthcheck port (default 8383)
  * metrics - metrics port (default 8080)
  * debug - expose golang [debug](https://pkg.go.dev/net/http/pprof) information (default 1234)
* log - logging settings
  * json - if set to 'true', json logging will be enabled (default false)
  * level - log level (default info), one of [panic|fatal|error|warn|info|debug|trace]
* callback - HTTP call definition, made for every successful file update
  * url - URL to call for file updates
  * method - HTTP method to use for callback (default GET), one of [GET|POST|HEAD|PUT|PATCH|DELETE]
  * body - HTTP request body, sent for file updated (default empty). Supports [golang template](https://pkg.go.dev/text/template) syntax
  * contenttype - request body content type (default 'application/json' if body is sent)
* secret - configuration for secret access and target mappings
  * selector - selector configuration
    * label - [Label selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) for selecting secrets (either label **or** name selector **must** be set)
    * name - name selector for accessing secrets in Regex format (either label **or** name selector **must** be set)
    * namespace - optional, comma separated list of namespaces to check secrets for (default empty, meaning, all namespaces are checked)
    * content - (optional) select specific fields from the secret in [golang template](https://pkg.go.dev/text/template) syntax
  * file - target file configuration
    * single - if set to true, each key in each secret will get it's own file with the value as only content (default false)
    * name.pattern - naming pattern of the target file, supporting [golang template](https://pkg.go.dev/text/template) syntax. If *single* is set, this will be used as target *directory* pattern for the single files.
    * property.pattern - (optional) property base path to map the secret content under, supporting [golang template](https://pkg.go.dev/text/template) syntax
  * key.transformation - (optional) transformation function for the keys in the secret; one of [ToCamel|ToLowerCamel|ToKebab|ToScreamingKebab|ToSnake|ToScreamingSnake]
  * deletion.watch - (optional) if set to *true*, sidecar will watch for secret deletion and drop their content from the
  file-system as well. Note that **should not be used** at the moment, as this implementation currently adds finalizers
  to secrets, which will not get removed.

## Examples

### Copy into single properties file

Example Config
```
SECRET_SELECTOR_NAME="auth-client-.*"
SECRET_SELECTOR_CONTENT="{{.Data.CLIENT_ID}}"
SECRET_FILE_NAME_PATTERN="/var/config/secret.yaml"
SECRET_FILE_PROPERTY_PATTERN='spring.oauth.clients.{{with $arr := splitN .ObjectMeta.Name "-" 4}}{{index $arr 3}}{{end}}.clientId'
```

Example Result (/var/config/secret.yaml)
```
spring:
  oauth:
    clients:
      acme:
        clientId: 123-456
      company:
        clientId: 789-012
```

### Copy into multiple files

Example Config
```
SECRET_SELECTOR_LABEL="type in (jwt, oauth)"
SECRET_FILE_NAME_PATTERN='/var/config/{{.ObjectMeta.Labels.company}}/credentials.yaml'
SECRET_FILE_PROPERTY_PATTERN="spring.oauth.clients"
SECRET_KEY_TRANSFORMATION="ToSnake"
```

Example Result
``` 
$ cat /var/config/secret-auth-client-acme.yaml
spring:
  oauth:
    clients:
      client_id: "123-456"
      client_secret: "mySuperSecretSecret"
``` 
Example Result
``` 
$ cat /var/config/secret-auth-client-company.yaml
spring:
  oauth:
    clients:
      client_id: "789-012"
      client_secret: "ImSecure...believeIt!"
``` 

### One directory per secret with multiple files in it

Example Config
```
SECRET_SELECTOR_LABEL="type in (jwt, oauth)"
SECRET_FILE_NAME_PATTERN='/var/config/{{.ObjectMeta.Labels.company}}'
SECRET_FILE_SINGLE="true
SECRET_KEY_TRANSFORMATION="ToLowerCamel"
```

Example Results
``` 
$ cat /var/config/acme/clientId
123-456
$ cat /var/config/acme/clientSecret
mySuperSecretSecret
$ cat /var/config/company/clientId
789-012
$ cat /var/config/company/clientSecret
ImSecure...believeIt!
``` 

## Local Developmet

**Preconditions**
* Installed and set up [Golang 1.19](https://go.dev/doc/install) or newer
* Installed [make](https://www.tutorialspoint.com/unix_commands/make.htm)

To build the tool and run all tests, just use 
```
make all
```

For building a docker container, run
```
make docker-build
``` 
which will create by default a *jaconi.io/secret-file-provider:latest* image.

## Contributing

**TODO** 

## Open Issues

* Deletion case 
  * When using approach with finalizers, those will get stuck forever if the pod is just terminated, as 
  there is no cleanup logic in place