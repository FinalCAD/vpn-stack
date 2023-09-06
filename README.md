# vpn-stack

OpenVpn stack service to synchronise IAM and OpenVpn Users.

## aws-openvpn-updater

Golang binary running on openvpn server as a service. Check at interval if all users in a specific IAM group have a valid openvpn connexion. If they are not allowed on the vpn, users are added to openpvn, a client configuration can be uploaded to a s3 bucket, and a presign-url can be send by mail with SES. Users having a valid openvpn configuration but not present in the IAM group will see their access revoked, client configuration file and S3 uploaded file are deleted.

### Command parameters

- `debug` : Enable debug logging, default : false
- `config` : Path to toml configuration file, default: ./config.toml
- `env` : Environment, required

### Configuration file

| key  	| Details  	| Default\Required  	|
|---	|---	    |---	    |
| **settings** |
| dry-run           | Enable Dry-run (disable openvpn file change)  | false                           |
| request-interval  | Internal loop for synchronization in seconds  | 300                             |
| s3-upload         | Activate S3 upload of openvpn configuration   | true                            |
| sender            | Default mail from                             | required when send-mail is true |
| send-mail         | Activate SES to send presign-url              | true                            |
| **openvpn** |
| easy-rsa-path     | path from root to easy-rsa directory          | /etc/openvpn/server/easy-rsa    |
| key-directory     | name of directory in easy-rsa that holds keys | pki                             |
| server-path       | path from root to openvpn server              | /etc/openvpn/server             |
| **aws** |
| profile           | aws profile to assume                         | none                            |
| region            | aws region                                    | eu-central-1                    |
| s3-bucket-name    | aws s3 bucket name                            | required                        |
| vpn-group         | aws IAM group name                            | required                        |
| assume-role       | aws assume role                               | none                            |

#### Exemple

```toml
[settings]
request-interval = 10
s3-upload = false
send-mail = false
dry-run = true

[openvpn]
easy-rsa-path = "./easy-rsa"
server-path = "./easy-rsa"

[aws]
vpn-group = "tf-vpn-sandbox"
profile = "master"
region = "eu-central-1"
s3-bucket-name = "tf-vpn-config"
```
