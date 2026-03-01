# Dump Event Schema and Transport Contract

Date: 2026-02-28
Status: v1

## Purpose

Define one stable event format for dump ingestion across HTTP requests, CLI commands, workers, and cron jobs.

## Event format

Each dump event is a single JSON object with this top-level shape.

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `schemaVersion` | integer | yes | Current version is `1`. |
| `id` | string | yes | Unique event ID (UUID/ULID acceptable). |
| `timestamp` | string | yes | RFC3339Nano UTC timestamp. |
| `sourceType` | string | yes | One of `http`, `cli`, `worker`, `cron`. |
| `projectRoot` | string | yes | Absolute project root path when known. |
| `phpSapi` | string | yes | e.g. `fpm-fcgi`, `cli`. |
| `requestId` | string or null | yes | HTTP request correlation ID when available, else `null`. |
| `http` | object | no | Present for HTTP context. |
| `command` | object | no | Present for CLI/worker/cron context. |
| `isDd` | boolean | yes | `true` if event originated from `dd()`. |
| `payloadFormat` | string | yes | Payload encoding. v1 uses `json`. |
| `payload` | object/array/string/number/boolean/null | yes | Captured dump payload in normalized JSON form. |
| `trace` | array | yes | Stack trace frames, may be empty. |
| `host` | object | yes | Host/process metadata. |

### `http` object (optional)

| Field | Type | Required |
| --- | --- | --- |
| `method` | string | yes |
| `scheme` | string | yes |
| `host` | string | yes |
| `path` | string | yes |
| `query` | string | no |
| `statusCode` | integer | no |
| `clientIp` | string | no |
| `userAgent` | string | no |

### `command` object (optional)

| Field | Type | Required |
| --- | --- | --- |
| `name` | string | yes |
| `args` | array of strings | no |
| `cwd` | string | no |

### `trace[]` item

| Field | Type | Required |
| --- | --- | --- |
| `file` | string | no |
| `line` | integer | no |
| `func` | string | no |

### `host` object

| Field | Type | Required |
| --- | --- | --- |
| `hostname` | string | yes |
| `pid` | integer | yes |

## Transport framing

- Transport: Unix domain socket stream.
- Framing: newline-delimited JSON (NDJSON).
- Rule: each line is exactly one JSON object encoded in UTF-8 and terminated by `\n`.
- Sender behavior:
  - must write one complete event per line;
  - should avoid multiline JSON pretty-printing;
  - should use short write timeout to avoid blocking PHP execution.
- Receiver behavior:
  - read stream line-by-line;
  - ignore empty lines;
  - parse each line as JSON object;
  - reject invalid lines without terminating the socket session unless protocol corruption is unrecoverable.

## Versioning and compatibility

- `schemaVersion` is a major integer.
- v1 consumer compatibility rules:
  - accepts only `schemaVersion: 1`;
  - ignores unknown extra fields for forward-compatible additive changes;
  - rejects events missing required fields;
  - rejects unsupported major versions.
- Producer guidance:
  - keep required fields stable within a major version;
  - add only optional fields in backward-compatible updates.

## Examples (NDJSON lines)

HTTP:

```json
{"schemaVersion":1,"id":"01JNFKEC8Q4Y8S97R2M5W12Q9H","timestamp":"2026-02-28T11:20:31.331Z","sourceType":"http","projectRoot":"/home/ronald/code/example-app","phpSapi":"fpm-fcgi","requestId":"f2a1a3d2-2087-4dc4-9fc4-3f8e75ae3202","http":{"method":"GET","scheme":"https","host":"example.test","path":"/users/42","query":"include=roles","statusCode":200,"clientIp":"127.0.0.1","userAgent":"Mozilla/5.0"},"isDd":false,"payloadFormat":"json","payload":{"user":{"id":42,"name":"Ada"}},"trace":[{"file":"/var/www/html/routes/web.php","line":12,"func":"{closure}"}],"host":{"hostname":"ronald-linux","pid":48211}}
```

CLI:

```json
{"schemaVersion":1,"id":"01JNFKEPA3A4CNV3K2E12YVYTG","timestamp":"2026-02-28T11:21:18.011Z","sourceType":"cli","projectRoot":"/home/ronald/code/example-app","phpSapi":"cli","requestId":null,"command":{"name":"artisan","args":["queue:work","--queue=emails"],"cwd":"/home/ronald/code/example-app"},"isDd":false,"payloadFormat":"json","payload":{"job":"SendWelcomeEmail","attempt":1},"trace":[{"file":"/var/www/html/app/Jobs/SendWelcomeEmail.php","line":54,"func":"handle"}],"host":{"hostname":"ronald-linux","pid":49302}}
```

Worker (`dd()`):

```json
{"schemaVersion":1,"id":"01JNFKF5AS6ZD76B8J6BPD0TEW","timestamp":"2026-02-28T11:22:09.892Z","sourceType":"worker","projectRoot":"/home/ronald/code/example-app","phpSapi":"cli","requestId":null,"command":{"name":"artisan","args":["horizon"],"cwd":"/home/ronald/code/example-app"},"isDd":true,"payloadFormat":"json","payload":{"message":"worker halted","context":{"queue":"default"}},"trace":[{"file":"/var/www/html/app/Jobs/ProcessPodcast.php","line":88,"func":"handle"}],"host":{"hostname":"ronald-linux","pid":50077}}
```