# USSD TCP Client Application

## Project Description
A robust Go-based USSD (Unstructured Supplementary Service Data) TCP client application designed for reliable communication with telecom servers. This application provides a flexible and secure method for establishing and maintaining USSD communication channels.

## 🚀 Features
- Secure TCP connection to USSD server
- Automatic session management
- Periodic Enquire Link to maintain connection
- Advanced logging system with file rotation
- Environment-based configuration
- Modular and extensible architecture

## 🛠 Prerequisites
- Go 1.21 or higher
- Access to a USSD server
- Basic understanding of TCP and USSD protocols

## 📦 Installation

### 1. Clone the Repository
```bash
git clone https://github.com/abeloha/USSDTCP.git
cd USSDTCP
```

### 2. Set Up Environment
```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your specific configuration
nano .env
```

### 3. Install Dependencies
```bash
go mod tidy
```

## 🔧 Configuration
Configuration is managed through the `.env` file. Key configurations include:

| Variable       | Description                     | Example                |
|---------------|--------------------------------|------------------------|
| SERVER_HOST   | USSD Server IP Address         | 0.0.3.0          |
| SERVER_PORT   | USSD Server Port               | 8000                   |
| USERNAME      | Authentication Username        | User123               |
| PASSWORD      | Authentication Password        | Pwd123               |
| CLIENT_ID     | Client Identifier              | 12345                   |
| LOG_PATH      | Directory for log files        | ./storage/logs         |
| DEBUG         | Enable debug mode              | false                  |

## 🏃 Running the Application
```bash
# Run the application
go run main.go

# Build for production
go build -o ussdtcp main.go
```

## 📝 Logging
- Logs are stored in the `storage/logs/` directory
- Daily log files are created with timestamp
- Supports multiple log levels: INFO, WARN, ERROR, DEBUG

## 🔒 Security Considerations
- Never commit sensitive information to version control
- Use `.env.example` as a template for configuration
- Ensure proper access controls on `.env` file
- Rotate credentials periodically

## 🐛 Troubleshooting
1. Verify server credentials
2. Check network connectivity
3. Review log files for detailed error information

### Common Issues
- Connection Timeout: Check server address and port
- Authentication Failure: Verify username and password
- Logging Problems: Ensure log directory is writable

## 🤝 Contributing
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Development Guidelines
- Follow Go coding standards
- Write unit tests for new features
- Update documentation
- Ensure code passes linting checks

## 📋 TODO
- [ ] Add unit tests
- [ ] Implement more robust error handling
- [ ] Create Docker support
- [ ] Add CI/CD pipeline

## 📜 License
Distributed under the MIT License. See `LICENSE` for more information.

## 📞 Contact
Abel Oha - [@abeloha](https://twitter.com/abeloha)

Project Link: [https://github.com/abeloha/USSDTCP](https://github.com/abeloha/USSDTCP)

---

**Disclaimer**: This project is for educational and development purposes. Always ensure compliance with telecom regulations and obtain necessary permissions.