# wallet-api
Player wallets in a casino, dummy api.

## Description
Management of the wallets of the players of an online casino. API for getting and updating their account balances.
A wallet balance cannot go below 0. Amounts sent in the credit and debit operations cannot be negative.
- balance : retrieves the balance of a given wallet id
```
GET /api/v1/wallets/{wallet_id}/balance
```
- credit : credits money on a given wallet id
```
POST / api/v1/wallets/{wallet_id}/credit
```
- debit : debits money from a given wallet id
```
POST / api/v1/wallets/{wallet_id}/debit
```

## Instructions
### Setup: 
- cd `$GOPATH/src/wallet-api` (or wherever the repo was cloned)
- run `docker-compose up -d` to set up redis and mysql
- run `go mod tidy`
- run `mkdir bin` (if it's not already present)
- Run To get the binary file `go build -o $GOPATH/src/github.com/knave-de-coeur/wallet-api/bin/wallet_api $GOPATH/src/github.com/knave-de-coeur/wallet-api/cmd/api/main.go`
- Run `cp -R ./internal/migrations/ ./bin/migrations`
- Run `./bin/wallet-api`
- Api should be up and running with dummy data inserted.
- To run tests, with logger:
```
cd ./internal/services
go test -cover
```
- Otherwise from ./wallet-api run `go test ./internal/services -cover` will give simple output


### Assumptions
- Auth endpoint is `/login` taking in the following request body:
```json
{
    "username": "alexm1496",
    "Password": "pass123"
}
```
- Project is taking the assumption that a user can have many wallets but only one wallet can belong to a user at any given point.
- Wallet balances (funds) are saved as whole numbers. Response are that divied by 100 to get the cents (emulating euro).
- Wallets and users have been pre-populated
- opted to not stop flow when errors in cache crop up
- opted to not add transaction, bet, win tables to avoid complexity
- opted to not add permissions in jwt to avoid complexity
- commented out routes for user CRUD to focus on wallet structure
- opted not to store tokens in redis to avoid complexity
- Unit tests were only added to wallet buisness logic covering requirements mentioned for the wallet endpoints
- Gorm was used so that changing databases won't affect buisness logic.
