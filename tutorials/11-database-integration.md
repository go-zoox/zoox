# Tutorial 11: Database Integration

## 📖 Overview

Learn to integrate databases with Zoox applications. This tutorial covers database connections, query builders, ORM integration, and migration strategies for building data-driven applications.

## 🎯 Learning Objectives

- Connect to databases (MySQL, PostgreSQL, SQLite)
- Use query builders and raw SQL
- Implement ORM patterns
- Handle database migrations
- Optimize database performance

## 📋 Prerequisites

- Completed [Tutorial 01: Getting Started](./01-getting-started.md)
- Basic understanding of SQL and databases
- Familiarity with Go database/sql package

## 🚀 Getting Started

### Database Connection Manager

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "time"
    
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/lib/pq"
    _ "github.com/mattn/go-sqlite3"
    "github.com/go-zoox/zoox"
)

type DatabaseConfig struct {
    Driver   string
    Host     string
    Port     int
    User     string
    Password string
    Database string
    SSLMode  string
}

type DatabaseManager struct {
    db     *sql.DB
    config DatabaseConfig
}

func NewDatabaseManager(config DatabaseConfig) *DatabaseManager {
    return &DatabaseManager{
        config: config,
    }
}

func (dm *DatabaseManager) Connect() error {
    var dsn string
    
    switch dm.config.Driver {
    case "mysql":
        dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
            dm.config.User, dm.config.Password, dm.config.Host, dm.config.Port, dm.config.Database)
    case "postgres":
        dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
            dm.config.Host, dm.config.Port, dm.config.User, dm.config.Password, dm.config.Database, dm.config.SSLMode)
    case "sqlite3":
        dsn = dm.config.Database
    default:
        return fmt.Errorf("unsupported database driver: %s", dm.config.Driver)
    }
    
    db, err := sql.Open(dm.config.Driver, dsn)
    if err != nil {
        return err
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    // Test connection
    if err := db.Ping(); err != nil {
        return err
    }
    
    dm.db = db
    log.Printf("Connected to %s database", dm.config.Driver)
    return nil
}

func (dm *DatabaseManager) Close() error {
    if dm.db != nil {
        return dm.db.Close()
    }
    return nil
}

func (dm *DatabaseManager) GetDB() *sql.DB {
    return dm.db
}

// User model
type User struct {
    ID        int       `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// UserRepository handles user database operations
type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (ur *UserRepository) CreateTable() error {
    query := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username VARCHAR(255) UNIQUE NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`
    
    _, err := ur.db.Exec(query)
    return err
}

func (ur *UserRepository) Create(user *User) error {
    query := `
    INSERT INTO users (username, email, created_at, updated_at)
    VALUES (?, ?, ?, ?)`
    
    now := time.Now()
    result, err := ur.db.Exec(query, user.Username, user.Email, now, now)
    if err != nil {
        return err
    }
    
    id, err := result.LastInsertId()
    if err != nil {
        return err
    }
    
    user.ID = int(id)
    user.CreatedAt = now
    user.UpdatedAt = now
    return nil
}

func (ur *UserRepository) GetByID(id int) (*User, error) {
    query := `
    SELECT id, username, email, created_at, updated_at
    FROM users WHERE id = ?`
    
    user := &User{}
    err := ur.db.QueryRow(query, id).Scan(
        &user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("user not found")
        }
        return nil, err
    }
    
    return user, nil
}

func (ur *UserRepository) GetByUsername(username string) (*User, error) {
    query := `
    SELECT id, username, email, created_at, updated_at
    FROM users WHERE username = ?`
    
    user := &User{}
    err := ur.db.QueryRow(query, username).Scan(
        &user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("user not found")
        }
        return nil, err
    }
    
    return user, nil
}

