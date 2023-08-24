# URL Shortener Project

## Project Overview
This is a URL shortener project written in Golang. The main purpose of this project is to create shortened URLs for long URLs, making it easier to share links.

## Features
- Shorten long URLs to shorter, more manageable links.
- Store mappings of short URLs to their corresponding long URLs.
- Redirect users to the original long URL when they access a short URL.
- Analytics of shortURls created

## Technologies Used
- Golang programming language
- SQLite database 


## How It Works
1. User submits a long URL to the application.
2. The application generates a unique short key for the long URL.
3. The short key and its corresponding long URL are stored in the database.
4. When a user accesses a short URL, the application looks up the short key in the database and redirects the user to the original long URL.

## Future Plans
Currently, the project uses a simple SQLite to store mappings between short keys and long URLs. In the future, this will be enhanced with REDIS as Cache Layer.

## Usage
To run the project:
```bash
go run /cmd/api/
```
