# nats-centralized-auth-callout-examples

## static

The current nats-server code base only supports centralized with static accounts.  i.e. you can't create accounts on demand without a nats-server code change.  

### Bring up the nats server

```shell
docker-compose -f .\docker-compose-static.yml up
```

### Run the auth callout service

```shell
go build .\cmd\cli\.
.\cli.exe callout services static
```

### Simple request reply

```shell
.\cli.exe handlers request
```

```shell
.\cli.exe clients request_reply
```
