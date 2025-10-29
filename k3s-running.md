# Running `k3s` under this

The following is experimental, but shows a way of getting `k3s` running with this tool, with installation of Concourse.

First, let's install `htvend` as a local daemon - this will feed `k3s`.

```bash
# create user
useradd --system htvend

# create dir for blobs
mkdir -p /var/lib/htvend/etc /var/lib/htvend/store
test -f /var/lib/htvend/store/assets.json || echo "{}" > /var/lib/htvend/store/assets.json
chown -R htvend /var/lib/htvend

# anyone can read/list the config dir
chmod 755 /var/lib/htvend /var/lib/htvend/etc

# only our user can deal with store dir
chmod 700 /var/lib/htvend/store

# create service file
cat <<'EOF' > /etc/systemd/system/htvend.service
[Unit]
Description=htvend service
After=network.target
StartLimitIntervalSec=0
[Service]
Type=simple
Restart=always
RestartSec=1
User=htvend
ExecStart=/usr/bin/env \
    htvend offline \
        --manifest=/var/lib/htvend/store/assets.json \
        --blobs-dir=/var/lib/htvend/store/blobs \
        --listen-addr=127.0.0.1:4532 \
        --tls-generate-if-missing \
        --tls-cert-pem=/var/lib/htvend/etc/cert.pem \
        --tls-key-pem=/var/lib/htvend/etc/key.pem \
        --daemon \
        --allow-rpc-updates \
        --daemon-rpc-socket=/var/lib/htvend/rpc

[Install]
WantedBy=multi-user.target
EOF

# enable and start it
systemctl enable htvend
systemctl start htvend

# skip start, as we need to modify the env file that it creates
# set version, as it otherwise relies on some clever Location field behaviour
htvend build -m bootstrap.json -- bash -c "curl -sfL https://get.k3s.io | INSTALL_K3S_SKIP_START=true INSTALL_K3S_VERSION=v1.33.3+k3s1 sh -"
htvend import -m bootstrap.json --destination=/var/lib/htvend/rpc

# next, add our CA and proxy to the CONTAINERD config only:
cat <<EOF > /etc/systemd/system/k3s.service.env
CONTAINERD_HTTP_PROXY=http://127.0.0.1:4532
CONTAINERD_HTTPS_PROXY=http://127.0.0.1:4532
CONTAINERD_NO_PROXY=
CONTAINERD_SSL_CERT_FILE=/var/lib/htvend/etc/cert.pem
EOF

systemctl enable k3s
systemctl start k3s
```

```bash
# poke in an image - this does not need be run as root
htvend build -m rancher-images.json -- download-image \
    rancher/mirrored-pause:3.6 \
    rancher/local-path-provisioner:v0.0.31 \
    rancher/mirrored-metrics-server:v0.7.2 \
    rancher/mirrored-coredns-coredns:1.12.1 \
    rancher/klipper-helm:v0.9.8-build20250709 \
    rancher/mirrored-library-traefik:3.3.6 \
    rancher/klipper-lb:v0.4.13 \
    rancher/mirrored-library-busybox:1.36.1 \
    concourse/concourse:7.14.0 \
    bitnami/postgresql:17.5.0-debian-12-r12

# ingest it
sudo htvend import -m rancher-images.json --destination=/var/lib/htvend/rpc
```


Then, to install, in that terminal:

```bash
# start up k3s service
systemctl start k3s

# install helm binary
htvend build -m helm.json -- bash -c "curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"
htvend import -m helm.json --destination=/var/lib/htvend/rpc

# add Concourse helm repo
htvend build -m concourse-helm-repo.json -- helm repo add concourse https://concourse-charts.storage.googleapis.com/
htvend import -m concourse-helm-repo.json --destination=/var/lib/htvend/rpc

# install Concourse (remove 7.14.0 once chart is updated)
KUBECONFIG=/etc/rancher/k3s/k3s.yaml \
    helm install \
        --set imageTag=7.14.0 \
        my-release \
        concourse/concourse

# set up port-forward to see Concourse locally (test/test):
# as root:
k3s kubectl port-forward \
    --namespace default \
    $(
        k3s kubectl get pods \
            --namespace default \
            -l "app=my-release-web" \
            -o jsonpath="{.items[0].metadata.name}" \
    ) \
    8080:8080

# and to stop and uninstall to try again:
systemctl stop k3s
/usr/local/bin/k3s-uninstall.sh
rm -rf /root/.config/helm /root/.cache/helm /usr/local/bin/helm
```
