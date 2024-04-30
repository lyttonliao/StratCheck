#!/bin/bash
set -eu

# ==================================================================================== #
# VARIABLES
# ==================================================================================== #

TIMEZONE=America/New_York

USERNAME=stratcheck

# Prompt to enter a password for PostgreSQL stratcheck user
read -p "Enter password for stratcheck DB user: " DB_PASSWORD

export LC_ALL=en_US.UTF-8

# ==================================================================================== #
# SCRIPT LOGIC
# ==================================================================================== #

# Enable the "universe" repository
add-apt-repository --yes universe

# Update all software packages. Using the --force-confnew flag means that config files will be replaced if newer ones are available.
apt update
apt --yes -o Dpkg::Options::="--force-confnew" upgrade

# Set the system timezone and install all locales.
timedatectl set-timezone ${TIMEZONE}
apt --yes install locales-all

# Add the new user (and give them sudo privileges)
if ! id "${USERNAME}" >/dev/null 2>&1; then
    echo "USER '${USERNAME}' does not exist. Creating user..."

    useradd --create-home --shell "/bin/bash" --groups sudo "${USERNAME}"

    # Force a password to be set for the new user the first time they log in.
    passwd --delete "${USERNAME}"
    chage --lastday 0 "${USERNAME}"

    # Copy the SSH keys from the root user to the new user.
    rsync --archive --chown=${USERNAME}:${USERNAME} /root/.ssh /home/${USERNAME}
else
    echo "USER '${USERNAME}' already exists. Skipping user creation."
fi

# Configure the firewall to allow SSH, HTTP and HTTPS traffic.
ufw allow 22
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# Install fail2ban.
apt --yes install fail2ban

# Install the migrate CLI tool.
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz
mv migrate /usr/local/bin/migrate


# Install PostgreSQL
apt --yes install postgresql

# Set up the stratcheck DB and create a user account with the password entered earlier
sudo -i -u postgres psql -c "CREATE DATABASE stratcheck"
sudo -i -u postgres psql -d stratcheck -c "CREATE EXTENSION IF NOT EXISTS citext"
sudo -i -u postgres psql -d stratcheck -c "CREATE ROLE stratcheck WITH LOGIN PASSWORD '${DB_PASSWORD}'"


# Add a DSN for connecting to the stratcheck database to teh system-wide environment variables in the /etc/environment file
echo "STRATCHECK_DB_DSN='postgres://stratcheck:${DB_PASSWORD}@localhost/stratcheck'" >> /etc/environment

# Install Caddy (see https://caddyserver.com/docs/install#debian-ubuntu-raspbian).
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy

echo "Script complete! Rebooting..."
reboot
