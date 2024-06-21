# Harbor - Docker Container Manager

## Features
- Manage your containers
- Execute commands inside your containers
- Manage your images
- Easily create a mysql database
- Pull images
- Create containers from an image
- Analytics
- Cleanup docker with prune

![containers](https://firebasestorage.googleapis.com/v0/b/portfolio-21d2a.appspot.com/o/harbor2.png?alt=media&token=ece6675e-df23-4bbc-a792-b29ef8094b84)
![images](https://firebasestorage.googleapis.com/v0/b/portfolio-21d2a.appspot.com/o/harbor3.png?alt=media&token=ece6675e-df23-4bbc-a792-b29ef8094b84)
![execute](https://firebasestorage.googleapis.com/v0/b/portfolio-21d2a.appspot.com/o/harbor1.png?alt=media&token=ece6675e-df23-4bbc-a792-b29ef8094b84)
![analytics](https://firebasestorage.googleapis.com/v0/b/portfolio-21d2a.appspot.com/o/harbor4.png?alt=media&token=ece6675e-df23-4bbc-a792-b29ef8094b84)


## .env
Create a .env in the root dir with

```json
SECRET_KEY="secretkey"
```

For the SECRET_KEY any value will do. This value will be used by the cookieStore.

## Run with docker

### Requirements
- [Docker](https://docs.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/install/)

### Run it
```
docker compose up -d
```

## Run without docker

### Requirements
- [Go](https://go.dev/doc/install)
- [Air](https://github.com/cosmtrek/air)


### Windows/Linux/MacOS
In the .air.toml file change the "./tmp/main" to "./tmp/main.exe" for windows, for Linux and MacOS no changes have to be made.

### Run it
``` bash
air
```

## Making changes

### Requirements
- [Tailwind CSS CLI](https://tailwindcss.com/docs/installation)

### Tailwind CSS
When making changes to the code with tailwind make sure to run.
```bash
tailwindcss -i css/input.css -o css/output.css --watch
```
