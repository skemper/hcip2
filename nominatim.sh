sudo mkfs -t ext4 /dev/nvme1n1
sudo mkdir /work
sudo mount /dev/nvme1n1 /work
sudo echo "/dev/nvme1n1 /work ext4 defaults 0 0" >>/etc/fstab


sudo apt-get update -qq
sudo apt-get install -y build-essential cmake g++ libboost-dev libboost-system-dev \
                        libboost-filesystem-dev libexpat1-dev zlib1g-dev \
                        libbz2-dev libpq-dev libproj-dev \
                        postgresql-server-dev-12 postgresql-12-postgis-3 \
                        postgresql-contrib postgresql-12-postgis-3-scripts \
                        apache2 php php-pgsql libapache2-mod-php \
                        php-intl python3-setuptools python3-dev python3-pip \
                        python3-psycopg2 python3-tidylib
sudo useradd -d /srv/nominatim -s /bin/bash -m nominatim
export USERNAME=nominatim
export USERHOME=/work
chmod a+x $USERHOME

echo "Please make PostgreSQL database config changes..."
read

sudo systemctl restart postgresql

sudo -u postgres createuser -s $USERNAME
sudo -u postgres createuser www-data
sudo tee /etc/apache2/conf-available/nominatim.conf << EOFAPACHECONF
<Directory "$USERHOME/build/website">
  Options FollowSymLinks MultiViews
  AddType text/html   .php
  DirectoryIndex search.php
  Require all granted
</Directory>

Alias /nominatim $USERHOME/build/website
EOFAPACHECONF
sudo a2enconf nominatim
sudo systemctl restart apache2

sudo -iu nominatim

cd $USERHOME
wget https://nominatim.org/release/Nominatim-3.5.2.tar.bz2
tar xf Nominatim-3.5.2.tar.bz2
mkdir build
cd build
cmake $USERHOME/Nominatim-3.5.2
make
tee settings/local.php << EOF
<?php
 @define('CONST_Website_BaseURL', '/nominatim/');
EOF

# https://download.geofabrik.de/north-america/us-latest.osm.pbf