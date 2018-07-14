# Foreign Currency BE Exercise
Example Rest API with Go, MySQL, and Docker

## Development
To run locally, import database from <b>forex_be/db/schema.sql</b> to your local mysql server.

and set the following environment variables:
```bash
DB_HOST
DB_PORT
DB_USER
DB_PASSWORD
DB_DATABASE
```
Run the server:
```bash
# build and run
go build
./forex-be
```

### Testing
```bash
# run test
go test
```

## Deployment
* Download/clone this repo, run docoker compose from root directory of the project
```
git clone https://github.com/dieehard/forex-be.git
cd forex-be/
````
* Edit docker-compose.yml file
```
# for mysql service
  environment:
    - MYSQL_DATABASE=exchange_rate
    - MYSQL_USER=user_forex
    - MYSQL_PASSWORD=secret
    - MYSQL_ROOT_PASSWORD=root
    - MYSQL_HOST=192.168.99.100
    - MYSQL_PORT=3306

```

```
# for api service
  environment:
    - DB_DATABASE=exchange_rate
    - DB_USER=user_forex
    - DB_PASSWORD=secret
    - DB_HOST=192.168.99.100
    - DB_PORT=3306
```
* Run compose-up
```
docker-compose up --build
```

## API and Documentation

List of API endpoint and documentation can be found here:

[forex-be API documentation](https://documenter.getpostman.com/view/2293128/RWMCt9av#5b4e768c-6cd2-48b7-a8ef-61854ebbf2c6)

## Database schema

Database design document can be found in <b>docs</b> directory