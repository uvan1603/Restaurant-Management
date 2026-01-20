# Restaurant Management Backend (Go + MongoDB)

A production-style RESTful backend application built using **Go (Gin framework)** and **MongoDB**, designed to manage a complete restaurant workflow — from user authentication to menu management, order processing, and invoice generation.

This project focuses on **clean API design, scalable architecture, and real-world backend practices**, making it suitable for learning, interviews, and portfolio showcase.

---

## 🚀 Features

### 🔐 Authentication & Authorization

- User signup & login
- Secure password hashing
- JWT-based authentication
- Role-based access (Admin / Staff)

### 🍽️ Menu & Food Management

- Create and manage menus
- Add, update, and list food items
- Menu–Food relationships using MongoDB references

### 🪑 Table Management

- Create and manage restaurant tables
- Track table availability

### 🛒 Order Management

- Place orders for a table
- Add multiple items per order
- Track order status

### 🧾 Invoice Generation

- Generate invoices from completed orders
- Calculate totals dynamically
- MongoDB aggregation pipelines for reporting

---

## 🧱 Tech Stack

- **Language:** Go (Golang)
- **Framework:** Gin
- **Database:** MongoDB
- **Authentication:** JWT
- **Validation:** go-playground/validator
- **Environment Management:** godotenv

---

## 📁 Project Structure

```
.
├── controllers/        # Request handlers (business logic)
├── routes/             # API route definitions
├── models/             # MongoDB models
├── services/           # Helper and shared logic
├── middlewares/        # Auth & request middleware
├── database/           # MongoDB connection
├── utils/              # Utility functions
├── main.go             # Application entry point
├── go.mod
└── go.sum
```

---

## 📌 API Modules

The backend is organized into **7 controllers**:

1. **User Controller** – Signup, login, authentication
2. **Menu Controller** – Menu CRUD operations
3. **Food Controller** – Food item management
4. **Table Controller** – Restaurant table handling
5. **Order Controller** – Order creation and tracking
6. **OrderItem Controller** – Items within an order
7. **Invoice Controller** – Invoice generation & summaries

---

## 🔁 Typical Application Flow

1. User signs up and logs in
2. Admin creates menus and food items
3. Tables are created and managed
4. Orders are placed for tables
5. Order items are added
6. Invoice is generated after order completion

---

## 🧪 How to Run Locally

### Prerequisites

- Go 1.21+
- MongoDB (local or cloud)

### Steps

```bash
# Clone the repository
git clone <repo-url>
cd restaurant-backend

# Install dependencies
go mod tidy

# Setup environment variables
cp .env.example .env

# Run the server
go run main.go
```

The server will start on:

```
http://localhost:8080
```

---

## 🌱 Environment Variables

```
PORT=8080
MONGODB_URI=mongodb://localhost:27017
DB_NAME=restaurant_db
JWT_SECRET=your_secret_key
```

---

## 🧠 Key Learnings

- Designing scalable REST APIs in Go
- JWT authentication and middleware design
- MongoDB aggregation pipelines
- Structuring Go applications for maintainability
- Handling real-world backend workflows

---

## 📈 Future Enhancements

- Docker containerization
- API documentation with Swagger
- Role-based permissions expansion
- Caching with Redis
- Deployment to cloud platforms

---

## 👨‍💻 Author

**Uvan N**  
Full Stack Developer

---

## ⭐ Why This Project?

This project was built to simulate **real-world backend systems**, focusing on clean architecture, scalability, and interview-ready concepts rather than just CRUD operations.

If you are preparing for backend interviews or learning Go for production use, this repository serves as a strong reference.
