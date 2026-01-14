#!/bin/bash
set -euo pipefail

echo "=== Setting up authentication on server ==="

# Generate JWT secret
JWT_SECRET=$(openssl rand -base64 48)
echo "Generated JWT secret: ${JWT_SECRET:0:20}..."

# Create environment file
sudo mkdir -p /etc/gcsim-webui
sudo tee /etc/gcsim-webui/.env > /dev/null <<EOF
# JWT Authentication Secret
GCSIM_JWT_SECRET=$JWT_SECRET

# Admin Credentials (change password after first login)
GCSIM_ADMIN_USERNAME=admin
GCSIM_ADMIN_PASSWORD=TempPass123!

# CORS Configuration
GCSIM_CORS_ALLOWED_ORIGINS=https://gcsim-uoc.linole.net

# Trusted Proxy Configuration (Cloudflare Tunnel)
GCSIM_TRUSTED_PROXIES=127.0.0.1
EOF

# Set permissions
sudo chown www-data:www-data /etc/gcsim-webui/.env
sudo chmod 600 /etc/gcsim-webui/.env
echo "Created /etc/gcsim-webui/.env with secure permissions"

# Create data directory
sudo mkdir -p /var/www/html/data
sudo chown www-data:www-data /var/www/html/data
sudo chmod 750 /var/www/html/data
echo "Created data directory"

# Update systemd unit file
sudo tee /etc/systemd/system/gcsim-webui.service > /dev/null <<'UNIT'
[Unit]
Description=gcsim-webui HTTP API with JWT Authentication
After=network.target

[Service]
User=www-data
WorkingDirectory=/var/www/html
EnvironmentFile=-/etc/gcsim-webui/.env
ExecStart=/usr/local/bin/gcsim-webui -addr=:8382
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
UNIT

echo "Updated systemd unit file"

# Reload and restart service
sudo systemctl daemon-reload
sudo systemctl restart gcsim-webui
echo "Service restarted"

# Wait a moment for service to start
sleep 2

# Check service status
echo ""
echo "=== Service Status ==="
sudo systemctl status gcsim-webui --no-pager -l | head -30

echo ""
echo "=== Recent Logs ==="
sudo journalctl -u gcsim-webui -n 20 --no-pager

echo ""
echo "=== Setup Complete ==="
echo "Admin username: admin"
echo "Admin password: TempPass123!"
echo "Please change the password after first login!"
