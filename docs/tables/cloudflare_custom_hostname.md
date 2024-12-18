# Table: cloudflare_custom_hostname

Custom Hostnames allow you to use your own domain name with SSL/TLS protection for your Cloudflare zone.

## Examples

### List all custom hostnames with their SSL configuration

```sql
select
  hostname,
  certificate_authority,
  ssl_method,
  ssl_status,
  ssl_type,
  ssl_wildcard
from
  cloudflare_custom_hostname;
```

### Find custom hostnames using Let's Encrypt certificates

```sql
select
  hostname,
  certificate_authority,
  ssl_status,
  created_at
from
  cloudflare_custom_hostname
where
  certificate_authority = 'lets_encrypt';
```

### List custom hostnames for a specific zone

```sql
select
  hostname,
  ssl_status,
  ssl_method,
  created_at
from
  cloudflare_custom_hostname
where
  zone_id = 'your-zone-id';
```

### Find custom hostnames with non-active SSL certificates

```sql
select
  hostname,
  ssl_status,
  certificate_authority,
  created_at
from
  cloudflare_custom_hostname
where
  ssl_status != 'active';
```

### List recently added custom hostnames

```sql
select
  hostname,
  certificate_authority,
  ssl_status,
  created_at
from
  cloudflare_custom_hostname
where
  created_at > (current_timestamp - interval '7 days')
order by
  created_at desc;
```

### Get SSL certificate distribution by certificate authority

```sql
select
  certificate_authority,
  count(*) as total_certificates,
  count(*) filter (where ssl_status = 'active') as active_certificates
from
  cloudflare_custom_hostname
group by
  certificate_authority
order by
  total_certificates desc;
