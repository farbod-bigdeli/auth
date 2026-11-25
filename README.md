# Auth: Identity-Verified Authentication Service

An authentication service in Go for a platform that needs to know who its users really are. Before an account is created, the service checks the user's phone number and Iranian national ID with [Finnotech](https://finnotech.ir), an open-banking API provider. Users can also prove ownership of a bank account and upload supporting documents for review.

It was built in 2021 with Gin, GORM and PostgreSQL, and is deployed as a Docker image.

## What it does

**Sign-up with identity checks**

1. The user enters a phone number, and Finnotech sends them a one-time code by SMS.
2. The service verifies the code with Finnotech and gets an OAuth token for that user.
3. It checks the national ID, name, birth date and gender against Finnotech's records.
4. If everything matches, it creates the account and texts the user a generated password (SMS via Kavenegar).

**Two-step login:** the user signs in with username and password (stored as bcrypt hashes), then confirms with an SMS code. Only then does the service issue a JWT, which is valid for 60 hours and is revoked on logout.

**Password reset:** the user verifies by SMS code and receives a new password by SMS.

**Proofs:** logged-in users can submit two kinds of proof, and each one moves through review states:
- **Commitment documents:** uploaded images, stored on disk
- **Bank account:** card number and IBAN, with ownership checked through Finnotech

## API

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/auth/sendSMS` | none | Send a sign-up verification code |
| POST | `/auth/signup` | none | Verify the code and identity, then create the account |
| POST | `/auth/login` | none | Check username and password, then send a login code |
| POST | `/auth/login_verify` | none | Verify the login code and return a JWT |
| POST | `/auth/forgot_password` | none | Send a password-reset code |
| POST | `/auth/verify_forgot_password` | none | Verify the code and send a new password |
| GET | `/auth/validate` | JWT | Check that a token is valid |
| GET | `/auth/user/get` | JWT | Current user's profile |
| POST | `/auth/change_password` | JWT | Change password |
| POST | `/auth/proof/store` | JWT | Submit a document or bank-account proof |
| GET | `/auth/user/bank` | JWT | The user's verified bank account |
| GET | `/auth/logout` | JWT | Revoke the token |

## Project layout

| Path | Responsibility |
|---|---|
| `main.go` | Routes and JWT middleware |
| `auth/` | Sign-up, login, password reset and SMS code checks |
| `finnotech/` | Finnotech client: OAuth tokens, SMS verification, national ID and bank checks |
| `bankAccount/` | Bank account input checks and ownership verification |
| `proof/` | Proof submission and lookup |
| `jwt/` | Token creation, verification and revocation |
| `notification/` | SMS sending through Kavenegar |
| `model/`, `database/` | GORM models and the PostgreSQL connection |
| `config/` | Loads `config.json` |

## Configuration

Database and server settings live in `config.json` (or `config_server.json` for the Docker image):

```json
{
    "localPort": ":8000",
    "envPath": ".env",
    "dbHost": "localhost",
    "dbPort": "5432",
    "dbName": "pg2",
    "dbUserName": "CHANGE_ME",
    "dbPassword": "CHANGE_ME",
    "baseFilePath": "storage/"
}
```

Secrets are read from a `.env` file, which is not committed:

```
JWT_SECRET=
SMS_API_KEY=
FINNOTECH_AUTHORIZE_STRING=
NID=
```

## Run

With Go 1.17+ and a PostgreSQL database:

```bash
go run .
```

With Docker:

```bash
docker build -t auth .
docker compose up -d
```

`docker-compose.yml` starts the service, a reverse proxy on port 8000 and PostgreSQL 13. The Finnotech endpoints point at Finnotech's sandbox (`apibeta.finnotech.ir`), and the client ID and redirect URL in `auth/auth.go` need to be set for your own Finnotech app.
