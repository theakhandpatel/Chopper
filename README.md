# Chopper (URL Shortener) 🚀

![GitHub last commit](https://img.shields.io/github/last-commit/theakhandpatel/url_shortner)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Chopper is a fast and efficient web application built with Go (Golang) that enables users to shorten long URLs into compact, easy-to-share short URLs. Whether you're sharing links on social media, embedding URLs in emails, or simply trying to manage long and complex URLs, this application provides a convenient solution for creating and expanding short URLs. 🔗

## Deployed
Access at [ch-op.onrender.com](https://ch-op.onrender.com/)

## API Endpoints 📡

Go to [API DOCS](https://ch-op.onrender.com/docs) 📚

## Features 🌟

- **URL Shortening**: Convert lengthy URLs into shorter, more manageable versions. ✂️
- **URL Expansion**: Restore short URLs back to their original long forms. 🔄
- **Analytics Tracking**: Record and track analytics data for every access to a short URL. 📊
- **Daily Limiting**: Set Daily limits for shortening for anonymous and non-premium users. 📆
- **Rate Limiting**: Prevent abuse by setting limits on the number of resolution requests from an IP address. 🚫
- **Collision Resolution**: Handle potential collisions in short URL generation to ensure uniqueness. ⚙️
- **Configuration Flexibility**: Customize the application behavior with command-line flags. 🛠️
- **Lightweight Framework**: Utilizes the Chi router for efficient HTTP routing. 🚀
- **Reliable Database**: Stores URL records and analytics data in an SQLite database. 🗃️
- **Nano Ids**: Uses NanoID of size 6 and 8 encoding for efficient and URL-friendly short URL generation. 🆔
- **Email Notifications**: Send email notifications for various actions, including:
  - User sign-up confirmation.
  - Password reset requests. 📧

## Technologies Used 💻

- **Programming Language**: Go (Golang) 🐹
- **Web Framework**: Chi Router 🛣️
- **Database**: SQLite 📂
- **Rate Limiting**: Token Bucket Algorithm ⏳
- **Encoding**: Base(A-Za-z0.9_-) 🧮


# How to Run 🏃‍♂️

1. **Clone the Repository**:
   ```
   git clone https://github.com/theakhandpatel/url_shortner.git
   cd url_shortner
   ```

2. **Build and Run the Application**:
   To build and run the application using `go run` with command-line flags:
   ```
   go run ./cmd/api/ -port=8080
   ```

   Alternatively, you can build the application using `go build` and then run the executable with flags:
   ```
   go build -o url_shortener ./cmd/api/ ./url_shortener
   ```

   Note: You can adjust the flag values according to your preferences.

3. **Access the Application**:
   Open a web browser and navigate to `http://localhost:8080`.
## Configuration ⚙️

The application supports configuration through command-line flags. Here's a breakdown of the available options:

- `-port`: Specifies the port number to run the server on (default: 8080). 🌐
- `-limiter-enabled`: Enables rate limiting for incoming requests (default: true). ⏳
- `-limiter-rps`: Sets the rate limiter's maximum requests per second (default: 2). 🚀
- `-limiter-burst`: Sets the rate limiter's maximum burst capacity (default: 4). 💥
- `-dsn`: Specifies the path to the SQLite database file (default: `./database.db`). 📂
- `-dailylimiter-enabled`: Enables daily limiting for requests (default: true). 📅
- `-dailyLimiter-ip`: Sets the daily limit for anonymous users (by IP) (default: 3.0). 🕒
- `-dailyLimiter-id`: Sets the daily limit for authenticated users (default: 10.0). 🕙
- `-migrations`: Specifies the relative path to the migrations folder (default: `./migrations`). 🗃️
- `-smtp-host`: SMTP host for email notifications (default: `smtp.mailtrap.io`). 📧
- `-smtp-port`: SMTP port for email notifications (default: 2525). 📮
- `-smtp-username`: SMTP username for email notifications (default: `null`). 👤
- `-smtp-password`: SMTP password for email notifications (default: `null`). 🔑
- `-smtp-sender`: Sender email address for SMTP notifications (default: `null`). 📤

You can adjust these flags to configure the application according to your requirements. 🛠️


## Technical Decisions 🧐

- **Database**: SQLite was chosen for its simplicity and portability, suitable for this project's scope. 📁
- **Rate Limiting**: Token bucket algorithm offers a balance between simplicity and effectiveness. ⏳
- **NanoID**: A tiny, secure, URL-friendly, unique string ID generator. 🆔
- **Chi Router**: Provides a lightweight and efficient routing framework. 🛣️
- **Flags**: Command-line flags offer configuration flexibility without the need for an external configuration file. 🚩


## License 📄

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 📜
