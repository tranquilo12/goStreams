# GoStreams... or Lightning

## Project Structure
```bash
.
|-- cmd
|   |-- aggsPub.go
|   |-- aggsSub.go
|   |-- createTables.go
|   |-- root.go
|   |-- tickerTypes.go
|   `-- tickerVxes.go
|-- docker
|   |-- config
|   |   |-- grafana
|   |   |-- pgbouncer
|   |   |-- postgres
|   |   |-- provisioning
|   |   |-- redis
|   |   `-- config.env
|   |-- postgres
|   |   `-- data
|   |-- docker-compose.yml
|   |-- prom.env
|   `-- prometheus.yaml
|-- publisher
|   `-- publisher.go
|-- subscriber
|   `-- subscriber.go
|-- utils
|   |-- config
|   |   |-- config.go
|   |   `-- equities_list.csv
|   |-- db
|   |   |-- create_tables.go
|   |   |-- flatteners.go
|   |   |-- generate_urls.go
|   |   |-- inserts.go
|   |   |-- postgres.go
|   |   |-- redis.go
|   |   `-- requests.go
|   |-- mocks
|   |   `-- client.go
|   |-- responses
|   |   |-- responses.json
|   |   `-- tickers_response.json
|   `-- structs
|       `-- structs.go
|-- LICENSE
|-- README.md
|-- config.ini
|-- go.mod
|-- go.sum
|-- main.go
|-- make_ssh_tunnel.sh
|-- old_main.go
|-- old_main.txt
|-- pgbouncer.ini
`-- pgbouncer.pid

18 directories, 36 files 
```

## Docker Setup and Usage
To remain OS agnostic, postgres, pgbouncer and Redis are all docker containers. 

- Path to `docker-compose.yaml` - `./docker/docker-compose.yaml`
- `docker-compose.yaml` contains: 
    - [Postgres](#Postgres) - Store all data here. 
    - [Pgbouncer](#Pgbouncer) - To maintain 100's of potential postgres connections.
    - [Redis](#Redis) - To hold as many json structs in memory as possible before inserting into Postgres.
    - To start all these instances, go to the docker folder and run `docker-compose up`.
    - The containers might be running, but are they really functional? Check all with the script `check_docker_status.sh`
  
### Postgres 
- Postgres' settings can  be found in `./docker/config/postgres`
- If you don't want to use your local postgres instance, you can uncomment lines 25-37 in 
  `docker-compose.yaml`
- If you're using a local instance, make sure there's a database called "TimeScaleDB" in it.
- Once running, to connect to the docker postgres instance, try: 
  `psql postgresq://{username}:{password}@localhost:{port}/{database_name}`.
  
#### WSL2(Win 10) issues
- If you're looking to connect to your Windows Postgres instance, from your WSL2 instance... you're in for some effort.
- These are all things that must be done, I have not tested if some of these can be eliminated.
  - Make sure the Windows Firewall accepts incoming connections from the WSL2 instance. 
    [Link](https://serverfault.com/questions/1041981/how-can-i-connect-to-postgres-running-on-the-windows-host-from-inside-wsl2)
    - You may need to determine the Window's instance IP from the WSL2 instance: `cat /etc/resolv.conf`.
  - Find `pg_hba.conf` file location from the Windows psql instance by logging in and using the command `SHOW hba_file;`
      - Make sure the IPv4 local connection looks like: 
        `host    all             all             0.0.0.0/0               md`
      - Make sure the IPv6 local connections looks like:
        `host    all             all             ::/0                    md`
      - The above settings make the damn database accessible by everybody.
      - **CAREFUL WHEN IN PRODUCTION**.
  - In `postgresql.conf` found in `C:\Program Files\PostgreSQL\{version}\data\`: 
      - Change `listen_addresses` to `*`.
      - Make sure `port` is `5432`.
      - Make sure `max_connections` is reasonably high (e.g. 300).
  - Restart the Windows postgres instance, to ensure settings have taken effect. 
  - Using the Windows IP from the WSL2 instance, and FROM the WSL2 instance try: `psql -h {WINDOWS HOST IP FROM WSL2} -U postgres -lqt;`
  
### Pgbouncer
- Pgbouncer's settings can be found in `./docker/config/pgbouncer`
- Want to ensure that you always have a connection to your Postgres instance? 
  Never want to get a "too many connections to the database" error?
- Once running, this pgbouncer instance connects to the running Postgres instance by taking all the data through it port 6432 and forwarding it to port 5432. 
- To connect to the Postgres instance via the Pgbouncer instance try:  
  `psql postgresq://{username}:{password}@localhost:6432/{database_name}`

### Redis
- Redis settings can be found in `./docker/config/redis/conf`
- There are no databases here, and no usernames and password, it's a simple memory datastore.
- In the docker-compose.yaml file, redis port is redefined as `7000:6379`.
- To connect to the Redis docker instance, try:
  `redis-cli -p 7000`
  
### Useful commands
- (In Windows) `pg_ctl -D "P:\pg_db\data\" restart/start/stop;` - Start and stop postgres instance.

### Library Choices 
- Redigo exposed an interesting method of using redis with golang.
- Redis-go was the only library supporting rmq (redis message queue), so we'll have to switch to that.
