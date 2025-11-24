# TODO

-   [x] Auth System
-   [x] Auth Social ( missing env )
-   [x] Cache ( Lock, Data Layer Service )
-   [x] Amqp
-   [x] Kafka
-   [x] DTO
-   [x] Error handler
-   [x] Translate
-   [x] Logging (ELK - take too much resource, DB logger: TODO save error query to DB)
-   [x] Base Repository
-   [x] Encrypt Payload
-   [ ] Audit
-   [x] Outbox message
-   [x] Circuit Breaker
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

## Crypto payload

`pkg/crypto` mirrors the Node payload-encryption flow (RSA + AES-256-CBC) for REST handlers. Example:

```go
import (
	"go-base/pkg/crypto"
	"os"
)

svc, err := crypto.NewService(crypto.Config{
	PublicKeyPEM:  os.Getenv("RSA_PUBLIC_KEY"),
	PrivateKeyPEM: os.Getenv("RSA_PRIVATE_KEY"),
})
if err != nil {
	panic(err)
}

router.Use(svc.PayloadMiddleware(crypto.MiddlewareConfig{
	RequireHeader: true,
	// DisableSanitize: true, // optional
}))
```

Clients encrypt the AES password with the server public key, send it through the `X-Request-Encrypt-Key` header, and AES-encrypt (then base64) the JSON payload. Responses — including errors — are automatically re-encrypted before being returned.
