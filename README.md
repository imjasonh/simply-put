datastore-api
=============

REST API for the App Engine Datastore (like Parse, but stupider)

To use
------

* Get an OAuth access token for the scope `https://www.googleapis.com/auth/userinfo`

Create an object by sending a POST to `/<Kind>`
-----------------------------------------------

Request:

        POST https://datastore-api.appspot.com/datastore/v1dev/objects/HighScore?access_token=<access-token>
        Content-Type: application/json
        
        {"name": "Bob Smith", "highScore": 1000}
        
Response:

        {
         "name": "Bob Smith",
         "highScore": 1000,
         "_kind": "USERID--HighScore",
         "_id": 12345,
         "_created": "2012-12-11T15:47:46.938347Z"
        }

Get an object by sending a GET to `/<Kind>/ID`
----------------------------------------------

Request:

        GET https://datastore-api.appspot.com/datastore/v1dev/objects/HighScore/12345?access_token=<access-token>
        
Response:

        {
         "name": "Bob Smith",
         "highScore": 1000,
         "_kind": "USERID--HighScore",
         "_id": 12345,
         "_created": "2012-12-11T15:47:46.938347Z"
        }

Update an object by sending a POST to `/<Kind>/ID`
--------------------------------------------------

Request:

        POST https://datastore-api.appspot.com/datastore/v1dev/objects/HighScore/12345?access_token=<access-token>
        Content-Type: application/json
        
        {"name": "Bob Smith", "highScore": 2000}
        
Response:

        {
         "name": "Bob Smith",
         "highScore": 2000,
         "_kind": "USERID--HighScore",
         "_id": 12345,
         "_created": "2012-12-11T15:47:46.938347Z",
         "_updated": "2012-12-11T15:51:12.938347Z"
        }

List objects by sending a GET to `/<Kind>` without the ID
---------------------------------------------------------

Request:

        GET https://datastore-api.appspot.com/datastore/v1dev/objects/HighScore?access_token=<access-token>
        
Response:

        {
         "items": [
          {
           "name": "Bob Smith",
           "highScore": 2000,
           "_kind": "USERID--HighScore",
           "_id": 12345,
           "_created": "2012-12-11T15:47:46.938347Z",
           "_updated": "2012-12-11T15:51:12.938347Z"
          }
         ],
         "nextStartToken": "fj230afaslkf302"
        }

Delete an object by sending a DELETE to `/<Kind>/ID`
--------------------------------------------------

Request:

        DELETE https://datastore-api.appspot.com/datastore/v1dev/objects/HighScore/12345?access_token=<access-token>
        
Response:

        <No content>