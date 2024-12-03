# Evaluation Task: Distributed Notification and Workflow System

## Scenario

Implement a service for a logistics company that needs to handle package
delivery notifications.
The service will receive an information about new deliveries and notify
corresponding users so they can confirm the delivery.

**Service's workflow**

- receive package delivery events from an external source via HTTP API call;
- publish a new delivery event to AWS SQS for delayed processing;
- implement an SQS consumer that uses [Temporal](https://temporal.io/) to
  orchestrate a workflow that includes:
    - initiate a new workflow;
    - call a [webhook](https://webhook.site/) with the package details to
      simulate notifying an external service;
    - waits for user confirmation (done via HTTP API call). Inside the workflow,
      you’ll "pause" at the confirmation step and listen for an external
      signal (API call);
    - retries failed steps up to a configurable limit (if needed);
    - the package data is saved to the database only after the webhook confirms
      success.
- expose a RESTful API for:
    - creating a package delivery event;
    - querying delivery statuses;
    - include HTTP API documentation (Swagger);
- write unit tests for temporal workflow (see docs).

**Technical Stack Requirements**

- HTTP server: [Gin](https://gin-gonic.com/)
  or [Chi](https://github.com/go-chi/chi) HTTP framework.
- Database: PostgreSQL (with [gorm](https://gorm.io/) or native database/sql).
- AWS Services: integrate with SQS
  using [localstack](https://docs.localstack.cloud/getting-started/installation/).
- Workflow
  Orchestrator: [Temporal](https://learn.temporal.io/getting_started/?_gl=1*cg1k7t*_gcl_au*MTUwMzQ1OTg2MC4xNzMyNTQ5NTEx*_ga*MjExNDcyOTk4MC4xNzMyNTQ5NTEx*_ga_R90Q9SJD3D*MTczMjU0OTUxMC4xLjEuMTczMjU0OTc3Ni4wLjAuMA..#run-your-first-program) (
  with their Go SDK).
- Documentation: Swagger/OpenAPI specification

**Required API endpoints**

- `POST /api/v1/packages` - creates a new delivery event

```json
{
  "package_id": "123",
  "customer_email": "customer@example.com",
  "delivery_address": "123 Main Street",
  "status": "pending"
}
```

- `GET /api/v1/packages/{id}` - returns package info. The response body is the
  same as for creating
- `POST /api/v1/packages/{id}/confirm` - confirms a package delivery by client (
  resumes the workflow)
- `GET /api/docs` - serves API documentation (Swagger/OpenAPI)

**Non-Functional Requirements**

- Write clean and idiomatic Go code.
- Include Dockerized setup for local development with LocalStack and Temporal.

**Evaluation Criteria**

- correctness of implementation;
- quality of code (structure, readability, idioms);
- test coverage;
- usage of Temporal for workflow orchestration;
- integration with AWS SNS/SQS (via LocalStack);
- proper API documentation.
