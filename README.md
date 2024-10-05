# qolboard-api

## Getting started with development

**Database & Supabase**

**Note:**
The local supabase postgresql database setup is not required but is convenient, feel free to use any postgresql database setup that you prefer, you only need to set the connection details in the your `.env` file **HOWEVER**, this project does make use of **Supabase GoTrue API** for authentication, so it is still **recommended** to go through with the **local Subapase environment setup**.

*REQUIREMENTS*
+ Docker (If on macos, linux, or windows/wsl2 [Docker Desktop](https://www.docker.com/products/docker-desktop/) is a convenient way to to install docker and docker-compose binaies)
+ [Supabase CLI](https://supabase.com/docs/guides/cli/getting-started?platform=npx) (uses docker containers to setup a local supabase development environment) can be used conveniently through npx

*STEPS*
1. Start the local supabase environment (this will use the exisitng supabase configuration in `supabase` directory generated from the initial `npx supabase init` command) e.g.
    
    ```
    npx supabase start
    ```

    Note the output of the above command to find `SUPABASE_...` and `DATABASE_...` values to environment variables (see `.env.example`) that need to be set in your local `.env` file (these will be needed if you would like the app to interact with the local supabase environment)

2. You can access the local supabase environment dashboard on [localhost:54323](http://localhost:54323). run `npx supabase status` for a reminder of local supabase serviecs ports and secrets

easy local testing supabase links:
+ [supabase web dashboard](http://127.0.0.1:54323)
+ [mock email service (Inbucket)](http://127.0.0.1:54324)

**Golang**

*REQUIREMENTS*
+ A running database to connect to
+ A `.env` with environment variables that suits your development environment (KEEP EXCLUDED FROMN VERSION CONTROL)

    Feel free to use the following command from the project root direcoty as a starting point for your `.env` file, the provided example `.env.example` (For `SUPABASE_...` and `DATABASE_...` env values see output of `npx supabase init` command documented above)
    ```
    cp .env.example .env
    ```

*STEPS*
1. From the project root directory, run the following command to start the web app
    ```
    go run main.go
    ```
