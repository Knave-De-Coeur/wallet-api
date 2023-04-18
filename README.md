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

### Instructions
This is taking the assumption that a user can have many wallets but only one wallet can belong to a user at any given point.

- run `docker-compose up -d` to set up redis and mysql
- wallet balances (funds) are saved as whole numbers then divised by 100 to get the cents.
- opted to not stop flow when errors in cache crop up
- opted to not add transaction, bet, win tables to avoid complexity
