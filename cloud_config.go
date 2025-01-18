package main

import (
	"fmt"
	"strings"
)

func generateCloudConfig(cfg *Config) string {
	return fmt.Sprintf(`#cloud-config
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
  - apt-get install -y tor
write_files:
  - path: /etc/tor/torrc
    content: |
      SocksPort 0.0.0.0:%d
      %s
`, remotePassword, cfg.DNS.Primary, cfg.DNS.Secondary, cfg.Ports.Remote, strings.Join(cfg.TorOptions, "\n      "))
}
