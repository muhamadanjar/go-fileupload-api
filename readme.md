# Fileupload Api
### Install Gin framework
go get -u github.com/gin-gonic/gin

### Install GORM and PostgreSQL driver
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres

### Install UUID package
go get -u github.com/google/uuid

### Install dotenv for configuration
go get -u github.com/joho/godotenv

## Database

### Access PostgreSQL
psql -U postgres

### Create the database
CREATE DATABASE fileuploader;


### Run the application
```bash
go run cmd/api/main.go
```

## Build

### Build the application
```bash
go build -o fileuploader cmd/api/main.go
```

### Run the executable
```bash
./fileuploader
```


# Author: Muhamad Anjar
