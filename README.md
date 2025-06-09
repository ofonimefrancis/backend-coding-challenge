# Backend Senior Coding Challenge 🍿

Welcome to our Movie Rating System Coding Challenge! We appreciate you taking
the time to participate and submit a coding challenge! 🥳

In this challenge, you'll be tasked with designing and implementing a robust
backend system that handles user interactions and provides movie ratings. We
don't want to check coding conventions only; **we want to see your approach
to systems design!**

**⚠️ As a tech-agnostic engineering team, we ask you to pick the technologies
you are most comfortable with and those that will showcase your strongest
performance. 💪**

## ✅ Requirements

- [x] The backend should expose RESTful endpoints to handle user input and
  return movie ratings.
- [x] The system should store data in a database. You can use any existing
  dataset or API to populate the initial database.
- [x] Implement user endpoints to create and view user information.
- [x] Implement movie endpoints to create and view movie information.
- [x] Implement a rating system to rate the entertainment value of a movie.
- [x] Implement a basic profile where users can view their rated movies.
- [x] Include unit tests to ensure the reliability of your code.
- [x] Ensure proper error handling and validation of user inputs.

## ✨ Bonus Points

- [ ] Implement authentication and authorization mechanisms for users.
- [ ] Provide documentation for your API endpoints using tools like Swagger.
- [x] Implement logging to record errors and debug information.
- [x] Implement caching mechanisms to improve the rating system's performance.
- [ ] Implement CI/CD quality gates.

## 📋 Evaluation Criteria

- **Systems Design:** We want to see your ability to design a flexible and
  extendable system. Apply design patterns and software engineering concepts.
- **Code quality:** Readability, maintainability, and adherence to best
  practices.
- **Functionality:** Does the system meet the requirements? Does it provide
  movie
  ratings?
- **Testing:** Adequate test coverage and thoroughness of testing.
- **Documentation:** Clear documentation for setup, usage, and API endpoints.

## 📐 Submission Guidelines

- Fork this GitHub repository.
- Commit your code regularly with meaningful commit messages.
- Include/Update the README.md file explaining how to set up and run your
  backend, including any dependencies.
- Submit the link to your repository.

## 🗒️ Notes

- You are encouraged to use third-party libraries or frameworks to expedite
  development but be prepared to justify your choices.
- Feel free to reach out if you have any questions or need clarification on the
  requirements.
- Remember to approach the challenge as you would a real-world project, focusing
  on scalability, performance, and reliability.

## 🚀 Running the Application

### Using Docker (Recommended)

1. Make sure you have Docker and Docker Compose installed on your system
2. Clone the repository:
   ```bash
   git clone <repository-url>
   cd <repository-name>
   ```
3. Build and start the containers:
   ```bash
   docker-compose up --build
   ```
   This will start:
   - PostgreSQL database on port 5432
   - Redis cache on port 6379
   - The application on ports 8080 (main API) and 8081 (metrics)

4. To stop the application:
   ```bash
   docker-compose down
   ```

### Running Locally

1. Prerequisites:
   - Go 1.21 or later
   - PostgreSQL 15
   - Redis 7

2. Set up the database:
   ```bash
   # Create the database
   createdb test_db
   
   # Run migrations (if using golang-migrate)
   migrate -path ./internal/platform/repository/migrations -database "postgresql://postgres:postgres@localhost:5432/test_db?sslmode=disable" up
   ```

3. Set up environment variables:
   ```bash
   export POSTGRESQL_DSN="host=localhost dbname=test_db user=postgres password=postgres sslmode=disable"
   export REDIS_DSN="localhost:6379"
   export POSTGRES_HEALTH_CHECK="true"
   export REDIS_HEALTH_CHECK="true"
   ```

4. Run the application:
   ```bash
   go run cmd/api/main.go
   ```

The application will be available at:
- Main API: http://localhost:8080
- Metrics: http://localhost:8081

### Health Checks

The application includes health check endpoints:
- API Health: http://localhost:8080/health
- Metrics: http://localhost:8081/metrics

## 🤔 What if I don't finish?

Part of the exercise is to see what you prioritize first when you have a limited
amount of time. For any unfinished tasks, please do add `TODO` comments to
your code with a short explanation. You will be given an opportunity later to go
into more detail and explain how you would go about finishing those tasks.
