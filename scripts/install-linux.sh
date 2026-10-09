#!/bin/bash
SERVICE_FILE="$HOME/.config/systemd/user/raven.service"
mkdir -p "$HOME/.config/systemd/user"

cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=Raven Minimal App Backend
After=network.target

[Service]
ExecStart=/opt/raven/raven
WorkingDirectory=/opt/raven
Restart=always

[Install]
WantedBy=default.target
EOF

systemctl --user daemon-reload
systemctl --user enable raven.service
systemctl --user start raven.service
echo "Raven service installed and started for Linux user."
