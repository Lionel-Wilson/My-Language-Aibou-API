# My-Language-Aibou-API
The API for the "My Language Aibou" application. The AI Powered language learning partner

## How to run

### Prerequisites

- Docker Desktop installed.
- Git installed.
- Golang installed.

### How to run the whole application

1. Clone the repository
2. Create a .env file at the root of the repository.
3. Fill the .env file with the following variables:
```
PORT=8080
OPENAI_API_KEY=<your own>
SECRET=<your own>
ENV=dev
LOG_LEVEL=debug
DATABASE_URL=<your own>
JWT_SECRET=<your own>
STRIPE_SECRET_KEY=<your own>
STRIPE_PAID_PRICE_ID= <your own>
STRIPE_WEBHOOK_SECRET=<your own>
CHECKOUT_SUCCESS_URL=<your own>
CHECKOUT_CANCEL_URL=<your own>
```
4. Open a terminal and run the following commands. Make sure you're in the root of the repository:

```
Make deps
Make build
```
