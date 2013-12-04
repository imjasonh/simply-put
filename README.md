Simple REST JSON API for putting, getting, listing, updating and deleting data objects
-----

First of all, you will need to get an OAuth access token for the scope `https://www.googleapis.com/auth/userinfo.email`

You can do this using a client library, or for testing, by going to the OAuth Playground: https://developers.google.com/oauthplayground/

**Create an object by sending a POST to `/<Kind>`**

For all examples, the kind being used is `MyKindOfData`

        $ curl https://simply-put.appspot.com/MyKindOfData?access_token=$ACCESS_TOKEN \
              -H "Content-Type: application/json" \
              -X POST \
              -d '{"a":1,"b":false,"c":["ho",1,true]}' | python -m json.tool
        {
            "_created": 1386021382,
            "_id": 1001,
            "a": 1,
            "b": false,
            "c": [
                "ho",
                1,
                true
            ]
        }

**Get an object by sending a GET to `/<Kind>/ID`**

        $ curl https://simply-put.appspot.com/MyKindOfData/1001?access_token=$ACCESS_TOKEN
        {
            "_created": 1386021382,
            "_id": 1001,
            "a": 1,
            "b": false,
            "c": [
                "ho",
                1,
                true
            ]
        }

**Update an object by sending a POST to `/<Kind>/ID`**

        $ curl https://simply-put.appspot.com/MyKindOfData/1001?access_token=$ACCESS_TOKEN \
              -H "Content-Type: application/json" \
              -X POST \
              -d '{"a":3,"b":true,"c":["ho",1,true]}' | python -m json.tool
        {
            "_created": 1386021382,
            "_id": 1001,
            "_updated": 1386021425,
            "a": 3,
            "b": true,
            "c": [
                "ho",
                1,
                true
            ]
        }

**List objects by sending a GET to `/<Kind>` without the ID**

        $ curl https://simply-put.appspot.com/MyKindOfData?access_token=$ACCESS_TOKEN | python -m json.tool
        {
            "items": [
                {
                    "_created": 1386021382,
                    "_id": 1,
                    "a": 1,
                    "b": false,
                    "c": [
                        "ho",
                        1,
                        true
                    ]
                },
                {
                    "_created": 1386021382,
                    "_id": 1001,
                    "a": 1,
                    "b": false,
                    "c": [
                        "ho",
                        1,
                        true
                    ]
                }
            ],
            "nextStartToken": "<<next_page_token>>"
        }


**Delete an object by sending a DELETE to `/<Kind>/ID`**

        $ curl https://simply-put.appspot.com/MyKindOfData/1001?access_token=$ACCESS_TOKEN \
              -X DELETE
        (There is no response in this case)
