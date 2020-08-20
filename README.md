[![GoDoc](https://godoc.org/github.com/bokwoon95/nusskylabx?status.svg)](https://godoc.org/github.com/bokwoon95/nusskylabx)
![CI](https://github.com/bokwoon95/nusskylabx/workflows/CI/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/bokwoon95/nusskylabx)](https://goreportcard.com/report/github.com/bokwoon95/nusskylabx)
[![Coverage Status](https://coveralls.io/repos/github/bokwoon95/nusskylabx/badge.svg?branch=master)](https://coveralls.io/github/bokwoon95/nusskylabx?branch=master)

# Setup Instructions
**NOTE:** The `$` sign that comes before a line of code denotes the prompt, and should not be copied and pasted as part of the command.

## Prequisites
- You will need to install Docker, Go and Node.
    - Make sure you can run `docker`, `go` and `npm` from the command line.
    - For Windows users, use Powershell for the command line (instead of Command Prompt).
- Install [air](https://github.com/cosmtrek/air) (or [modd](https://github.com/cortesi/modd)) for automatically recompiling and running your server on changes.
- Install an IDE or text editor of your choice. I recommend you use JetBrains's Goland IDE, which is free for students.

## Create .env and .air.conf
- Make a copy of `.env.default` as a new file called `.env` in project root directory.
    - You will have to fill in `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET`. You can either use the `nusskylab@gmail.com`'s google OAuth2 credentials or you can generate your own (next step).
- Make a copy `.air.conf.sample` as a new file called `.air.conf` in project root directory.
    - You will only need to do this if you are using [air](https://github.com/cosmtrek/air). You can also use [modd](https://github.com/cortesi/modd) instead.
    - Windows Users: You will have to change all occurrences of 'main' to 'main.exe' inside .air.conf file (under 'cmd' and 'bin')

## Get OAuth2 credentials
- Obtain Google Oauth2 Client ID and Client Secret from [here](https://developers.google.com/adwords/api/docs/guides/authentication#webapp).
    - Add the following authorized redirect urls to the oauth2 credential in your [google api credential console](https://console.developers.google.com/apis/credentials).
    ```
    http://localhost:8080/join/callback
    http://localhost:8080/login/callback
    ```
    - Update `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` in .env accordingly.

## Setup the database
```bash
# Create and run the postgres container
$ docker-compose up -d

# Make sure the image exists on your computer
$ docker images # you should see a nusskylab2/postgres_plpgsql_check_pgtap image

# Make sure a container instance of the image is running.
# If you don't see it, run `docker-compose up -d` again.
$ docker container ls # you should see a nusskylab2/postgres_plpgsql_check_pgtap container
```

## Start the application
```bash
# Load the sql views/functions/dummy data into the database.
$ go run cmd/loadsql/main.go

# Install npm packages
$ npm install

# Transpile the JavaScript files with webpack
$ npx webpack

# Finally make sure the program can run
$ go run main.go

# To avoid manually recompiling and running the server everytime you make a
# change, run the server with air or modd instead. Since you have already
# installed those programs, just run `air` or `modd` directly on the command line
# (without any arguments).
```

macOS Users: macOS imposes a [256 file limit](https://www.macobserver.com/tips/deep-dive/evade-macos-many-open-files-error-pushing-limits/) which prevents `air` from starting (since we have more than 256 files).
To get around it, temporarily increase the open file descriptor limit to 1024 with `ulimit -n 1024` before running `air` again. You can also combine it into one command and run `ulimit -n 1024; air` instead.

## Add yourself as an admin
### Manually
You can add yourself as an admin manually by visiting the server, clicking on the 'Create User' button on the navbar and entering following line into the text box:
```
,admin,<YOUR_NAME>,<YOUR_NUSNET_EMAIL>
```
Notice that the cohort field (the first field) is left blank so that the current cohort will automatically be used.
You can manually pass in the current cohort as well.

The downside of adding yourself manually is that you will have to re-add yourself again anytime you reset the database.

### As part of the sql data loading pipeline
The other way (which is recommended over the first) is to add your user details into an sql data script that will be run as part of `go run cmd/loadsql/main.go -clean` or when you run `go run cmd/loadsql/main.go` for the first time.
- Create a file called `temp.sql` inside the directory `app/db/data/`.
    - Any sql files starting with `temp` will be gitignored. You can put whatever developer specific data you want into this sql file without it being committed into the repository.
- Add the following contents into `temp.sql`:
```sql
-- app/db/data/temp.sql
DO $$ BEGIN
    -- app.create_user_role is a plpgsql function that was added into the database
    -- by running `go run cmd/loadsql/main.go`. You can see its source code
    -- definition inside the file app/db/functions/create_user_role.sql.
    PERFORM app.create_user_role(NULL, 'admin', '<YOUR_NAME>', '<YOUR_NUSNET_EMAIL>');
END $$;
```
- Do a clean database reset
```bash
$ go run cmd/loadsql/main.go -clean
```
You should now be able to login via NUS OpenID (the top right button on the navbar).
