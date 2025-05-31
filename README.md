# Vision Backend

*A comprehensive backend service for the FarmVision mobile application, providing agricultural community features with
modern technology stack.*

🌟 **Features**

- **User Authentication**
    - SMS verification code sending
    - User registration and login
    - JWT-based authentication

- **Search & AI**
    - Keyword-based search functionality
    - AI-powered Q&A system

- **Community Features**
    - Community management
    - Post creation and management
    - Comment system
    - Like/Unlike functionality
    - Voting system

- **Content & Media**
    - Video feed system
    - Content recommendation

🚀 **Quick Start**

1. **Clone the repository:**

  ```sh
  git clone https://github.com/ShawnJeffersonWang/vision.git
  cd vision
  ```

2. 🐳 **Docker Setup:**

  ```sh
  docker compose -f docker-compose.yml -p vision up -d
  ```

🛠️ **Tech Stack**

- **Backend Framework**: Go with Gin web framework
- **Database**: MySQL with GORM ORM
- **Cache**: Redis for session management and caching
- **Authentication**: JWT (JSON Web Tokens)
- **Containerization**: Docker & Docker Compose
- **Architecture**: RESTful API design

📈 **Performance Features**

- **Redis Caching**: Implements caching for frequently accessed data
- **Database Optimization**: Proper indexing and query optimization
- **JWT Authentication**: Stateless authentication for scalability
- **Middleware**: Request logging, CORS, rate limiting
- **Container Optimization**: Multi-stage Docker builds for minimal image size
- **Service Isolation**: Microservices architecture with Docker Compose

🙏 **Acknowledgments**
Thanks to the Go community for excellent libraries
Gin framework for fast HTTP routing
GORM for elegant database operations
Redis for high-performance caching

⭐ **Star this repository if you find it helpful!**
