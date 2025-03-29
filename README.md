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
OPENAI_API_KEY= < Your own Open AI API key >
SECRET= < Can be any secret you want >
ENV=dev
```
4. Open a terminal and run the following commands. Make sure you're in the root of the repository:

```
Make deps
Make start
```
