# Operation Borderless Backend

## Overview

Operation Borderless is a fintech platform for managing digital wallets with multiple stablecoin balances. This backend, built in Go with a PostgreSQL database, powers the core functionality, serving a React + Tailwind CSS frontend (`github.com/toluhikay/fx-exchange-front`). It supports user registration and login (JWT-based authentication), wallet creation, deposits, swaps, transfers between wallets, transaction history, and viewing balances with their USD equivalent. The backend is deployed on Railway with auto-deployment from GitHub (`github.com/toluhikay/fx-exchange`) and uses a manually configured PostgreSQL database via Beekeeper Studio. It’s designed for security, scalability, and compliance, with audit logging for all operations.

## Features

- **User Management**: Register and log in users, issuing JWT tokens tied to user IDs for secure wallet access.
- **Wallet Creation**: Create wallets linked to user IDs, storing balances in multiple stablecoins (e.g., cNGN, USDx).
- **Deposit**: Add funds to a wallet in a specified stablecoin.
- **Swap**: Convert funds between stablecoins using real-time exchange rates.
- **Transfer**: Send funds to another wallet, tracking both sender and receiver.
- **Transaction History**: View all wallet operations (deposits, swaps, transfers).
- **Balances**: Display all stablecoin balances and their total USD equivalent.
- **Authentication**: Secure endpoints with JWT, extracting user IDs to fetch associated wallets.
- **Audit Logging**: Record all operations in a database for compliance, including client IP and user agent.
- **WebSocket Rates**: Stream real-time exchange rates via `/ws/fx-rates`.

## Tech Stack

- **Go**: For building a high-performance backend.
- **PostgreSQL**: Database for storing users, wallets, transactions, rates, and audit logs.
- **Chi Router**: For routing and middleware (CORS, logging, recovery).
- **JWT**: For user authentication and authorization.
- **WebSocket**: For streaming exchange rates.
- **Railway**: For hosting and auto-deployment.
- **Beekeeper Studio**: For manual database schema setup.

## Prerequisites

- **Go**: Version 1.18 or higher.
- **PostgreSQL**: Version 13 or higher.
- **Beekeeper Studio**: For manual database setup.
- **Railway Account**: For deployment.
- **Node.js/React Frontend**: For API consumption (e.g., `https://fx-exchange-front.vercel.app` or `http://localhost:5173`).

## Installation

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/toluhikay/fx-exchange.git
   cd fx-exchange
   ```

2. **Install Dependencies**:

   ```bash
   go mod tidy
   ```

3. **Configure Environment Variables**:

   - Create a `.env` file in the root directory.
   - Add the following:
     ```
     DATABASE_URL=postgres://<user>:<password>@<host>:<port>/<dbname>?sslmode=disable
     JWT_SECRET=<your-secret-key>
     USE_MOCK_FX=true  # Set to false for real FX rates
     ```
   - Example for local setup:
     ```
     DATABASE_URL=postgres://project-delta-user:project-delta@localhost:5432/project-delta?sslmode=disable
     JWT_SECRET=your-secret-key-here
     USE_MOCK_FX=true
     ```

4. **Set Up Database**:

   - Open Beekeeper Studio and connect to your PostgreSQL database using credentials from `DATABASE_URL`.
   - Copy and execute the SQL commands from `migrations/init.sql` to create tables (`users`, `wallets`, `transactions`, `fx_rates`, `audit_logs`) and enable the `uuid-ossp` extension.
   - Verify schema:
     ```sql
     \dt
     \d users
     \d wallets
     \d transactions
     \d fx_rates
     \d audit_logs
     ```

5. **Run the Backend Locally**:
   ```bash
   go run main.go
   ```
   The server runs at `http://localhost:8080`.

## Deployment

The backend is deployed on Railway with auto-deployment from the `main` branch of `github.com/toluhikay/fx-exchange`. To deploy:

1. Push changes to the `main` branch.
2. Railway automatically builds and deploys using `go build`.
3. Configure environment variables in Railway’s dashboard:

## Usage

1. **Access the API**:

   - The API is available at the deployed URL or `http://localhost:8080`.
   - All `/api/wallets/*` endpoints require a JWT token obtained via `/api/user/login`.

