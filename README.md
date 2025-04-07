# D-CAS (Distributed Content-Addressable Storage)

D-CAS is a distributed content-addressable storage system that allows you to store and retrieve files using their content hash or user-defined code words.

## Features

- **Content-Addressable Storage**: Files are stored and retrieved using their SHA-256 hash
- **Code Word Support**: Optionally assign user-friendly code words to files
- **Web Interface**: Modern, responsive UI for file uploads and searches
- **Multiple Storage Backends**: Support for local filesystem and MinIO
- **RESTful API**: Easy integration with other applications

## Getting Started

### Prerequisites

- Go 1.21 or later
- MinIO (optional, for distributed storage)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/D-CAS.git
cd D-CAS
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o bin/D-CAS.exe main.go
```

### Configuration

The application can be configured using environment variables:

```bash
# Server configuration
PORT=8080                    # Web server port
NODE_PORT=8081              # Node communication port
MODE=web                    # Operation mode: web or node

# Storage configuration
STORAGE_TYPE=local          # Storage type: local or minio
STORAGE_ENDPOINT=           # MinIO endpoint (required for minio storage)
STORAGE_ACCESS_KEY=         # MinIO access key
STORAGE_SECRET_KEY=         # MinIO secret key
STORAGE_BUCKET=            # MinIO bucket name
STORAGE_USE_SSL=false      # Use SSL for MinIO connection
```

### Running the Application

1. Start the web server:
```bash
./bin/D-CAS.exe -mode=web -port=8080
```

2. Open your browser and navigate to `http://localhost:8080`

## API Endpoints

### Upload File
```
POST /api/upload
Content-Type: multipart/form-data

Parameters:
- file: The file to upload
- code_word: Optional code word for the file
```

### Search Files
```
GET /api/search?q=<hash_or_code_word>
```

### Download File
```
GET /api/download/<hash>
```

## Development

### Project Structure

```
D-CAS/
├── bin/                    # Compiled binaries
├── config/                 # Configuration package
├── storage/                # Storage interface and implementations
├── web/                    # Web server and static files
│   └── static/            # Frontend assets
├── main.go                 # Application entry point
└── README.md              # This file
```

### Adding New Features

1. Create a new branch for your feature
2. Make your changes
3. Add tests if applicable
4. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request