Simple REST JSON API built on BoltDB
-----

[![Build Status](https://travis-ci.org/ImJasonH/simply-put.svg)](https://travis-ci.org/ImJasonH/simply-put)

First, run your server:

```
$ go run main.go server.go
```

By default this creates a file `bolt.db` that stores your data using [BoltDB](https://github.com/boltdb/bolt) -- you can change the location of this file with the `-db` flag.

Then send HTTP requests to interact with data:

**Create an object by sending a POST to `/<Kind>`**

For all examples, the kind being used is `Data` but it could be anything, `User`, `Object`, `Kittens`, knock yourself out.

        $ curl http://localhost:8080/Data \
              -H "Content-Type: application/json" \
              -X POST \
              -d '{"a":1,"b":false,"c":["foo",1,true]}' | python -m json.tool
        {
            "_created": 1386021382,
            "_id": <uuid>,
            "a": 1,
            "b": false,
            "c": [
                "foo",
                1,
                true
            ]
        }

This responds with the same JSON you provided, plus two new keys: `"_id"` is the assigned ID of the new entity, and `"_created"` is the timestamp it was created.

You can use the `<uuid>` to `GET` the data:

**Get an object by sending a GET to `/<Kind>/<uuid>`**

        $ curl http://localhost:8080/Data/<uuid>
        {
            "_created": 1386021382,
            "_id": <uuid>,
            "a": 1,
            "b": false,
            "c": [
                "foo",
                1,
                true
            ]
        }

**Update an object by sending a POST to `/<Kind>/ID`**

        $ curl http://localhost:8080/Data/<uuid> \
              -H "Content-Type: application/json" \
              -X POST \
              -d '{"a":3,"b":true,"c":["foo",1,true]}' | python -m json.tool
        {
            "_created": 1386021382,
            "_id": <uuid>,
            "_updated": 1386021425,
            "a": 3,
            "b": true,
            "c": [
                "foo",
                1,
                true
            ]
        }

Note that now the object has a new key, `"_updated"` which indicates that it has been updated, and when.

**List objects by sending a GET to `/<Kind>` without the ID**

        $ curl http://localhost:8080/Data | python -m json.tool
        {
            "items": [
                {
                    "_created": 1386021382,
                    "_id": <uuid>,
                    "a": 1,
                    "b": false,
                    "c": [
                        "foo",
                        1,
                        true
                    ]
                },
                {
                    "_created": 1386021382,
                    "_id": <uuid>,
                    "a": 1,
                    "b": false,
                    "c": [
                        "foo",
                        1,
                        true
                    ]
                }
            ],
            "nextStartToken": "<<next_page_token>>"
        }


**Delete an object by sending a DELETE to `/<Kind>/<uuid>`**

        $ curl http://localhost:8080/Data/<uuid> \
              -X DELETE
        (There is no response in this case)


----------

License
-----

    Copyright 2015 Jason Hall

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
