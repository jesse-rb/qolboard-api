# qolboard-api

## Getting started with development

**Database**

*REQUIREMENTS*
+ You can use this below setup, or feel free to use any postgresql database setup for your development environment

*STEPS*
1. Create a development database using the provided `Dockerfile` using the following command in the project root directory
    ```
    docker build -t qolboard-postgres ./
    ```
2. Ensure the development database is running
    ```
    docker run -d --name qolboard-postgres-container -p 5432:5432 qolboard-postgres
    ```

**Golang**

*REQUIREMENTS*
+ A running database to connect to
+ A `.env` with environment variables that suits your development environment

    Feel free to use the following command from the project root direcoty as a starting point for your `.env` file
    ```
    cp .env.example .env
    ```

*STEPS*
1. From the project root directory, run the following command to start the web app
    ```
    go run main.go
    ```
