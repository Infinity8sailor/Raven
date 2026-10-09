#!/bin/bash
PLIST_FILE="$HOME/Library/LaunchAgents/com.raven.app.plist"

cat <<EOF > "$PLIST_FILE"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.raven.app</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/raven</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/usr/local/opt/raven</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
EOF

launchctl load "$PLIST_FILE"
echo "Raven LaunchAgent installed on macOS."
