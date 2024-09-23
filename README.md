# ratelimiter


## Install

`go get github.com/OwodDEV/ratelimiter`

## Example

Local usage:
```
rlManager := ratelimiter.New()

rl, _ := rlManager.GetOrCreate("ip")
_ = rl.SetRules("limit=2;reset=10s")
_ = rl.Inc("127.0.0.1")
isAllow, _ := rl.Allow("127.0.0.1")
```

Remote usage:

```
rlMasterManager := ratelimiter.New()
rlMasterManager.Serve("8080")
```


```
rlSlaveManager := ratelimiter.NewRemote("127.0.0.1:8080")
```

## Rules

`limit` - for limit of count usage

`reset` - for rule of reset (duration or calendar period);

Examples:

```
limit=20;reset=1h10m2s
limit=20;reset=25s
limit=10;reset=calendar@hour
limit=10;reset=calendar@day
```
