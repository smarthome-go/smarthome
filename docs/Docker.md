# Docker
*Unfortunately, installation via Docker is currently only supported for AMD64 devices. However, support for other architectures, such as ARM is planned.*

## Installation
### Instructions
1. Make sure an up-to-date version of both Docker and docker-compose is installed on the target machine.
2. Copy the contents of the *docker-compose.yml* file or [download](../docker-compose.yml) it into your target directory.
3. Tweak the configuration parameters to fit your needs.
You should also learn about [environment variables](###configuration-via-environment-variables).
4. *Optional*: prepare a valid `setup.json` file.
5. Run `docker-compose up [-d]` in order to start the service.

*Note*: It is recommended to omit the `-d` flag the first time you start the service.
This allows you to follow the application logs in real-time, thus potential debugging becomes easier.

### Configuration via environment Variables
#### Database configuration
*Note*: Most of these values must be equal to the ones you configured in your Database.

##### `SMARTHOME_DB_DATABASE`
Specifies the name of the database Smarthome should connect to.  

##### `SMARTHOME_DB_HOSTNAME`
Specifies the hostname or *IP* address of the Database server.  
*Recommended*: In case of your database server is spun up in the same *docker-compose* file, you can use the `container_name` of your database container as an alias for the exact IP.

##### `SMARTHOME_DB_PORT`
Specifies the port which is used to connect to the database server.  
*Note*: Must be a *numeric* value, otherwise Smarthome will attempt to use the value from `config.json`.

##### `SMARTHOME_DB_USER`
Specifies the *Mysql* user which Smarthome will use for all connections.

*Note*: Correct access and permission sesstings inside your Database are required for a working connection.  
If using the *docker-compose* template from below, permissions won't likely be an issue.

##### `SMARTHOME_DB_PASSWORD`
Specifies the password which Smarthome will use to authenticate connections as `SMARTHOME_DB_USER`.

#### Smarthome configuration
##### `SMARTHOME_PORT`
Specifies on which port Smarthome should listen on.  

*Note*: In case of docker, this is the *internal* port, which must first be *exposed* using the `ports:` section in the compose file.
You need to pay close attention that the right hand side of the `ports:` section is equal to `SMARTHOME_PORT`.
```yml
    ports:
      - 8123:80 # Right side must be equal to `SMARTHOME_PORT`
                # Left side specifies the port on your host server on which you can connect
```

##### `SMARTHOME_LOG_LEVEL`
This value specifies which log level Smarthome should use.  
Valid log levels are listed below using low to high order.  
*Note*: A *low* log level means a lot of information output whilst a *high* log level will mean only important events will be logged.

- `TRACE`
- `DEBUG`
- `INFO`
- `WARN`
- `ERROR`
- `FATAL`

If an invalid log level is specified, the service defaults to `TRACE` and continues startup.

##### `SMARTHOME_ENV_PRODUCTION`
If set to `TRUE`, Smarthome will use some *production* optimizations which improve speed and security.
For non-development applications, it is highly recommended setting this value to `TRUE`.  
Valid values are `TRUE` and `FALSE`.

##### `SMARTHOME_SESSION_KEY`
Specifies a manual key which is used for session encryption. For larger installations, setting this value to a randomly generated string is recommended.  
If set, this will prevent a log-out of all users in case the server is restarted.  
*Note*: This setting is only effective when using `production` mode. During `development` mode, the server will use a static, empty string for encryption.  
During production, if you specify an empty value, Smarthome will generate a 32-bit random key to maintain security.

##### `SMARTHOME_ADMIN_PASSWORD`
During the first start, a `admin` user is created. It will receive this specified password if set.  
Otherwise, the default password `admin` is selected.  
However, you can always change a user's password using the Web-UI at a later point.
*Note*: It is highly recommended setting a secure password before starting the server.

#### Example Docker-Compose File

```yml
version: '3.7'
services:
  smarthome:
    image: mikmuellerdev/smarthome:0.0.29
    container_name: smarthome
    hostname: smarthome
    depends_on:
      - smarthome-mariadb
    environment:
      # Timezone is required for accurate scheduling
      - TZ=Europe/Berlin
      
      # Smarthome database configuration
      - SMARTHOME_DB_DATABASE=smarthome
      - SMARTHOME_DB_HOSTNAME=smarthome-db
      - SMARTHOME_DB_PORT=3306
      - SMARTHOME_DB_USER=smarthome
      - SMARTHOME_DB_PASSWORD=password
      
      # Smarthome sever configuration
      - SMARTHOME_PORT=80                           # Default is 8082
      - SMARTHOME_LOG_LEVEL=TRACE                   # Default is INFO
      - SMARTHOME_ENV_PRODUCTION=TRUE               # Default is TRUE
      - SMARTHOME_SESSION_KEY=random_key_here       # Default uses a random key generated by Smarthome
      - SMARTHOME_ADMIN_PASSWORD=password           # Only set on first start
    ports:
      - 8123:80 # Right side must be equal to `SMARTHOME_PORT`
    restart: unless-stopped
    # volumes:
      #  - /path/to/data:/app/data

  smarthome-mariadb:
    image: mariadb
    container_name: smarthome-db
    hostname: smarthome-db
    environment:
      # Timezone is required for created-at dates and logging 
      - TZ=Europe/Berlin

      # Config for access from Smarthome
      - MARIADB_DATABASE=smarthome
      - MARIADB_USER=smarthome
      - MARIADB_PASSWORD=password

      # Root password for maintenance only
      - MARIADB_ROOT_PASSWORD=password
    restart: unless-stopped
    # volumes:
      # - /path/to/db:/var/lib/mysql
```
