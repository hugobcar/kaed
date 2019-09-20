```sh
export AWS_ACCESS_KEY_ID='YOUR KEY'  
export AWS_SECRET_ACCESS_KEY='YOUR SECRET'  
export ANUS_TTL='60'  
export ANUS_DOMAIN='anus.dominio-devel.com.br'  
export ANUS_ZONEID='ROUTE53 ZONE ID'
```

Especific policy ANUS zoneID:
```yaml
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "route53:*"
            ],
            "Resource": [
                "arn:aws:route53:::hostedzone/<ZONE_ID>"
            ]
        }
    ]
}
```
