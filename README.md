# URL Shortener

![GitHub last commit](https://img.shields.io/github/last-commit/theakhandpatel/url_shortner)
![GitHub license](https://img.shields.io/github/license/theakhandpatel/url_shortner)

URL Shortener is a fast and efficient web application built with Go (Golang) that enables users to shorten long URLs into compact, easy-to-share short URLs. Whether you're sharing links on social media, embedding URLs in emails, or simply trying to manage long and complex URLs, this application provides a convenient solution for creating and expanding short URLs.

## Features

- **URL Shortening**: Convert lengthy URLs into shorter, more manageable versions.
- **URL Expansion**: Restore short URLs back to their original long forms.
- **Analytics Tracking**: Record and track analytics data for every access to a short URL.
- **Rate Limiting**: Prevent abuse by setting limits on the number of requests from an IP address.
- **Collision Resolution**: Handle potential collisions in short URL generation to ensure uniqueness.
- **Configuration Flexibility**: Customize the application behavior with command-line flags.
- **Lightweight Framework**: Utilizes the Chi router for efficient HTTP routing.
- **Reliable Database**: Stores URL records and analytics data in an SQLite database.
- **Base62 Encoding**: Uses Base62 encoding for efficient and URL-friendly short URL generation.

## Technologies Used

- **Programming Language**: Go (Golang)
- **Web Framework**: Chi Router
- **Database**: SQLite
- **Rate Limiting**: Token Bucket Algorithm
- **Encoding**: Base62
- **External Libraries**: `github.com/mattn/go-sqlite3`, `github.com/go-chi/chi`, `github.com/asaskevich/govalidator`

## API Endpoints

1. **Health Check**
   - Endpoint: `/`
   - Method: GET
   - Description: Check the health status of the application.


2. **Shorten URL**
   - Endpoint: `/api/shorten`
   - Method: GET
   - Description: Shorten a provided long URL and receive a short URL.
   - Query Parameter: `URL` (required) - The long URL to be shortened.


3. **Expand URL**
   - Endpoint: `/{shortURL}`
   - Method: GET
   - Description: Expand a short URL and redirect to the original long URL.
   - Path Parameter: `shortURL` (required) - The short URL to be expanded.


4. **Analytics Data**
   - Endpoint: `/api/stats`
   - Method: GET
   - Description: Retrieve analytics data for a given short URL.
   - Query Parameter: `URL` (required) - The short URL for which analytics data is requested.


# How to Run

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/theakhandpatel/url_shortner.git
   cd url_shortner
   ```

2. **Build and Run the Application**:
   To build and run the application using `go run` with command-line flags:
   ```bash
   go run ./cmd/api/ -port=8080 -limiter-enabled=true -limiter-rps=2 -limiter-burst=4 -dsn=./database.db
   ```

   Alternatively, you can build the application using `go build` and then run the executable with flags:
   ```bash
   go build -o url_shortener ./cmd/api/
   ./url_shortener -port=8080 -limiter-enabled=true -limiter-rps=2 -limiter-burst=4 -dsn=./database.db
   ```

   Note: You can adjust the flag values according to your preferences.

3. **Access the Application**:
   Open a web browser and navigate to `http://localhost:8080`.


## Configuration

The application supports configuration through command-line flags. Below are notable flags:

- `-port`: Specifies the port number to run the server on (default: 8080).
- `-limiter-enabled`: Enables rate limiting for incoming requests (default: false).
- `-limiter-rps`: Sets the rate limiter's maximum requests per second (default: 2).
- `-limiter-burst`: Sets the rate limiter's maximum burst capacity (default: 4).
- `-dsn`: Specifies the path to the SQLite database file (default: `./database.db`).
- `-temp-redirect`: Enables temporary redirects (default: false).

## Technical Decisions

- **Database**: SQLite was chosen for its simplicity and portability, suitable for this project's scope.
- **Rate Limiting**: Token bucket algorithm offers a balance between simplicity and effectiveness.
- **Base62 Encoding**: Used for generating short URLs, provides a compact representation.
- **Chi Router**: Provides a lightweight and efficient routing framework.
- **Flags**: Command-line flags offer configuration flexibility without the need for an external configuration file.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
