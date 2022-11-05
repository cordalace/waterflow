# waterflow

Avoid obstacles.

Waterflow is a collection of tools for resolving blocked domain names. Uses
[Rublacklist registry of blocked websites](https://reestr.rublacklist.net/en/article/api/)

## List blocked domain names

```bash
wf-domains
```

## Resolve blocked domain names

```bash
wf-resolve \
    --workers 300 \
    --timeout 5 \
    --upstream https://1.1.1.1/dns-query \
    --upstream https://1.0.0.1/dns-query \
    --upstream https://dns-unfiltered.adguard.com/dns-query \
    --upstream https://dns.google/dns-query \
    --upstream https://doh.opendns.com/dns-query
```
