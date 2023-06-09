# apithings

`apithings` is a collection of tooling I always find myself rebuilding when a I need a simple "thing" i.e. I just need a basic status page or webhook relay or link shortener. This repository is the experimentation grounds.
Some things may move into their own repo.

## Components

Right now the only component I've finished is a basic status site "thing". As other components get written, they'll start to (hopefully) coalesce in a set of blocks that play well together (i.e. the status thing can trigger the webhook thing)

### StatusThing
*Note: StatusThing now has an in-development repository: https://github.com/lusis/statusthing 
This repository's `statusthing` will likely redirect you to the above repo in the future*

`statusthing` is a very simple status page tool. It doesn't support comments or historical data (yet? maybe soon?).

The docs for `statusthing` are [here](https://github.com/lusis/apithings/blob/main/README.md)

#### Primary usecase
In the past I've needed a simple internal status page. It didn't need to be fancy.
The api should be simple enough that you can fire off a webhook (or even add it to a service startup script) to update the status of something quickly and easily.
You could make it a sidecar/component that you could deploy with other apps.

The "dashboard" I've shipped here is tragic for now but seeing as this is generally something I needed internally, you could point your own UI at the api if you wanted.


#### StatusThing Quick Start

- Build and run the container
```
docker build --rm -f Dockerfile.statusthing -t statusthing .
docker run -p 9000:9000 --rm -i -t statusthing
```

- Add a service
```
curl -XPUT -H "Content-Type: application/json" -d '{"status":"STATUS_YELLOW", "name":"tryhard","description":"ehhhhhhh"}' http://localhost:9000/statusthings/api/
curl -XPUT -H "Content-Type: application/json" -d '{"status":"STATUS_GREEN", "name":"loool","description":"✅"}' http://localhost:9000/statusthings/api/
curl -XPUT -H "Content-Type: application/json" -d '{"status":"STATUS_RED", "name":"uggg","description":"💩"}' http://localhost:9000/statusthings/api/
```

- Open your browser to http://localhost:9000/statusthings

## Common behaviour
There is some configuration/behavior that will be common across all the 'things'

### ngrok support
You can expose your application via ngrok automatically as well by setting the `NGROK_AUTHTOKEN` environment variable.
This will expose the service over a random ngrok tunnel. Everything else behaves the same (i.e. setting an api key via `STATUSTHING_APIKEY` will still require an api key for requests).

If you would like to change the ngrok hostname/endpoint name, set the env var `NGROK_ENDPOINT` in addition to `NGROK_AUTHTOKEN`

### tailscale support
*coming soon*

### otel support 
*coming soon*

## FAQ
I've started an [FAQ](https://github.com/lusis/apithings/blob/main/FAQ.md) as well.

## Contributing
Unfortunately, I'm not quite ready for contributions/PRs yet.

## Hiring
I'm on the job market right now. If you like tooling like this and want someone to come work for your team either writing customer facing Go code or building out internal tooling for your engineering org, you should reach out:

- [LinkedIn](https://www.linkedin.com/in/lusis/)
- [Blog](https://blog.lusis.org)
- Email: `lusis.org at gmail.com`
