# ENDPOINTS

## Signup

- **Endpoint**: `/api/signup`

- **Request**:

  ```bash
  curl  -X POST 'http://localhost:8080/api/signup' --header 'Content-Type: application/json' --data-raw '{"name":"user1", "email":"user1@email.com", "password":"password"}'
  ```

- **Response**:
  ```json
  {
    "user": {
      "id": 2,
      "created_at": "0001-01-01T00:00:00Z",
      "name": "user1",
      "email": "user1@email.com",
      "activated": false,
      "type": 1
    }
  }
  ```

## Signin

- **Endpoint**: `/api/signin`

- **Request**:

  ```bash
  curl  -X POST 'http://localhost:8080/api/signin' --header 'Content-Type: application/json' --data-raw '{"email":"user1@email.com", "password":"password"}'
  ```

- **Response**:
  ```json
  {
    "authentication_token": {
      "token": "FAS5CUG6NPRPIQM6U7MRVF3XCY",
      "expiry": "2023-09-15T20:57:50.815544398+05:30"
    }
  }
  ```

## Signout

- **Endpoint**: `/api/signout`

- **Request**:

  ```bash
  curl  -X POST 'http://localhost:8080/api/signout' --header 'Authorization: Bearer FAS5CUG6NPRPIQM6U7MRVF3XCY'
  ```

- **Response**:
  ```json
  {
    "message": "logged out"
  }
  ```


## Premium

- **Endpoint**: `/api/premium`

- **Request**:

  ```bash
  curl  -X POST 'http://localhost:8080/api/premium' --header 'Authorization: Bearer RQ6AIO7ZIISWLUI7VXC72A7ZUU'
  ```

- **Response**:
  ```json
  {
    "message": "You are premium user"
  }
  ```

## Anonymous Users

### Shorten URL

- **Endpoint**: `/api/short`

- **Request**:
  ```bash
  curl  -X POST 'http://localhost:8080/api/short' --header 'Content-Type: application/json' --data-raw '{ "long" : "www.bing.com"}'
  ```

- **Response**:
  ```json
  {
    "url": {
      "ID": 12,
      "Long": "http://www.bing.com",
      "Short": "Rd6ee2mF",
      "Redirect": 308,
      "Modified": "2023-09-14T20:23:43.353311103+05:30"
    }
  }
  ```

### Expand URL

- **Endpoint**: `/{shortURL}`
- **Request**:

  ```bash
  curl -i  'http://localhost:8080/Rd6ee2mF'
  ```

- **Response**:
  ```
  HTTP/1.1 308 Permanent Redirect
  Location: http://www.bing.com
  Content-Length: 55
  <a href="http://www.bing.com">Permanent Redirect</a>.
  ```



## Authenticated Non-Premium Users

### Shorten URL

- **Endpoint**: `/api/short`

- **Request**:
  ```bash
  curl  -X POST 'http://localhost:8080/api/short' --header 'Authorization: Bearer LUQVPB52N4DX6UMDX4HLDHW7AU' --header 'Content-Type: application/json' --data-raw '{ "long" : "www.google.com"}'
  ```

- **Response**:
  ```json
  {
  "url": {
    "ID": 13,
    "Long": "http://www.google.com",
    "Short": "Aihat5",
    "Redirect": 308,
    "Modified": "2023-09-14T21:22:25.480533924+05:30"
    }
  }
  ```


### Get Short URL Details

- **Endpoint**: `/api/short/{shortCode}`

- **Request**:
  ```bash
  curl  -X GET \
  'http://localhost:8080/api/short/Aihat5' --header 'Authorization: Bearer LUQVPB52N4DX6UMDX4HLDHW7AU'
  ```

- **Response**:
  ```json
  {
  "url": {
    "ID": 13,
    "Long": "http://www.google.com",
    "Short": "Aihat5",
    "Redirect": 308,
    "Modified": "2023-09-14T21:22:25.480533924+05:30"
    }
  }
  ```


  
## Authenticated Premium Users

### Shorten URL
  No Daily limit


### Edit Short URL

- **Endpoint**: `/api/short/{shortCode}` (HTTP PUT)

- **Request**:
  ```bash
  curl -X PUT 'http://localhost:8080/api/short/Aihat5' \
    --header 'Authorization: Bearer AUTH_TOKEN' \
    --header 'Content-Type: application/json' \
    --data-raw '{ "long": "www.updated-example.com", "short": "UpdatedShortCode", "redirect": "permanent" }'
  ```
- **Response**
  ```json
  {
    "url": {
      "ID": 13,
      "Long": "http://www.updated-example.com",
      "Short": "UpdatedShortCode",
      "Redirect": 308,
      "Modified": "2023-09-14T21:45:00.123456789+05:30"
    }
  }
  ```

### Delete Short URL

- **Endpoint**: `/api/short/{shortCode}` (HTTP DELETE)

- **Request**:
  ```bash
  curl -X DELETE 'http://localhost:8080/api/short/Aihat5' --header 'Authorization: Bearer AUTH_TOKEN'
  ```
- **Response**
  ```json
  Status: 204
  ```