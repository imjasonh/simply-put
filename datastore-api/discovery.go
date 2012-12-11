package datastoreapi

import (
	"net/http"
)

func init() {
	http.HandleFunc("/discovery/v1/apis/datastore/v1dev/rest", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(discovery))
	})
	http.HandleFunc("/discovery/v1/apis", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(directory))
	})
}

const discovery = `
{
    "kind": "discovery#restDescription",
    "discoveryVersion": "v1",
    "id": "datastore:v1dev",
    "name": "datastore",
    "canonicalName": "Datastore",
    "version": "v1dev",
    "revision": "20121210",
    "title": "Datastore API",
    "description": "Simple REST access to the App Engine Datastore",
    "documentationLink": "TODO",
    "protocol": "rest",
    "baseUrl": "https://datastore-api.appspot.com/datastore/v1dev/",
    "basePath": "/datastore/v1dev",
    "rootUrl": "https://datastore-api.appspot.com/",
    "servicePath": "datastore/v1dev/",
    "parameters": {
        "oauth_token": {
            "type": "string",
            "description": "OAuth 2.0 token for the current user.",
            "location": "query"
        }
    },
    "auth": {
        "oauth2": {
            "scopes": {
                "https://www.googleapis.com/auth/userinfo": {
                    "description": "View your user information."
                }
            }
        }
    },
    "schemas": {
        "DataObject": {
            "id": "DataObject",
            "type": "object",
            "properties": {
                "_id": {
                    "type": "string",
                    "description": "Unique ID of this object."
                },
                "_created": {
                    "type": "string",
                    "format": "date-time",
                    "description": "Timestamp when the object was created."
                },
                "_updated": {
                    "type": "string",
                    "format": "date-time",
                    "description": "Timestamp when the object was last updated."
                },
                "additionalProperties": {
                    "type": "any"
                }
            }
        }
    },
    "resources": {
        "objects": {
            "methods": {
                "insert": {
                    "id": "datastore.objects.insert",
                    "path": "{kind}/{id}",
                    "parameters": {
                        "kind": {
                            "type": "string",
                            "required": true,
                            "location": "path"
                        }
                    },
                    "request": {
                        "$ref": "DataObject"
                    },
                    "response": {
                        "$ref": "DataObject"
                    },
                    "scopes": [
                        "https://www.googleapis.com/auth/userinfo"
                    ]
                },
                "update": {
                    "id": "datastore.objects.update",
                    "path": "{kind}/{id}",
                    "parameters": {
                        "id": {
                            "type": "integer",
                            "format": "uint32",
                            "required": true,
                            "location": "path"
                        },
                        "kind": {
                            "type": "string",
                            "required": true,
                            "location": "path"
                        }
                    },
                    "request": {
                        "$ref": "DataObject"
                    },
                    "response": {
                        "$ref": "DataObject"
                    },
                    "scopes": [
                        "https://www.googleapis.com/auth/userinfo"
                    ]
                },
                "get": {
                    "id": "datastore.objects.get",
                    "path": "{kind}/{id}",
                    "parameters": {
                        "id": {
                            "type": "integer",
                            "format": "uint32",
                            "required": true,
                            "location": "path"
                        },
                        "kind": {
                            "type": "string",
                            "required": true,
                            "location": "path"
                        }
                    },
                    "parameterOrder": [
                        "kind",
                        "id"
                    ],
                    "response": {
                        "$ref": "DataObject"
                    },
                    "scopes": [
                        "https://www.googleapis.com/auth/userinfo"
                    ]
                },
                "delete": {
                    "id": "datastore.objects.delete",
                    "path": "{kind}/{id}",
                    "parameters": {
                        "id": {
                            "type": "integer",
                            "format": "uint32",
                            "required": true,
                            "location": "path"
                        },
                        "kind": {
                            "type": "string",
                            "required": true,
                            "location": "path"
                        }
                    },
                    "parameterOrder": [
                        "kind",
                        "id"
                    ],
                    "response": {
                        "$ref": "DataObject"
                    },
                    "scopes": [
                        "https://www.googleapis.com/auth/userinfo"
                    ]
                },
                "list": {
                    "id": "datastore.objects.list",
                    "path": "{kind}",
                    "parameters": {
                        "kind": {
                            "type": "string",
                            "required": true,
                            "location": "path"
                        }
                    },
                    "response": {
                        "$ref": "DataObject"
                    },
                    "scopes": [
                        "https://www.googleapis.com/auth/userinfo"
                    ]
                }
            }
        }
    }
}`

const directory = `
{
    "kind": "discovery#directoryList",
    "discoveryVersion": "v1",
    "items": [
        {
            "kind": "discovery#directoryItem",
            "id": "datastore:v1dev",
            "name": "datastore",
            "version": "v1dev",
            "title": "Datastore API",
            "description": "Simple REST access to the App Engine Datastore.",
            "discoveryRestUrl": "https://datastore-api.appspot.com/discovery/v1/apis/datastore/v1dev/rest",
            "discoveryLink": "./apis/datastore/v1dev/rest",
            "documentationLink": "https://www.github.com/ImJasonH/datastore-api",
            "preferred": true
        }
    ]
}`
