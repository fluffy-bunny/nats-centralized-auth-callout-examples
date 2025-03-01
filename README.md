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
.\cli.exe clients request_reply --nats.user joe --nats.pass joe
```
or

```shell
.\cli.exe clients request_reply --nats.user alice --nats.pass alice
```

You can see from the [users](configs/users.json) who has the right to publish and handle the greet requests.