# nats-centralized-auth-callout-examples

## static

The current nats-server code base only supports centralized with static accounts. i.e. you can't create accounts on demand without a nats-server code change.

### Bring up the nats server

```shell
docker-compose -f .\docker-compose-static.yml up
```

### Run the auth callout service

```shell
go build .\cmd\cli\.
.\cli.exe callout services static --nats.user auth --nats.pass auth
```

### Request Reply

#### Request Handler

```shell
.\cli.exe handlers request --nats.user greeter --nats.pass greeter
```

#### Request Client

```shell
.\cli.exe clients request_reply --nats.user joe --nats.pass joe
```

or

```shell
.\cli.exe clients request_reply --nats.user alice --nats.pass alice
```

You can see from the [users](configs/users.json) who has the right to publish and handle the greet requests.

## Jetstream

```shell

.\cli.exe jetstream create          --nats.user god --nats.pass god --js.name  webhooks_inbound --js.subject webhooks.inbound --js.subject webhooks.inbound.>
.\cli.exe jetstream info            --nats.user god --nats.pass god --js.name  webhooks_inbound

.\cli.exe jetstream consumer add    --nats.user god --nats.pass god --js.name  webhooks_inbound --consumer.name wa1
.\cli.exe jetstream consumer info   --nats.user god --nats.pass god --js.name  webhooks_inbound --consumer.name wa1

.\cli.exe jetstream publish         --nats.user god --nats.pass god --subject webhooks.inbound.github --duration 10s --pause.duration 10ms
```
