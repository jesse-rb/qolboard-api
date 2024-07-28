# qolboard-api

## Getting started with development

**Database & Supabase**

Supabase is used for authentication and a postgresql database setup, though you can feel free to use any postgresql database setup for local development.

*REQUIREMENTS*
+ Docker (If on macos, linux, or windows/wsl2 [Docker Desktop](https://www.docker.com/products/docker-desktop/) is a convenient way to to install docker and docker-compose binaies)
+ [Supabase CLI](https://supabase.com/docs/guides/cli/getting-started?platform=npx) (uses docker containers to setup a local supabase development environment) can be used conveniently through npx

*STEPS*
1. Start the local supabase environment (this will use the exisitng supabase configuration in `supabase` directory generated from the initial `npx supabase init` command) e.g.
    
    ```
    npx supabase start
    ```
2. You can access the local supabase environment dashboard on [localhost:54323](http://localhost:54323)

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
