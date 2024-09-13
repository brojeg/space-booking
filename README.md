# Coding Challenge
Imagine it’s 2049 and you are working for a company called SpaceTrouble that sends people to different places in our solar system. 
You are not the only one working in this industry. 
Your biggest competitor is a less-known company called SpaceX. 
Unfortunately, you both share the same launchpads and you cannot launch your rockets from the same place on the same day. 

There is a list of available launchpads and your spaceships go to places like
- Mars,
- Moon,
- Pluto,
- Asteroid
- Belt,
- Europa,
- Titan,
- Ganymede.

Additional requirements
- Every day you change the destination for all the launchpads.
- Every day of the week from the same launchpad has to be a “flight” to a different place.

Information about available launchpads and upcoming SpaceX launches you can find by SpaceX API: https://github.com/r-spacex/SpaceX-API

Your task is to create an API that will let your consumers book tickets online.

In order to do that you have to create 2 endpoints:
- Endpoint to book a ticket where a client sends data like:
  - First Name
  - Last Name
  - Gender
  - Birthday
  - Launchpad ID
  - Destination ID
  - Launch Date

You have to verify if the requested trip is possible on the day from the provided launchpad ID and does not overlap with SpaceX launches, if that’s the case then your flight is cancelled.

- Endpoint to get all created Bookings.

Extra points:
- When you use docker/docker-compose to run the project.
- When you write unit/functional tests.
- When you create an endpoint to delete a booking.

Technical requirements:
- Please, use Golang and Postgres.
- Please, use GitHub or Bitbucket.
- Commit your changes often. Do not push the whole project in one commit.

## MakeFile

run all make commands with clean tests
```bash
make all build
```

build the application
```bash
make build
```

DB migrations new\up\down
```bash
make migrate
```

run the application
```bash
make run
```

Create DB container and Run the app
```bash
make docker-run
```

Shutdown DB container
```bash
make docker-down
```

live reload the application
```bash
make watch
```

run the test suite
```bash
make test
```

clean up binary from the last build
```bash
make clean
```

