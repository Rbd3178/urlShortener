# Go URL Shortener HTTP Server

This is a simple URL shortener implemented in Go, utilizing a custom Red Black Tree implementation for efficient storage. The server allows users to store long URLs by providing an alias, and supports concurrent searches.

## Usage

### Redirect to Stored URL

To redirect to the URL stored by the provided alias, send a GET request to `/go/<alias>`.

#### Example
``` bash
curl -X GET "http://localhost:8090/go/myalias"` 
```

### Get all aliases with a prefix

To retrieve a JSON array containing alias-URL pairs where the alias has the specified prefix, send a GET request.

#### Structure
- **HTTP Method:** `GET`
- **Endpoint URL:** `/search`
- **Parameters:** Parameters are passed in the query string.

#### Parameters
- **`prefix`** (string, required)

#### Example
```http
GET /search?prefix=my HTTP/1.1
Host: localhost:8090
```
```bash
curl -X GET "localhost:8090/search?prefix=my"
```


### Add a URL Alias

To add a URL alias to the server, Send a POST request to `/add`. 

#### Structure
- **HTTP Method:** `POST`
- **Endpoint URL:** `/add`
- **Body:** Parameters are passed in the body of the request.

#### Parameters
- **`alias`** (string, not required) If no alias is provided, the server will choose one that is lexicographically bigger than all of the others stored with a smallest possible length.
- **`url`** (string, required) The URL that wil be stored by the alias.

#### Example
```http
POST /add HTTP/1.1
Host: localhost:8090

alias=myalias&url=http://www.example.com
```
```bash
curl -X POST "localhost:8090/add" -d "alias=myalias&url=http://www.example.com"
```

### Delete a URL Alias

To delete a URL alias from the server, Send a DELETE request to `/delete`. 

#### Structure
- **HTTP Method:** `DELETE`
- **Endpoint URL:** `/delete`
- **Body:** Parameters are passed in the query string.

#### Parameters
- **`alias`** (string, required) The alias to be deleted.

#### Example
```http
DELETE /delete?alias=myalias HTTP/1.1
Host: localhost:8090
```
```bash
curl -X DELETE "localhost:8090/delete?alias=myalias"
```

## Implementation details
- **Concurrency**: Redirects and searches are executed concurrently with eachother. Modifications are done atomically, so reads only access a valid data structure. `sync.RWMutex`, `sync.WaitGroup` and `sync.Mutex` are used to achieve this.
- **Storage**: URLs and their corresponding aliases are stored in a Red Black Tree for efficient retrieval.
- **Handling conflicts**: Passed URLs are validated and conflicts like trying to add an already existing alias are handled. Proper HTTP status codes and error messages are returned for various scenarios.

## Limitations

- Data is stored in RAM in this implementation, so it will be lost when the server stops running.
