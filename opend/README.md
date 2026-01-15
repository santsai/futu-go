# Running Futu OpenD Gateway in Apple Container

## References
* Apple Container: https://github.com/apple/container
* Futu OpenD: https://openapi.futunn.com/futu-api-doc/en/opend/opend-cmd.html

## Setup
Generating RSA private key:
```bash
make -C .. genkey
```
Editing OpenD config file:
* filename defaults to ```opend-dev.xml``` and place in ```data``` folder.
* edit ip ```<ip>0.0.0.0</ip>```
* edit private key ```<rsa_private_key>/opend/key.pem</rsa_private_key>```
* edit ```<login_account>``` and ```<login_pwd_md5>``` accordingly

Container building:
```bash
make -C .. build_opend
```

Container running:
```bash
make -C .. start_opend
```
