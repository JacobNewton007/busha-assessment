# busha-assessment
> [Technologies](#technologies-used) &middot; [Testing Tools](#testing-tools) &middot; [Installations](#installations) &middot; [API Endpoints](#api-endpoints) &middot; [Tests](#tests) &middot; [Author](#author)


## Technologies Used

[golang]: (https://go.dev)

- [Golang](golang)
- [Postgres](Golang)
- [Redis](redis)



## Installations

#### Getting started

- You need to have Golang installed on your computer.

#### Clone

- Clone this project to your local machine `https://github.com/JacobNewton007/busha-assessment`

#### Setup

- Installing the project dependencies
  > Run the command below
  ```shell
  $ go mod download
  ```
- Start your server
  > run the command below
  ```shell
  $ docker-compose up --build -d
  ```
### Develop
- Use `http://localhost:4000/api/v1` as base url for endpoints
### Staging
- Use `https://busha-movie-api.onrender.com/api/v1` as base url for endpoints

## API Endpoints

| METHOD | DESCRIPTION                             | ENDPOINTS                 |
| ------ | --------------------------------------- | ------------------------- |
| POST   | Create comment                           | `/comments`    |
| GET   | Get a Characters                            | `/characters/`      |
| GET    | Get Movies                            | `/Moviess`|
| GET    | Get Comment by movie name                        | `/comments/:movie_name`|



## Author

   Kehinde Jacob
