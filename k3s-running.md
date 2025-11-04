# Running `k3s` under this

The following is experimental, but shows a way of getting `k3s` running with this tool, with installation of Concourse.

## Install `htvend` as a deamon

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
```

## Install `k3s`

Now let's install `k3s`. This should be run as root.

```bash
# skip start, as we need to modify the env file that it creates
# set version, as it otherwise relies on some clever Location field behaviour
htvend build -m k3s-install.json -- bash -c "curl -sfL https://get.k3s.io | INSTALL_K3S_SKIP_START=true INSTALL_K3S_VERSION=v1.34.1+k3s1 sh -"

# download all needed k3s images
htvend build -m k3s-images.json -- download-image \
    rancher/mirrored-pause:3.6 \
    rancher/local-path-provisioner:v0.0.32 \
    rancher/mirrored-metrics-server:v0.8.0 \
    rancher/mirrored-coredns-coredns:1.12.3 \
    rancher/klipper-helm:v0.9.8-build20250709 \
    rancher/mirrored-library-traefik:3.3.6 \
    rancher/klipper-lb:v0.4.13 \
    rancher/mirrored-library-busybox:1.36.1

# add all needed images to the htvend daemon
htvend import -m k3s-images.json --destination=/var/lib/htvend/rpc

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

## Install `helm`

Many deployments use `helm` to bootstrap.

```bash
# install helm binary
htvend build -m helm.json -- bash -c "curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"

# slurp into daemon
htvend import -m helm.json --destination=/var/lib/htvend/rpc
```

## Install Concourse

Concourse is super useful for visualising workflow. Let's install that using Helm.

```bash
# get all images needed for Concourse
htvend build -m concourse-images.json -- download-image \
    concourse/concourse:7.14.2 \
    library/postgres:17

# add all needed images to the htvend daemon so that k3s can access
htvend import -m concourse-images.json --destination=/var/lib/htvend/rpc

# add Concourse helm repo
htvend build -m concourse-helm-repo.json -- helm repo add concourse https://concourse-charts.storage.googleapis.com/
htvend import -m concourse-helm-repo.json --destination=/var/lib/htvend/rpc

# install Concourse
KUBECONFIG=/etc/rancher/k3s/k3s.yaml \
    helm install \
        my-release \
        concourse/concourse
```

## Test it 

```bash
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
```

## Uninstall

```bash
# and to stop and uninstall to try again:
systemctl stop k3s
/usr/local/bin/k3s-uninstall.sh
rm -rf /root/.config/helm /root/.cache/helm /usr/local/bin/helm
```
