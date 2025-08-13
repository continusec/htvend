# Running `k3s` under this

The following is experimental, but shows a way of getting `k3s` running with this tool.

```bash
# in one terminal
htvend build -d -v -s
```

In a separate terminal, as root:

```bash
# first, copy/paste all the env vars printed from above here, then:
# running as root - note that INSTALL_K3S_VERSION must be set as otherwise relies on reading Location header
curl -sfL https://get.k3s.io | INSTALL_K3S_SKIP_START=true INSTALL_K3S_VERSION=v1.33.3+k3s1 sh -

# append the cert file locations
echo SSL_CERT_FILE=$SSL_CERT_FILE >> /etc/systemd/system/k3s.service.env

# and start it
systemctl start k3s

# note that k3s.service.env will need to be updated each time - we can make that work better in a subsequent update.

# and to stop and uninstall to try again:
systemctl stop k3s
/usr/local/bin/k3s-uninstall.sh
```
