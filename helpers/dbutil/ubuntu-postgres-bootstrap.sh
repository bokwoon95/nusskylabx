#!/bin/bash
# How to get a postgres database running in virtualbox and accessible from the
# host computer. This assumes you are not already using docker to run the
# database.
# 1) Install virtualbox https://www.virtualbox.org/wiki/Downloads
# 2) Download the latest ubuntu ISO https://ubuntu.com/download/desktop
# 3) Setup a new Virtualbox VM with the Ubuntu ISO (2GB RAM & 8GB Disk, dynamically allocated. Adjust to taste)
# 4) Install Virtualbox guest additions for the Ubuntu VM
# Inside the VM, run this command in a terminal to download this script
#   wget https://raw.githubusercontent.com/bokwoon95/nusskylabx/master/helpers/dbutil/ubuntu-postgres-bootstrap.sh
# Then execute the downloaded script (or manually run each step below yourself)
#   chmod +x ubuntu-postgres-bootstrap.sh && ./ubuntu-postgres-bootstrap.sh
# There are additional steps to be completed once this script finishes. You can
# find them in the comments at the bottom.

# Add PostgreSQL APT repository https://www.postgresql.org/download/linux/ubuntu/
echo "deb http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -cs)-pgdg main" | sudo tee -a /etc/apt/sources.list.d/pgdg.list
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -

# apt-get update and install
sudo apt-get update
sudo apt-get install -y vim curl # Basic necessities
sudo apt-get install -y postgresql postgresql-common # Install PostgreSQL and its helper programs
sudo apt-get install -y git make build-essential libicu-dev postgresql-server-dev-all # Install plpgsql_check depedencies

# Create user and database
pg_version="$(pg_lsclusters | awk 'NR==2{print $1}')"
pg_cluster="$(pg_lsclusters | awk 'NR==2{print $2}')"
pg_status="$(pg_lsclusters | awk 'NR==2{print $4}')"
[ "$pg_status" != 'online' ] && sudo pg_ctlcluster "$pg_version" "$pg_cluster" start
echo "CREATE USER pg WITH ENCRYPTED PASSWORD 'pg';" | tee /dev/tty | sudo -u postgres psql
echo "ALTER USER pg WITH SUPERUSER;" | tee /dev/tty | sudo -u postgres psql
echo "CREATE DATABASE skylab_devx;" | tee /dev/tty | sudo -u postgres psql
echo "GRANT ALL PRIVILEGES ON DATABASE skylab_devx TO pg;" | tee /dev/tty | sudo -u postgres psql

# Install plpgsql_check extension
cd "$HOME" && git clone https://github.com/okbob/plpgsql_check
cd "$HOME/plpgsql_check" && make clean && sudo make install
cd "$HOME"

# Install pgTap extension
cd "$HOME" && git clone https://github.com/theory/pgtap
cd "$HOME/pgtap" && cpan TAP::Parser::SourceHandler::pgTAP && make clean && sudo make install
cd "$HOME"

# Start PostgreSQL server on boot
sudo systemctl enable postgresql

# Install go (used for running commands in cmd/)
sudo snap install go --classic

# Setup Virtualbox Port Forwarding
# Setup port forwarding for the VM: Settings > Network (you should see Adapter 1, NAT) > Advanced > Port Fowarding
# Then add a new rule
  # Name: PostgreSQL <anything you want>,
  # Protocol: TCP
  # Host IP: <blank>
  # Host Port: 5433
  # Guest IP: <blank>
  # Guest Port: 5432
# sudo vim /etc/postgresql/*/main/postgresql.conf
# Find the line in postgresql.conf with listen_addresses = 'localhost', uncomment it and change it to listen_addresses = '*'
# sudo vim /etc/postgresql/*/main/pg_hba.conf
# Find the line under the IPV4 section and add/change the line to: host all all all md5
# If anything goes wrong try to refer to this article
# https://improve-future.com/en/access-to-postgresql-in-virtualbox-guest-os-from-windows-host.html

# Download PgAdmin https://www.pgadmin.org/download/
# Add a new server
#   Name: skylab_devx
#   Host: 127.0.0.1 # 'localhost' will not work
#   Port: 5433
#   Maintainence Database: postgres # leave as postgres
#   Username: pg
#   Password: pg
# PgAdmin has a lousier UI than DBeaver, but supports step-by-step debugging of
# plpgsql functions with breakpoints (personally never used it, just relied on
# print statements).
#
# == OR ==
#
# Download DBeaver (recommended) https://dbeaver.io/
# New Database Connection
#   Host: localhost # or 127.0.0.1
#   Port: 5433
#   Database: skylab_devx
#   User: pg
#   Pass: pg
# DBeaver doesn't support step-by-step debugging of plpgsql functions like
# PgAdmin does, but can automatically generate ER diagrams from your database
# to visualize your database schema. Highly recommended.
#
# Try accessing the virtualbox database from pgadmin/dbeaver in the host OS

# Virtualbox guest additions may not work properly unless you have an
# up-to-date version of virtualbox. Assuming guest additions installed
# correctly, add current user to vboxsf user group. This step is needed because
# only users in the vboxsf user group may access the folders shared between VM
# OS and host OS
sudo adduser "$USER" vboxsf
# After this is done, go to shared folder settings and mount the repo you downloaded on your host OS
# onto another folder you can access in the Ubuntu VM. This will sync the folders.

gsettings set org.gnome.desktop.lockdown disable-lock-screen true # disable lockscreen
gsettings set org.gnome.desktop.screensaver lock-enabled false # disable lockscreen
