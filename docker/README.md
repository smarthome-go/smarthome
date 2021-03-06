# Docker
## Example Docker-Compose File

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