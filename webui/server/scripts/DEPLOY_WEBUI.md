# deploy_webui: Cloudflare Tunnel 前提のデプロイ手順

このドキュメントは `scripts/deploy_webui.ps1` を使って webui を Debian 12 サーバへデプロイする手順をまとめたものです。今回の運用は **Cloudflare Tunnel（cloudflared）を使って外部からポート番号無しでアクセスさせる** 方式を前提とします。

## 前提条件

- Windows 側に OpenSSH クライアント（ssh, scp）があること。
- デプロイ先サーバ（例: 192.168.1.233）は Debian 12、nginx を使うことを想定。
- Cloudflare アカウントでドメイン（例: `gcsim-uoc.linole.net`）を管理できること。
- サーバは外向きに任意のポート（例: 7056）のみしか開けない環境でも OK（cloudflared がアウトバウンドで Cloudflare と接続できれば問題ありません）。

## 概略フロー

1. サーバに `cloudflared` を導入し、Cloudflare にログインして `cert.pem`（origin 証明書）を取得/配置する。
2. `cloudflared tunnel create` でトンネルを作成し、`cloudflared tunnel route dns` でドメインをトンネルに割当てる。
3. `/etc/cloudflared/config.yml` を作成して `systemd` で常駐させる。
4. nginx はローカルで HTTP (80) または HTTPS (443, origin cert) を待ち受け、cloudflared は Cloudflare 側の TLS を終端して nginx へ転送する。
5. ローカルの `deploy_webui.ps1` で webui をアップロード・設置する。

## 重要な設計決定

- Cloudflare Tunnel を使う利点: ルータで 80/443 を公開しなくても Cloudflare 経由で HTTPS を提供できる。
- nginx と cloudflared の間を HTTPS にする場合は「Origin CA」または自己署名の origin 証明書を nginx に配置して TLS を有効にするとより安全です。deploy スクリプトは origin cert が無ければ自己署名証明書を生成して nginx に配置するように更新済みです。

## cloudflared のセットアップ（要点）

1. cloudflared のインストール（Debian 12）

```bash
curl -L -o cloudflared.tgz "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.tgz"
tar -xzf cloudflared.tgz
sudo mv cloudflared /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared
```

2. cloudflared でログイン（サーバ上でブラウザが開ける場合）

```bash
cloudflared tunnel login
```

成功すると `~/.cloudflared/cert.pem` が生成されます（ヘッドレスの場合はローカルで login して `cert.pem` をサーバへ転送してください）。

3. トンネル作成と DNS 割当

```bash
cloudflared tunnel create gcsim-uoc
cloudflared tunnel route dns gcsim-uoc gcsim-uoc.linole.net
```

4. `/etc/cloudflared/config.yml` を作成して systemd で常駐

例（nginx をローカル 80 で待ち受ける場合）:

```yaml
tunnel: <TUNNEL_ID>
credentials-file: /root/.cloudflared/<TUNNEL_ID>.json

ingress:
    - hostname: gcsim-uoc.linole.net
        service: http://localhost:80
    - service: http_status:404
```

その後 `sudo cloudflared service install` または systemd ユニットを作成して `systemctl enable --now cloudflared` で常駐させます。

## nginx と Origin 証明書

- セキュリティ向上のため、cloudflared→nginx 間を HTTPS にしたい場合は Cloudflare Origin CA を発行して nginx に入れるか、自己署名の origin 証明書を用いることができます。
- `deploy_webui.ps1` は、リモートで Origin 証明書が無ければ自動的に自己署名 cert を生成して `/etc/ssl/private/gcsim-origin.key` と `/etc/ssl/certs/gcsim-origin.pem` に置くように更新されています。

## systemd ユニットの改善点

- `Restart=on-failure` と `RestartSec=5s` を設定しておくのが無難です（再接続を自動化）。
- ログは systemd-journal に流れるため `journalctl -u cloudflared -f` で監視します。
- 必要なら監視ツール（Prometheus node exporter, systemd watchdog, monit など）でプロセスを監視し、異常時に通知する構成を追加してください。

## deploy_webui の使い方（Cloudflare Tunnel 前提）

ローカル（Windows、PowerShell）からの例：

```powershell
.\scripts\deploy_webui.ps1 `
    -Server 192.168.1.233 `
    -User uocuser `
    -KeyFile C:\Users\linol\.ssh\id_ed25519 `
    -UseNginx `
    -ConfigureNginx `
    -Domain gcsim-uoc.linole.net `
    -ReloadNginx
