# nats-centralized-auth-callout-examples

## static

The current nats-server code base only supports centralized with static accounts. i.e. you can't create accounts on demand without a nats-server code change.

This is the [server conf](./configs/static_callout.conf)

What you see is what you get here for accounts. If you want to add new account you have to add them to the conf and restart the nats server.

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

### Micro Request/Reply

#### Micro Request Handler

```shell
.\cli.exe handlers micro --nats.user greeter --nats.pass greeter
```

#### Micro Request Client

```shell
.\cli.exe clients micro --nats.user joe --nats.pass joe
```

or

```shell
.\cli.exe clients micro --nats.user alice --nats.pass alice
```

### jetstream

```shell

.\cli.exe jetstream create          --nats.user god --nats.pass god --js.name  write_scatter_gather --js.subject *.write.props --js.subject *.write.props.> --js.subject *.write.resp.>

.\cli.exe jetstream info            --nats.user god --nats.pass god --js.name  write_scatter_gather

.\cli.exe jetstream consumer add    --nats.user god --nats.pass god --js.name  write_scatter_gather --consumer.name middleware_write_props --consumer.filterSubjects *.write.props --consumer.filterSubjects *.write.props.>

.\cli.exe jetstream publish_one         --nats.user god --nats.pass god --subject org1234.write.props.github

.\cli.exe jetstream publish         --nats.user god --nats.pass god --subject org1234.write.props.github --duration 0s --pause.duration 10ms
```

## static/and_dynamic

This is for a customized build of nats-server that does account lookups the same why that the decentraized auth variant.

This is the [server conf](./configs/static_callout_lookup.conf)

### Bring up the nats server

```shell
docker-compose -f .\docker-compose-static-account-lookup.yml up
```

### Run the auth callout service

```shell
go build .\cmd\cli\.
.\cli.exe callout services static and_dynamic --nats.user auth --nats.pass auth
```

#### Request Handler

```shell
.\cli.exe handlers request --nats.user greeter@SVC --nats.pass greeter
```

#### Request Client

```shell
.\cli.exe clients request_reply --nats.user joe@SVC --nats.pass joe
```

or

```shell
.\cli.exe clients request_reply --nats.user alice@SVC --nats.pass alice
```

#### Jetstream Scatter-Gather

a request is sent to a jetstream consumer that will result in 1..n scatter requests which will take a while to respond and will publish their results to whatever reply subject that was passed in the message header.

```shell

.\cli.exe jetstream create          --nats.user god@sg --nats.pass god --js.name  write_scatter_gather --js.subject *.write.props --js.subject *.write.props.> --js.subject *.write.resp.>

.\cli.exe jetstream info            --nats.user god@sg --nats.pass god --js.name  write_scatter_gather

.\cli.exe jetstream consumer add    --nats.user god@sg --nats.pass god --js.name  write_scatter_gather --consumer.name middleware_write_props --consumer.filterSubjects *.write.props --consumer.filterSubjects *.write.props.>

.\cli.exe jetstream publish_one         --nats.user god@sg --nats.pass god --subject org1234.write.props.github

.\cli.exe jetstream publish         --nats.user god@sg --nats.pass god --subject org1234.write.props.github --duration 0s --pause.duration 10ms




.\cli.exe jetstream consumer info   --nats.user god@sg --nats.pass god --js.name  write_scatter_gather --consumer.name middleware_write_props



.\cli.exe jetstream consumer add    --nats.user god@sg --nats.pass god --js.name  write_scatter_gather --consumer.name middleware_write_resp --consumer.filterSubjects *.write.resp.>

.\cli.exe jetstream consumer info   --nats.user god@sg --nats.pass god --js.name  write_scatter_gather --consumer.name middleware_write_resp

.\cli.exe jetstream publish         --nats.user god@sg --nats.pass god --subject org1234.write.props.github --duration 10s --pause.duration 10ms

.\cli.exe jetstream consumer add    --nats.user god@sg --nats.pass god --js.name  write_scatter_gather --consumer.name write_props_ct1 --consumer.filterSubjects write.props.ct1
.\cli.exe jetstream consumer add    --nats.user god@sg --nats.pass god --js.name  write_scatter_gather --consumer.name write_props_ct2 --consumer.filterSubjects write.props.c2



.\cli.exe jetstream consumer info   --nats.user god --nats.pass god --js.name  write_scatter_gather --consumer.name middleware_write_resp --consumer.filterSubjects write.resp.>


.\cli.exe jetstream publish         --nats.user god --nats.pass god --subject webhooks.inbound.github --duration 10s --pause.duration 10ms

.\cli.exe jetstream consume         --nats.user god --nats.pass god --js.name  webhooks_inbound --consumer.name wa1

```

## Jetstream

```shell

.\cli.exe jetstream create          --nats.user god --nats.pass god --js.name  webhooks_inbound --js.subject webhooks.inbound --js.subject webhooks.inbound.>
.\cli.exe jetstream info            --nats.user god --nats.pass god --js.name  webhooks_inbound

.\cli.exe jetstream consumer add    --nats.user god --nats.pass god --js.name  webhooks_inbound --consumer.name wa1
.\cli.exe jetstream consumer info   --nats.user god --nats.pass god --js.name  webhooks_inbound --consumer.name wa1

.\cli.exe jetstream publish         --nats.user god --nats.pass god --subject webhooks.inbound.github --duration 10s --pause.duration 10ms

.\cli.exe jetstream consume         --nats.user god --nats.pass god --js.name  webhooks_inbound --consumer.name wa1

```