2. **API Endpoints**:

   - **User Registration**: `POST /api/user/register`
     - Payload: `{"email": "test@example.com", "password": "securepassword"}`
     - Creates a user and returns a user ID.
   - **User Login**: `POST /api/user/login`
     - Payload: `{"email": "test@example.com", "password": "securepassword"}`
     - Returns a JWT token containing the user ID.
   - **Get User**: `GET /api/user/`
     - Headers: `Authorization: Bearer {jwt_token}`
     - Returns user details based on the user ID from the JWT.
   - **Create Wallet**: `POST /api/wallets`
     - Headers: `Authorization: Bearer {jwt_token}`
     - Payload: `{"email_or_mobile": "test@example.com"}`
     - Creates a wallet linked to the user ID from the JWT.
   - **Get Wallet**: `GET /api/wallets`
     - Headers: `Authorization: Bearer {jwt_token}`
     - Returns the user’s wallet based on the user ID from the JWT.
   - **Deposit**: `POST /api/wallets/deposit`
     - Headers: `Authorization: Bearer {jwt_token}`
     - Payload: `{"currency": "cNGN", "amount": 1000.1234}`
     - Adds funds to the user’s wallet.
   - **Swap**: `POST /api/wallets/swap`
     - Headers: `Authorization: Bearer {jwt_token}`
     - Payload: `{"from_currency": "cNGN", "to_currency": "USDx", "amount": 500.5678}`
     - Converts funds using `fx_rates`.
   - **Transfer**: `POST /api/wallets/transfer`
     - Headers: `Authorization: Bearer {jwt_token}`
     - Payload: `{"receiver_wallet_id": "{receiverWalletID}", "currency": "cNGN", "amount": 200.9012}`
     - Transfers funds to another wallet, tracking sender and receiver.
   - **Transaction History**: `GET /api/wallets/history`
     - Headers: `Authorization: Bearer {jwt_token}`
     - Returns all transactions for the user’s wallet (deposits, swaps, transfers).
   - **Balances**: `GET /api/wallets/balances`
     - Headers: `Authorization: Bearer {jwt_token}`
     - Returns stablecoin balances and total USD equivalent (e.g., `{"balances": {"cNGN": 1000.1234, "USDx": 0.6000}, "total_usd": 1.2001}`).
   - **WebSocket Rates**: `GET /ws/fx-rates`
     - Streams real-time exchange rates (mock or live based on `USE_MOCK_FX`).

3. **Testing**:
   - Due to time constraints, I tested manually using Postman and Beekeeper Studio.
   - Example API calls:
     ```bash
     # Register user
     curl -X POST -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"securepassword"}' \
     https://operation-borderless-production.up.railway.app/api/user/register
     # Login to get JWT
     curl -X POST -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"securepassword"}' \
     https://operation-borderless-production.up.railway.app/api/user/login
     # Get balances
     curl -H "Authorization: Bearer {jwt_token}" -H "X-Forwarded-For: 192.168.1.100" -H "User-Agent: TestClient/1.0" \
     https://operation-borderless-production.up.railway.app/api/wallets/balances
     ```
     Expected balances response:
     ```json
     {
       "balances": { "cNGN": 1000.1234, "USDx": 0.6 },
       "total_usd": 1.2001
     }
     ```
   - Verify database state in Beekeeper Studio:
     ```sql
     SELECT * FROM users WHERE email = 'test@example.com';
     SELECT * FROM wallets WHERE user_id = '{userID}';
     SELECT * FROM transactions WHERE wallet_id = '{walletID}' OR receiver_wallet_id = '{walletID}';
     SELECT * FROM audit_logs WHERE wallet_id = '{walletID}';
     ```

## Challenges

- **Database Setup**: Manually applying `init.sql` in Beekeeper Studio required careful execution to handle foreign key dependencies and enable `uuid-ossp`. I fixed an initial error by adding the extension explicitly.
- **JWT and User ID**: Extracting user IDs from JWTs to fetch wallets required updating handlers to use `user_claims` from the context, ensuring secure wallet access.
- **Rate Limiting**: I planned to add rate limiting middleware but skipped it due to time constraints. The `routes.go` includes a TODO for future implementation.
- **Testing**: Without automated tests, I relied on manual API calls and SQL queries, which was time-consuming but effective for verifying functionality.
- **CORS**: Configuring CORS for the frontend (`https://fx-exchange-front.vercel.app`, `http://localhost:5173`) took trial and error to avoid browser errors.

## Future Improvements

- Add rate limiting middleware to prevent API abuse.
- Implement automated tests with Go’s `testing` package.
- Use a structured logger (e.g., Zap) for better log management.
- Add Redis for persistent rate limiting and caching.
- Enhance WebSocket rate streaming with error handling.

## Repository

- **GitHub**: `github.com/toluhikay/fx-exchange`
- **Frontend**: `github.com/toluhikay/fx-exchange-front`

## Contact

For issues or contributions, open a pull request or issue on GitHub. Reach out to me via GitHub for support.
