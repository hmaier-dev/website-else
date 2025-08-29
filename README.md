## How to develop locally
If you want to develop the website further, do it locally. The following wiki entry can be helpful: [Hugo](https://hmaier-dev.github.io/wiki/hugo/#develop-locally-with-hugo).

## Deploy the image
Use `earthly +deploy` to update the container running on the VM. Locally you will need a `.secret`-file containing the needed credentials. The secret file will look like this:
```bash
host=1223.445.667.88
port=22
username=myuser
dest="/path/to/dest"
key='-----BEGIN OPENSSH PRIVATE KEY-----
-----END OPENSSH PRIVATE KEY-----'
known_hosts=''
dir=''
```
On Github I strongly recommend you to store the credentials as repository secrets (Option can be found at: `settings/secrets/actions`).
