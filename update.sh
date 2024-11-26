git pull
cp site site.backup
go build .
sudo systemctl restart site
