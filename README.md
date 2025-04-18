# 🚀 GoDrop

**GoDrop** is a lightweight, no-login file sharing system built with **Go** and powered by **Cloudflare R2** (or any S3-compatible storage). With a minimal interface and no authentication barrier, sharing files becomes instant and hassle-free.

[Demo]([https://example.com](https://godrop.onrender.com/))

---

## ✨ Features

- 🌐 Clean web UI for uploading and downloading files
- ☁️ Supports S3-compatible storage (Cloudflare R2, MinIO, AWS S3, etc.)
- ⚙️ Fully configurable via `.env` or environment variables
- 🧠 Optionally assign user-friendly **code words** to files for easy reference

---

## 🚀 Getting Started

### 📋 Prerequisites

- [Go 1.21+](https://golang.org/dl/)

---

### 📥 Installation

1. **Clone the repository:**

```bash
git clone https://github.com/yourusername/GoDrop.git
cd GoDrop
```

2. **Install dependencies:**

```bash
go mod download
```

3. **Build the application:**

```bash
go build -o bin/godrop main.go
```

> 💡 You can rename the binary to whatever suits your workflow (`godrop`, `dropper`, etc.)

---

### ⚙️ Configuration

Set the following environment variables, either in your shell or in a `.env` file in the project root:

```env
# Server configuration
PORT=8080                    # Web server port
NODE_PORT=8081               # (Optional) Node communication port
MODE=web                     # Run mode: 'web' or 'node'

# Storage configuration
STORAGE_TYPE=s3              # Only 's3' is supported
STORAGE_ENDPOINT=            # Cloudflare R2 or other S3-compatible endpoint
STORAGE_ACCESS_KEY=          # Your storage access key
STORAGE_SECRET_KEY=          # Your storage secret key
STORAGE_BUCKET=              # Target bucket name
STORAGE_USE_SSL=true         # Set false if using HTTP
```

---

### ▶️ Running the Application

1. **Start the web server:**

```bash
go run main.go
```

2. **Open your browser:**

Navigate to [http://localhost:8080](http://localhost:8080)

> ✅ You can now upload and share files via the intuitive UI.

---

## 👨‍💻 Contributing

We welcome contributions to improve GoDrop!

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/your-feature`
3. Commit your changes: `git commit -m "Add your message"`
4. Push to the branch: `git push origin feature/your-feature`
5. Create a Pull Request

---

## 📄 License

This project is licensed under the **MIT License** — see the [LICENSE](./LICENSE) file for details.

---
