# FileStroageApplication


This project is a backend file storage service that uses **MySQL** for metadata storage and **local file system** for file persistence.  
The application is containerized using **Docker Compose** and requires environment configuration before running.

---

## Prerequisites

Make sure the following are installed on your system:

- Docker
- Docker Compose

---

## Project Setup

### Step 1: Start MySQL Using Docker Compose

From the **project root directory**, run:

```bash
MYSQL_ROOT_PASSWORD=<password> \
MYSQL_DATABASE=<databaseName> \
docker compose up -d


Replace:

<password> → MySQL root password

<databaseName> → Database name to be created
```



### Step 2: Configure Environment Variables
After starting the containers, configure the environment variables in both environment files.

```bash
Files to update:

cmd/config/.env
config/.env

Set the following values:
MYSQL_ROOT_PASSWORD=<password>
MYSQL_DATABASE=<databaseName>
⚠️ Important:
The values must match the ones used in the docker compose up command.
```

### Step 3: Configure File Storage Path

```bash
1. Create a directory to store uploaded files

Example:
mkdir /absolute/path/to/file-storage

2. Copy the absolute path

3. Set FILE_STORAGE_PATH in both environment files

Update both:

cmd/config/.env
config/.env

```

- To start the application:
  ``` 
  go run main.go
  ```

- To start the cosumer:
  ``` 
  cd ./cmd
  go run main.go go run main.go subscribe
  ```