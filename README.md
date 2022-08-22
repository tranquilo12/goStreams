# GoStreams... or Lightning

## Project Structure
```bash
.
├── LICENSE
├── README.md
├── main.go
├── cmd
│   ├── aggsPub.go
│   ├── aggsSub2.go
│   ├── createTables.go
│   ├── deleteFromDB.go
│   ├── deleteObjs.go
│   ├── refreshTickers.go
│   ├── questdbInsertAggs.go
│   ├── questdbRefreshTickers.go
│   ├── root.go
│   ├── tickerNews.go
│   └── tickerTypes.go
├── extras
│   ├── make_ssh_tunnel.sh
│   └── tmux-sessions.sh
├── subscriber
│   └── subscriber.go
│   └── kafkaSubscriber.go
├── publisher
│   └── publisher.go
│   └── kafkaPublisher.go
└── utils
    ├── config
    │   ├── config.go
    │   └── equities_list.csv
    ├── db
    │   ├── create_tables.go
    │   ├── dates.go
    │   ├── generate_urls.go
    │   ├── inserts.go
    │   ├── postgres.go
    │   ├── questDB.go
    │   ├── redis.go
    │   └── requests.go
    ├── mocks
    │   └── client.go
    └── structs
        └── structs.go
```

## Docker Setup and Usage
To remain OS-agnostic, postgres, pgbouncer and Redis are all docker containers. 

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

### Changes to postgresql.conf and pg_hba.conf
- postgresql.conf
  - usual location: /etc/postgresql/{version no}/main/postgresql.conf
  - changes: 
    - changed `data_directory` variable to `/mnt/p` to reflect the 2TB ext4 SDD.
    - changed `listen_addresses' to `*`
  
- pg_hba.conf
  - usual location: /etc/postgresql/{version no}/main/pg_hba.conf
  - changes: 
    - For local Unix domain socket: local, all, all, md5
    - For IPv4 local connections: host, all, all, 127.0.0.1/32, md5 
    
### Commands to bind windows drive to wsl2
- Make sure the drive is ext4 formatted. 
- Most of the info can be found [here](https://docs.microsoft.com/en-us/windows/wsl/wsl2-mount-disk).
  - Also check out [this](https://github.com/microsoft/WSL/issues/6319).
- From powershell:
  - Determine which drive is to be mounted: `wmic diskdrive list brief`
  - Then mount: `wsl --mount \\.\PHYSICALDRIVE{} --bare`
  
- From WSL:
  - Determine from `lsblk` the mount location. (will be something like `/dev/ss{}` ) - use `blkid --label 'PGDatabase'` 
  - Make a new directory in `/mnt/wsl/` and mount the `/dev/ss{}` to something else. 
    - `mkdir -p /mnt/wsl/PHYSICALDRIVE{} && mount /dev/sdd{} /mnt/wsl/PHYSICALDRIVE{}`
  - Now make sure it's usable by postgres
    - the new drive most prob has `root` ownership, so need to convert it to `postgres` level ownership
      - First get to root level: `su`
      - Then change ownership: `chown postgres /mnt/wsl/PHYSICALDRIVE{}`
      - Then login as postgres: `su postgres`
      - Make a new dir in `/mnt/wsl/PHYSICALDRIVE{}/data`, if it doesn't exist.
      - Make sure ` /etc/postgresql/13/main/postgresql.conf` mentions data path as `/mnt/wsl/PHYSICALDRIVE{}/data`
      - Shutdown postgres 13: `/usr/lib/postgresql/13/bin/pg_ctl stop`
      - InitDB: `/usr/lib/postgresql/13/bin/initdb -D /mnt/wsl/PHYSICALDRIVE{}/data`
      - Start postgres 13: `/usr/lib/postgresql/13/bin/pg_ctl start`
  - Need to change the default postgres password: 
    - change the pg_hba.conf (as in section above), from `md5` to `trust` for those local and IPv4 sections
    - restart postgresql: `/usr/lib/postgresql/13/bin/pg_ctl restart`
    - Get to root level: `su`
    - Then execute `sudo -u postgres psql -w -h 127.0.0.1 -p 5432` to login without password prompt
    - Then reset password using command `\password`
    - change the pg_hba.conf (as in section above), from `trust` to `md5` back again for those local and IPv4 sections
    - restart postgresql for changes to take effect. 
  
### Typical commands
Create Tables in a raw database:
- Ensure there's a database called 'polygonio"
- go install lightning; go createTables
  
- Publisher:
    - go install lightning; lightning aggsPub --dbtype={localdb/ec2db} --timespan minute --mult 1 --from 2020-01-01 --to 2021-05-01;
- Subscriber:
    - go install lightning; lightning aggsSub --dbtype={localdb/ec2db} --timespan minute --mult 1 --from 2020-01-01 --to 2021-05-01;

### Look at RabbitMQ statistics: 
- Go to ` http://localhost:15672`

### Installing TimescaleDB 
- Follow [this](https://docs.timescale.com/timescaledb/latest/how-to-guides/install-timescaledb/self-hosted/ubuntu/installation-apt-ubuntu/##apt-installation-ubuntu).

### If running pgbouncer on same instance
-  `sudo cp docker/config/pgbouncer/pgbouncer.ini /etc/pgbouncer/pgbouncer.ini`
-  `sudo cp docker/config/pgbouncer/userlist.txt /etc/pgbouncer/userlist.txt;`

### Create new docker volume attached to a current dir
- `docker volume create -d local-persist -o mountpoint=/mnt/wsl/PHYSICALDRIVE0/docker_data/ --name=polygonio-timescaledb-storage`

### Create Postgres dump from docker instance etc.
- `docker exec -i docker_timescale_1 /bin/bash -c "PGPASSWORD={} pg_dump --dbname=polygonio --host=host.docker.internal --port=5432 --username=postgres --schema='public' --table=public."aggregates_bars" --data-only" > C:\Users\SHIRAM\Documents\Numerai\aggregates_bars-2015-01-01-2021-06-05-dump.sql`
- Example for converting from sql to txt: `awk 'NR >= 57890000 && NR <= 57890010' /path/to/file > new_file.txt`
- or an easier way of doing things: `sudo docker exec -u postgres ${CONTAINER} psql -d ${DB} -c "COPY ${TABLE} TO STDOUT WITH CSV HEADER " > ${FILE}`

### AWS Firehose notes
- 