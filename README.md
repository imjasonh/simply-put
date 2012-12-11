REST API for the App Engine Datastore
=====================================

First of all, you will need to get an OAuth access token for the scope `https://www.googleapis.com/auth/userinfo`

You can do this using a client library, or for testing, by going to the OAuth Playground: https://developers.google.com/oauthplayground/

Create an object by sending a POST to `/<Kind>`
-----------------------------------------------
        $ curl https://datastore-api.appspot.com/datastore/v1dev/objects/MyKindOfData?access_token=$ACCESS_TOKEN \
              -H "Content-Type: application/json" \
              -X POST \
              -d '{"a":1,"b":false,"c":["ho",1,true]}' | python -m json.tool
        {
            "_created": "2012-12-11T16:22:20.943202Z", 
            "_id": 1001, 
            "a": 1, 
            "b": false, 
            "c": [
                "ho", 
                1, 
                true
            ]
        }

Get an object by sending a GET to `/<Kind>/ID`
----------------------------------------------
        $ curl https://datastore-api.appspot.com/datastore/v1dev/objects/MyKindOfData/1001?access_token=$ACCESS_TOKEN
        {
            "_created": "2012-12-11T16:22:20.943202Z", 
            "_id": 1001, 
            "a": 1, 
            "b": false, 
            "c": [
                "ho", 
                1, 
                true
            ]
        }

Update an object by sending a POST to `/<Kind>/ID`
--------------------------------------------------
        $ curl https://datastore-api.appspot.com/datastore/v1dev/objects/MyKindOfData/1001?access_token=$ACCESS_TOKEN \
              -H "Content-Type: application/json" \
              -X POST \
              -d '{"a":3,"b":true,"c":["ho",1,true]}' | python -m json.tool
        {
            "_created": "2012-12-11T16:22:20.943202Z", 
            "_id": 1001, 
            "_updated": "2012-12-11T16:28:56.943202Z", 
            "a": 3, 
            "b": true, 
            "c": [
                "ho", 
                1, 
                true
            ]
        }

List objects by sending a GET to `/<Kind>` without the ID
---------------------------------------------------------
        $ curl https://datastore-api.appspot.com/datastore/v1dev/objects/MyKindOfData?access_token=$ACCESS_TOKEN | python -m json.tool
        {
            "items": [
                {
                    "_created": "2012-12-11T16:21:08.322745Z", 
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
                    "_created": "2012-12-11T16:22:20.943202Z", 
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


Delete an object by sending a DELETE to `/<Kind>/ID`
--------------------------------------------------
        $ curl https://datastore-api.appspot.com/datastore/v1dev/objects/MyKindOfData/1001?access_token=$ACCESS_TOKEN \
              -X DELETE
        (There is no response in this case)