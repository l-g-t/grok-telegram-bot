# Grok Telegram Bot


### Why

Grok 4 in UI is available for a 30EUR/month basis. However if you setup xAI developer account, add billing then you can use Grok 4 via API.
In my personal usage API calls have never yet reached the free limit.. So Grok 4 is essentially free.

### Usage

- normal messages use grok 4
- messages prepended with `fast.` use grok 3

### Deployment

This is intentionally a 'dumb' systemd unit to just whach on some spare VM :)

1. Copy `deploy/init/artifacts/.env.example` as `deploy/init/artifacts/.env` and update values
2. `make -C deploy configure-vm`
2. `make -C deploy release`


### TODO

Parse mode formatting is still wonky, hard to force Grok to properly format for telegram.
