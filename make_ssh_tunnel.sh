#!/usr/bin/env bash
# ssh -L 5433:localhost:6432 ec2-18-195-240-223.eu-central-1.compute.amazonaws.com -i "C:\\Users\\SHIRAM\\Documents\\streams\\ssl_certs\\new_ts_pair.pem";
# ssh -L 5433:localhost:6432 ec2-18-195-240-223.eu-central-1.compute.amazonaws.com -i "/mnt/c/Users/SHIRAM/Documents/streams/ssl_certs/new_ts_pair.pem";
 ssh -L 5433:localhost:6432 -N -f ubuntu@ec2-18-195-240-223.eu-central-1.compute.amazonaws.com -i "~/.ssh/new_ts_pair.pem";
