# Start Application

1. `docker compose up --build`

---

## Idempotency Implementation

To ensure that duplicate requests with the same `transaction_id` do not create multiple records, idempotency is implemented using:

1. **Code Level**
    - Before processing, the service checks for an existing payment by `transaction_id`.
    - If found, it returns the same response, ensuring consistency.
2. **Database Level**
    - A unique constraint is added to the `transaction_id` column to prevent duplicate records at the database level.

---

## Request Flow

1. **Enter API Endpoint**
2. **Attempt to Lock by Idempotency Key (`transaction_id`)**
    - **Lock Failed:**
        - Another request with the same `transaction_id` is currently processing.
        - Return `201 Created` with the existing transaction record (if found).
3. **Check Existing Payment**
    - If a record with the same `transaction_id` exists, return it immediately.
4. **Start Processing**
    - Simulate processing with a `1s` delay and create the record.
    - Trigger goroutine for simulate processing
        - update payment status to `completed` or `failed`
        - update wallet balance if completed
5. **Return Response**
    - Return `201 Created` with the newly created payment record.

---

## Testing Instructions

### API Testing
1. Open the **Postman Collection** I created for this project: [Postman Collection](https://www.postman.com/aviation-geoscientist-80328098/workspace/emb).  
   > This collection includes all the API endpoints with example requests and can be used directly for testing.

2. Switch to the **DEV** environment.
3. Set the following environment variable:

| Variable | Value |
| --- | --- |
| `base-url` | `http://localhost:8080/api/v1` |

### How to connect PostgreSQL DB
You can directly connect to the postgreSQL database to view the data through the following config:
- `host`: `localhost`
- `port`: `9999`
- `username`: `postgres`
- `password`: `postgres`


---

## Unit Test
We implemented unit tests for the Payment Service using [Testcontainers-Go](https://golang.testcontainers.org/)  and service-level testing. This approach allows us to spin up a **PostgreSQL** instance during tests, providing:

- Realistic database testing environment
- Isolation per test session
- Automatic cleanup after tests



## Concurrency Testing

To test concurrency and ensure the idempotency logic works as expected, we use the [hey](https://github.com/rakyll/hey) library to do concurrency testing

#### Running the Test

Run the following command under the project root to simulate **100 requests** with a concurrency level of **10**:

```bash
hey -n 100 -c 10 -m POST \
  -H "Content-Type: application/json" \
  -d '{"transaction_id":"unique123","amount":100,"user_id":"1"}' \
  http://localhost:8080/api/v1/pay

```

- **`n 100`** → Total number of requests to send.
- **`c 10`** → Number of concurrent workers.
- **`m POST`** → HTTP method used for the request.
- **`H`** → Add request headers.
- **`d`** → Request body data.

<details>
<summary>My Concurrent Testing Result</summary>

Based on the test results:

- **100 requests** were sent concurrently with **10 workers**.
- **Only 1 request** successfully acquired the lock and proceeded with full processing (indicated by the 1-second latency due to the simulated `time.Sleep(1 * time.Second)`).
- The remaining **99 requests** returned immediately after detecting that the same `transaction_id` was already being processed, which aligns with the **idempotency** design.
- All requests returned **HTTP 201 Created**, confirming that the system consistently returns the same response for duplicate requests with the same `transaction_id`.

    ```bash
        $ hey -n 100 -c 10 -m POST -H "Content-Type: application/json" -d '{"transaction_id":"unique6","amount":100,"user_id":"1"}' http://localhost:8080/api/v1/pay

        # Send 100 requests

        Summary:
        Total:        1.9285 secs
        Slowest:      1.0136 secs
        Fastest:      0.1010 secs
        Average:      0.1129 secs
        Requests/sec: 51.8549

        Total data:   10203 bytes
        Size/request: 102 bytes

        Response time histogram:
        0.101 [1]     |
        0.192 [98]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
        0.283 [0]     |
        0.375 [0]     |
        0.466 [0]     |
        0.557 [0]     |
        0.649 [0]     |
        0.740 [0]     |
        0.831 [0]     |
        0.922 [0]     |
        1.014 [1]     | # only 1 request process 1s (sleep 1s)

        Latency distribution:
        10% in 0.1014 secs
        25% in 0.1016 secs
        50% in 0.1020 secs
        75% in 0.1027 secs
        90% in 0.1141 secs
        95% in 0.1197 secs
        99% in 1.0136 secs

        Details (average, fastest, slowest):
        DNS+dialup:   0.0009 secs, 0.1010 secs, 1.0136 secs
        DNS-lookup:   0.0008 secs, 0.0000 secs, 0.0083 secs
        req write:    0.0000 secs, 0.0000 secs, 0.0003 secs
        resp wait:    0.1118 secs, 0.1009 secs, 1.0047 secs
        resp read:    0.0001 secs, 0.0000 secs, 0.0005 secs

        Status code distribution:
        [201] 100 responses # all responses return 201
    ```
</details>    

<details>
<summary>Application Logging</summary>

```bash
# Examples
emb-payment-backend  | [GIN] 2025/09/13 - 09:10:02 | 201 |  100.881567ms |      172.19.0.1 | POST     "/api/v1/pay"
emb-payment-backend  | [Info] message=Payment processing, failed to acquired the lock...                                 
emb-payment-backend  | [GIN] 2025/09/13 - 09:10:02 | 201 |  100.781749ms |      172.19.0.1 | POST     "/api/v1/pay"
emb-payment-backend  | [Info] message=Payment processing, failed to acquired the lock...                                 
emb-payment-backend  | [GIN] 2025/09/13 - 09:10:02 | 201 |  100.782061ms |      172.19.0.1 | POST     "/api/v1/pay"
emb-payment-backend  | [Info] message=Payment processing, failed to acquired the lock...                                 
emb-payment-backend  | [GIN] 2025/09/13 - 09:10:02 | 201 |  100.804661ms |      172.19.0.1 | POST     "/api/v1/pay"
```
</details>