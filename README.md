# TODO

-   [ ] Auth System
-   [ ] Auth Social ( missing env )
-   [x] Cache ( Lock, Data Layer Service )
-   [x] Amqp
-   [x] Kafka
-   [x] DTO
-   [x] Error handler
-   [ ] Translate
-   [x] Logging (ELK - take too much resource, DB logger: TODO save error query to DB)
-   [x] Base Repository
-   [ ] Encrypt Payload
-   [ ] Audit
-   [ ] Outbox message
-   [ ] Circuit Breaker
-   [x] Cron
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