```

- まずは `-KeepRemoteTmp` を付けて動作確認することを推奨します。リモートにアップロードされたファイル（deploy_remote.sh, install_nginx.sh 等）を目視チェックできます。
- デフォルトではスクリプトは nginx のルートにファイルを置き、`www-data:www-data` に chown します。

## 自動更新（certbot）について

- Cloudflare Tunnel を使う場合は Cloudflare 側で TLS 終端するためサーバ側の Let’s Encrypt は必須ではありません。ただし nginx 側で HTTPS を有効にして origin cert を使う場合は、origin cert の更新手順（Cloudflare の Origin CA の再発行）を運用ルールとしておく必要があります。

## トラブルシューティング（FAQ）

- **curl で 200 が返るがブラウザで繋がらない**: Cloudflare 側の DNS キャッシュまたはブラウザキャッシュを確認。`dig` で CNAME が `*.cfargotunnel.com` を指しているか確認。
- **cloudflared サービスが落ちる**: `journalctl -u cloudflared -n 200` でログを確認。credentials-file の権限やパスが正しいかチェック。

## 付録: origin cert を Cloudflare から取得する場合

Cloudflare ダッシュボードで Origin CA を発行し、`/etc/ssl/certs/gcsim-origin.pem` と `/etc/ssl/private/gcsim-origin.key` に置けば、nginx 側の設定はそのまま使えます（ただし Origin CA は Cloudflare のみに信頼される証明書なのでローカルで直接ブラウザ検証すると警告が出ます）。

---

このドキュメントは `deploy_webui.ps1` の更新に合わせて随時更新してください。
This repository contains a simple PowerShell script `deploy_webui.ps1` that uploads the `webui/` folder to a remote Debian server over SSH/SCP. The script accepts a `-Port` parameter which specifies the SSH port to connect to on the remote host.

Assumptions
-----------
- The `-Port` parameter is the SSH port. Note: if port 7056 is currently used only for external web access (HTTP) and is NOT forwarded to the server's SSH, do NOT use 7056 as the `-Port` value. Instead either:
    - use the server's real SSH port (often 22) if it is reachable directly from your dev machine, or
    - change your router's port forwarding so an external port maps to the server's SSH port and then use that external port here.
- OpenSSH client (ssh, scp) is available on your Windows dev machine.
- The remote user can use `sudo` to move files into the target directory (the script uses sudo).
- Caddy is installed on the server and serves files from the target directory (for example `/var/www/gcsim-uoc`).

Usage
-----
From the repository root on your Windows machine (PowerShell):

```powershell
.\scripts\deploy_webui.ps1 -Server "192.168.1.233" -User "deploy" -RemotePath "/var/www/gcsim-uoc"
```

Optional flags:
- `-KeyFile "C:\path\to\private_key"` : use this key for SSH auth.
- `-BuildCmd "npm run build --prefix ui"` : run a local build command before uploading.
- `-KeepRemoteTmp` : keep the remote temporary upload directory for debugging.
- `-ReloadCaddy` : reload Caddy after deploy (the script will run `sudo systemctl reload caddy` when this flag is given).
- `-EnablePasswordlessSudo` : (dangerous) attempt to install a temporary sudoers entry for the deploy user so that the deploy commands (mv, rm, chown, mkdir, systemctl) can run without prompting for a sudo password. The script uploads a file to `/etc/sudoers.d/gcsim-deploy` on the remote host. Use only if you trust the environment. The file will be backed up if an existing `/etc/sudoers.d/gcsim-deploy` exists.
- `-SudoPasswordFile "C:\path\to\encpwd.txt"` : (safer than passing plain password) path to a DPAPI-encrypted password file created by `scripts/encode_sudo_password.ps1`. When provided, the deploy script will decrypt the file locally (only the same Windows user can decrypt) and feed the password into remote `sudo -S` calls so the deploy can run non-interactively even if sudoers is not installed.

Server-side notes (Debian 12)
-----------------------------
1. Ensure Caddy's site config points to the `RemotePath` as the site root. Example Caddyfile snippet:

```
gcsim-uoc.linole.net:7056 {
    root * /var/www/gcsim-uoc
    file_server
}
```

Note: The Caddyfile example above shows serving on port 7056. This is separate from SSH. If 7056 is only reachable as a web port from outside, the deploy script must still reach the server's SSH port. Either forward an external port to SSH or run the script from inside the same LAN where SSH (usually port 22) is reachable.

2. If Caddy runs as a system service and expects files owned by `www-data`, the script sets ownership accordingly. Adjust `chown` in the script if you use a different user.

Security
--------
- Prefer key-based SSH authentication. If you must use passwords, the script will prompt during SSH/SCP operations.
- The script uses `sudo` on the remote host; configure `sudo` and the remote user's privileges appropriately.

Troubleshooting
---------------
- If SCP fails, confirm connectivity: `ssh deploy@192.168.1.233` (the script does not pass an explicit port; ssh will use the default port or your `~/.ssh/config`).

Creating an encrypted sudo password file (DPAPI)
---------------------------------------------
1. Run the helper to create an encrypted password file (only readable by your Windows user):

```powershell
.\scripts\encode_sudo_password.ps1 -OutFile C:\Users\linol\.ssh\enc_sudo_pwd.txt
```

2. Provide that file to the deploy script:

```powershell
.\scripts\deploy_webui.ps1 -Server "192.168.1.233" -User "uocuser" -RemotePath "/var/www/gcsim-uoc" -KeyFile "C:\Users\linol\.ssh\id_ed25519" -SudoPasswordFile "C:\Users\linol\.ssh\enc_sudo_pwd.txt"
```

Security notes:
- The encrypted file is protected by Windows DPAPI and can only be decrypted by the same Windows user account. This is safer than embedding a plain password in the script, but store the encrypted file securely and remove it when not needed.
- After a successful deploy you may want to remove the file:

```powershell
Remove-Item C:\Users\linol\.ssh\enc_sudo_pwd.txt
```
- If the web files aren't served, check Caddy logs and ensure the site root matches the deployed path.

Next steps
----------
- (Optional) Add an automated step to reload Caddy after deploy.
- (Optional) Run the built web UI locally before deploy to verify correctness.