func (ur *UserRepository) GetAll() ([]*User, error) {
    query := `
    SELECT id, username, email, created_at, updated_at
    FROM users ORDER BY created_at DESC`
    
    rows, err := ur.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []*User
    for rows.Next() {
        user := &User{}
        err := rows.Scan(
            &user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }
    
    return users, nil
}

func (ur *UserRepository) Update(user *User) error {
    query := `
    UPDATE users 
    SET username = ?, email = ?, updated_at = ?
    WHERE id = ?`
    
    user.UpdatedAt = time.Now()
    _, err := ur.db.Exec(query, user.Username, user.Email, user.UpdatedAt, user.ID)
    return err
}

func (ur *UserRepository) Delete(id int) error {
    query := `DELETE FROM users WHERE id = ?`
    _, err := ur.db.Exec(query, id)
    return err
}

func main() {
    app := zoox.New()
    
    // Database configuration
    config := DatabaseConfig{
        Driver:   "sqlite3",
        Database: "./users.db",
    }
    
    // Initialize database
    dbManager := NewDatabaseManager(config)
    if err := dbManager.Connect(); err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer dbManager.Close()
    
    // Initialize repository
    userRepo := NewUserRepository(dbManager.GetDB())
    if err := userRepo.CreateTable(); err != nil {
        log.Fatal("Failed to create table:", err)
    }
    
    // API endpoints
    app.Post("/users", func(ctx *zoox.Context) {
        var user User
        if err := ctx.BindJSON(&user); err != nil {
            ctx.JSON(400, map[string]string{"error": "Invalid request"})
            return
        }
        
        if err := userRepo.Create(&user); err != nil {
            ctx.JSON(500, map[string]string{"error": err.Error()})
            return
        }
        
        ctx.JSON(201, user)
    })
    
    app.Get("/users", func(ctx *zoox.Context) {
        users, err := userRepo.GetAll()
        if err != nil {
            ctx.JSON(500, map[string]string{"error": err.Error()})
            return
        }
        
        ctx.JSON(200, users)
    })
    
    app.Get("/users/:id", func(ctx *zoox.Context) {
        id := ctx.ParamInt("id")
        user, err := userRepo.GetByID(id)
        if err != nil {
            ctx.JSON(404, map[string]string{"error": err.Error()})
            return
        }
        
        ctx.JSON(200, user)
    })
    
    app.Put("/users/:id", func(ctx *zoox.Context) {
        id := ctx.ParamInt("id")
        
        var user User
        if err := ctx.BindJSON(&user); err != nil {
            ctx.JSON(400, map[string]string{"error": "Invalid request"})
            return
        }
        
        user.ID = id
        if err := userRepo.Update(&user); err != nil {
            ctx.JSON(500, map[string]string{"error": err.Error()})
            return
        }
        
        ctx.JSON(200, user)
    })
    
    app.Delete("/users/:id", func(ctx *zoox.Context) {
        id := ctx.ParamInt("id")
        
        if err := userRepo.Delete(id); err != nil {
            ctx.JSON(500, map[string]string{"error": err.Error()})
            return
        }
        
        ctx.JSON(200, map[string]string{"message": "User deleted successfully"})
    })
    
    // Web interface
    app.Get("/", func(ctx *zoox.Context) {
        html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>User Management</title>
            <style>
                body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
                .form { background: #f5f5f5; padding: 20px; margin: 20px 0; border-radius: 5px; }
                .form input { margin: 5px; padding: 8px; width: 200px; }
                .form button { padding: 8px 16px; margin: 5px; }
                .users { margin-top: 20px; }
                .user { border: 1px solid #ddd; padding: 10px; margin: 10px 0; border-radius: 5px; }
                .user button { margin: 5px; padding: 5px 10px; }
            </style>
        </head>
        <body>
            <h1>User Management</h1>
            
            <div class="form">
                <h3>Add User</h3>
                <input type="text" id="username" placeholder="Username">
                <input type="email" id="email" placeholder="Email">
                <button onclick="addUser()">Add User</button>
            </div>
            
            <div class="users">
                <h3>Users</h3>
                <button onclick="loadUsers()">Refresh</button>
                <div id="usersList"></div>
            </div>
            
            <script>
                async function addUser() {
                    const username = document.getElementById('username').value;
                    const email = document.getElementById('email').value;
                    
                    if (!username || !email) {
                        alert('Please fill in all fields');
                        return;
                    }
                    
                    try {
                        const response = await fetch('/users', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json',
                            },
                            body: JSON.stringify({ username, email })
                        });
                        
                        if (response.ok) {
                            document.getElementById('username').value = '';
                            document.getElementById('email').value = '';
                            loadUsers();
                        } else {
                            const error = await response.json();
                            alert('Error: ' + error.error);
                        }
                    } catch (error) {
                        console.error('Error adding user:', error);
                    }
                }
                
                async function loadUsers() {
                    try {
                        const response = await fetch('/users');
                        const users = await response.json();
                        
                        const usersList = document.getElementById('usersList');
                        usersList.innerHTML = '';
                        
                        users.forEach(user => {
                            const userDiv = document.createElement('div');
                            userDiv.className = 'user';
                            userDiv.innerHTML = \`
                                <strong>\${user.username}</strong> (\${user.email})
                                <br>Created: \${new Date(user.created_at).toLocaleDateString()}
                                <br>
                                <button onclick="deleteUser(\${user.id})">Delete</button>
                            \`;
                            usersList.appendChild(userDiv);
                        });
                    } catch (error) {
                        console.error('Error loading users:', error);
                    }
                }
                
                async function deleteUser(id) {
                    if (!confirm('Are you sure you want to delete this user?')) {
                        return;
                    }
                    
                    try {
                        const response = await fetch(\`/users/\${id}\`, {
                            method: 'DELETE'
                        });
                        
                        if (response.ok) {
                            loadUsers();
                        } else {
                            const error = await response.json();
                            alert('Error: ' + error.error);
                        }
                    } catch (error) {
                        console.error('Error deleting user:', error);
                    }
                }
                
                // Load users on page load
                loadUsers();
            </script>
        </body>
        </html>
        `
        ctx.HTML(200, html, nil)
    })
    
    log.Println("Server starting on :8080")
    log.Println("Database: SQLite (users.db)")
    log.Println("Interface: http://localhost:8080")
    
    app.Listen(":8080")
}
```

## 📚 Key Takeaways

1. **Database Connections**: Manage database connections efficiently
2. **Repository Pattern**: Separate data access logic
3. **Connection Pooling**: Optimize database performance
4. **Error Handling**: Handle database errors gracefully
5. **Migrations**: Manage database schema changes

## 🎯 Next Steps

- Learn [Tutorial 12: Caching Strategies](./12-caching-strategies.md)
- Explore [Tutorial 13: Monitoring & Logging](./13-monitoring-logging.md)
- Study [Tutorial 14: Testing Strategies](./14-testing-strategies.md)

---

**Congratulations!** You've mastered database integration in Zoox! 