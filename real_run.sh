#!/bin/bash
# Example test file for initial testing
# TODO MUST be removed after initial testing phase

EXAMPLE=$1
## Example 1: mutiple key value pairs within one yaml file
if [[ "$EXAMPLE" == "1" ]]; then 
    # select secrets based on name
    export SECRET_SELECTOR_NAME='^keycloak-client-secret-management-.*$'
    # basic property path, based on secret name
    export SECRET_FILE_PROPERTY_PATTERN='spring.oauth.clients.{{with $arr := splitN .ObjectMeta.Name "-" 4}}{{index $arr 3}}{{end}}'
    # fix filename
    export SECRET_FILE_NAME_PATTERN=$(pwd)"/foo.yaml"
    # select all
    export SECRET_SELECTOR_CONTENT=''
    # property key transformation function
    export SECRET_KEY_TRANSFORMATION="ToLowerCamel"

    make run 
## Example 2: One file per secret
elif [[ "$EXAMPLE" == "2" ]]; then 
    # select secrets based on name
    export SECRET_SELECTOR_NAME='^keycloak-client-secret-management-api-.*$'
    # create one file per secret and use the last part of secret name as dir
    export SECRET_FILE_NAME_PATTERN=$(pwd)'/temp/{{with $arr := splitN .ObjectMeta.Name "-" 6}}{{index $arr 5}}{{end}}/credentials.yaml'
    # just put into root
    export SECRET_FILE_PROPERTY_PATTERN=""
    # get all
    export SECRET_SELECTOR_CONTENT=''
    # use snake case
    export SECRET_KEY_TRANSFORMATION="ToSnake"

    make run 
## Example 3: One directory per secret, one file per secret key
elif [[ "$EXAMPLE" == "3" ]]; then 
    # select secrets based on name
    export SECRET_SELECTOR_NAME='^keycloak-client-secret-management-api-.*$'
    # create one file per secret and use the last part of secret name as dir
    export SECRET_FILE_NAME_PATTERN=$(pwd)'/temp/{{with $arr := splitN .ObjectMeta.Name "-" 6}}{{index $arr 5}}{{end}}'
    # use one file for each secret key
    export SECRET_FILE_SINGLE="true"
    # just put into root
    export SECRET_FILE_PROPERTY_PATTERN=""
    # get all
    export SECRET_SELECTOR_CONTENT=''
    # use snake case
    export SECRET_KEY_TRANSFORMATION="ToLowerCamel"

    make run 
else 
    echo "Unknown exanple"
    exit 1
fi 
