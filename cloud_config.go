package main

import (
	"fmt"
	"strings"
)

func generateCloudConfig(cfg *Config) string {
	cloudCfg := fmt.Sprintf(`#cloud-config
chpasswd:
  list: |
    root:%s
  expire: false
manage-resolv-conf: true
resolv_conf:
  nameservers:
    - '%s'
    - '%s'
package_update: true
package_upgrade: true
packages:
  - apt-transport-https
  - gnupg
  - tor
runcmd:
  - ufw allow from %s to any port 22
  - ufw allow from %s to any port %d
  - ufw enable
  - echo "TOR-DROPLET CONFIGURED"
write_files:
  - path: /etc/tor/torrc
    content: |
      SocksPort 0.0.0.0:%d
      %s
`, remotePassword,
		cfg.DNS.Primary,
		cfg.DNS.Secondary,
		cfg.IP,
		cfg.IP, cfg.Ports.Remote,
		cfg.Ports.Remote,
		strings.Join(cfg.TorOptions, "\n      "))
	return cloudCfg
}
