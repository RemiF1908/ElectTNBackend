You need to make sure that the code compiles and runs without errors
Make the code as clean as possible
You need to clean up the codebase to make it more readable and maintainable
You can use docker compose to deploy the database
Add the health check endpoint in the openapi.yml and don't create a manual handler
Implement the authentification using gorilla/sessions, you should use the APP_MODE to use store.Options Secure true or false
You should rename election handler to handler
You should put the helper functions in a separate file
You should create a middleware for authentification, for normal user and for admin user, so the router are protected using this middleware
The admin route should be prefixed with /admin it makes it simple to use the admin middleware to secure the admin routes
You need to add integration tests in tests folder
Use logrus for logging and add an env variable for the log level