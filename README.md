# Blockchain from zero
This is Bitcoin based simple and scratch blockchain implementation
Mainly used for personal study

## Run
```
# run bellows by this order in different terminal

DEBUG_LOCAL_IP=1 go run src/core1.go
DEBUG_LOCAL_IP=1 go run src/core2.go
DEBUG_LOCAL_IP=1 DEBUG_ADD_COIN=1 go run src/wallet_app.go
DEBUG_LOCAL_IP=1 go run src/wallet_app2.go
```


### Wallet Feature
- Send coin to other wallet
image TBD

- Send direct message to other wallet
image TBD


#### Options

- DEBUG_LOCAL_IP
This forces to connect to 127.0.0.1
If you don't set this, need to edit codes

- DEBUG_ADD_COIN
This add 90 coins as default. mainly used by wallet
90 = 3 coinbase transaction * 30 coin

