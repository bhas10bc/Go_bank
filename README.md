# Go Bank Project

This project is a simple banking application built using Go and Fiber. It features a discovery service that acts as a proxy to route requests to multiple backend servers for creating and managing accounts. The application utilizes PostgreSQL for data storage and Docker for containerization.

## Table of Contents

- [Features](#features)
- [Technologies Used](#technologies-used)
- [Installation](#installation)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Features

- Create and manage accounts through a RESTful API
- Discovery service for load balancing between multiple backend servers
- PostgreSQL database for data persistence
- Load testing using the `hey` tool
- Concurrent processing with Go goroutines and channels

## Technologies Used

- **Go**: Programming language used for backend development.
- **Fiber**: Web framework for building APIs in Go.
- **PostgreSQL**: Relational database for storing account data.
- **Docker**: Containerization technology for packaging and deploying the application.
- **Discovery Service**: Proxy service to route requests to multiple backend servers.
- **JSON**: Data format used for API requests and responses.
- **Goroutines**: For concurrent request handling.
- **Channels**: For communication between goroutines.
- **Queue Service**: For handling tasks asynchronously.
- **HTTP Load Testing (`hey`)**: For performance testing of the API.

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/bhas10bc/Go_bank.git
   cd go-bank
   ```

2. Ensure you have Docker and Docker Compose installed.

3. Build and run the application using Docker:

   ```bash
   docker-compose up --build
   ```

4. Set up your PostgreSQL database and configure the connection in your application (if not using Docker for the database).

## Usage

The application exposes several API endpoints. You can interact with these endpoints using tools like Postman or `curl`.

### Example Request

To create an account, you can use the following command with `hey` for load testing:

```bash
hey -n 1000 -c 100 -m POST -d '{"firstName":"John","lastName":"Doe"}' -H "Content-Type: application/json" http://127.0.0.1:8084/create-account
```

## API Endpoints

- **POST /create-account**: Create a new account.
- **POST /get-accounts**: Retrieve all accounts.

## Testing

You can run load tests using the `hey` tool to measure performance and response times:

```bash
hey -n 1000 -c 100 -m POST -d '{"firstName":"John","lastName":"Doe"}' -H "Content-Type: application/json" http://127.0.0.1:8084/create-account
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
