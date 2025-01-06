git pull origin main
cp site site.backup
go build .
sudo systemctl restart site
