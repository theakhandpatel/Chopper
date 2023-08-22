# URL Shortener Project

## Project Overview
This is a URL shortener project written in Golang. The main purpose of this project is to create shortened URLs for long URLs, making it easier to share links.

## Features
- Shorten long URLs to shorter, more manageable links.
- Store mappings of short URLs to their corresponding long URLs.
- Redirect users to the original long URL when they access a short URL.

## Technologies Used
- Golang programming language
- Simple in-memory array as the initial database
- SQLite database (planned for future implementation)


## How It Works
1. User submits a long URL to the application.
2. The application generates a unique short key for the long URL.
3. The short key and its corresponding long URL are stored in the database.
4. When a user accesses a short URL, the application looks up the short key in the database and redirects the user to the original long URL.

## Future Plans
Currently, the project uses a simple in-memory array to store mappings between short keys and long URLs. In the future, this will be replaced with an SQLite database for better data persistence.

## Usage
To run the project:
```bash
go run /cmd/api/
```
