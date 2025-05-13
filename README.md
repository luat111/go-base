# TODO

-   [ ] Auth System
-   [ ] Auth Social ( missing env )
-   [x] Cache ( Lock )
-   [x] Amqp
-   [x] Kafka
-   [x] DTO
-   [x] Error handler
-   [ ] Translate
-   [x] Logging (ELK - take too much resource, DB logger: TODO save error query to DB)
-   [ ] Base Repository (TODO: mongoose) - apply per repo per entity
-   [ ] Encrypt Payload
-   [ ] Audit
-   [ ] Outbox message
-   [ ] Circuit Breaker
-   [ ] Cron
-   [x] GRpc
-   [ ] CI/CD
-   [ ] Monitor
-   [ ] Permission

# SCRIPT GEN RSA KEY

```bash
#private
openssl genpkey -algorithm RSA -out private_key.pem

#public
openssl rsa -in private_key.pem -pubout -out public_key.pem
```
