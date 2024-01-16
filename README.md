# siasn-jf-backend

SIASN JF Management backend.

## Vendoring

This repository use vendoring and all packages are pushed to the repository. Why?

The reason we do this is to simplify the CI/CD process. Although it makes the repository multitude larger, so that
during build external libraries are not pulled. Some external libraries can only be pulled with GOPRIVATE. To comply
with other code standards preexisting in BKN, we use Dockerfile that follows the same flow as probably other Dockerfile
that has already exist in BKN. This requires libraries to be available at build time.

Because of that, we also add `go.sum` to the repository as opposed to ignoring it.

**Because of this, make sure go.sum and the libraries this repo depends on are in sync by calling `go mod tidy` and
`go mod verify` before pushing.**

## Configurations

As a 12-factor app, siasn-jf-backend utilizes purely environment variables as configurations. For array values,
siasn-jf-backend *uses space as separator*.

**All configurations have prefix** `SIASN_JF_`.

| Configuration Name                                | Description                                                                                                          | Default Value                                        |
|---------------------------------------------------|----------------------------------------------------------------------------------------------------------------------|------------------------------------------------------|
| LISTEN_ADDRESS                                    | The listen address combinations for siasn-jf                                                                         | 127.0.0.1:8080                                       |
| ENABLE_TLS                                        | Set true to enable TLS listener (and its configurations)                                                             | 0                                                    |
| TLS_LISTEN_ADDRESS                                | The listen address combinations (TLS)                                                                                | 127.0.0.1:8443                                       |
| TLS_CERT_FILE                                     | TLS public certificate file                                                                                          | certs/localhost.crt                                  |
| TLS_KEY_FILE                                      | TLS private key file                                                                                                 | certs/localhost.key                                  |
| PROMETHEUS_LISTEN_ADDRESS                         | Prometheus metric server listen address                                                                              | 0.0.0.0:9080                                         |
| CORS_ALLOWED_HEADERS                              | CORS allowed headers, array of string                                                                                | <see below>                                          |
| CORS_ALLOWED_METHODS                              | CORS allowed methods, array of string                                                                                | <see below>                                          |
| CORS_ALLOWED_ORIGINS                              | CORS allowed origins, array of string                                                                                | <see below>                                          |
| OIDC_PROVIDER_URL                                 | Used to retrieve OIDC discovery settings, available under <OidcProviderUrl>/.well-known.                             | https://iam-siasn.bkn.go.id/auth/realms/public-siasn |
| OIDC_CLIENT_ID                                    | Client ID registered with OpenID Connect IdP                                                                         | manajemen-jf                                         |
| OIDC_CLIENT_SECRET                                | Client secret registered with OpenID Connect IdP                                                                     |                                                      |
| OIDC_CLIENT_END_SESSION_ENDPOINT                  | OIDC end session endpoint (deprecated)                                                                               | <OIDC_PROVIDER_URL>/protocol/openid-connect/logout   |
| OIDC_CLIENT_REDIRECT_URL                          | Redirect URL registered with the IdP                                                                                 | http://training-manajemen-jf.bkn.go.id/api/oauth     |
| OIDC_CLIENT_SUCCESS_REDIRECT_URL                  | The URL to which the user is redirected after successful login                                                       | http://training-manajemen-jf.bkn.go.id               |
| POSTGRES_URL                                      | Full postgres:// URL to connect to PostgreSQL                                                                        | postgres://postgres:@localhost:5432/siasn_jf         |
| PROFILE_POSTGRES_URL                              | Full postgres:// URL to connect to PostgreSQL that store read-only ASN profile data                                  | postgres://postgres:@localhost:5432/db_profilasn     |
| REFERENCE_POSTGRES_URL                            | Full postgres:// URL to connect to PostgreSQL that store read-only BKN reference data                                | postgres://postgres:@localhost:5432/db_referensi     |
| EMC_ECS_ENDPOINT                                  | Full EMC ECS endpoint URL                                                                                            |                                                      |
| EMC_ECS_ACCESS_KEY                                | EMC ECS access key                                                                                                   |                                                      |
| EMC_ECS_SECRET_KEY                                | EMC ECS secret key                                                                                                   |                                                      |
| EMC_ECS_REGION                                    | S3 region that is configured in the EMC ECS                                                                          |                                                      |
| TEMP_BUCKET                                       | Bucket name to store temporary files                                                                                 |                                                      |
| TEMP_ACTIVITY_DIR                                 | Directory relative to TEMP_BUCKET without leading/trailing slash to store temporary activity files                   | activity                                             |
| ACTIVITY_BUCKET                                   | Bucket name to store activity support doc files                                                                      |                                                      |
| ACTIVITY_DIR                                      | Directory relative to ACTIVITY_BUCKET without leading/trailing slash to store activity files                         | activity                                             |
| TEMP_REQUIREMENT_DIR                              | Directory relative to TEMP_BUCKET without leading/trailing slash to store temporary req. files                       | requirement                                          |
| REQUIREMENT_BUCKET                                | Bucket name to store requirement support doc files                                                                   |                                                      |
| REQUIREMENT_DIR                                   | Directory relative to REQUIREMENT_BUCKET without leading/trailing slash to store requirement files                   | requirement                                          |
| REQUIREMENT_TEMPLATE_DIR                          | Directory for storing template documents.                                                                            | requirement-template                                 |
| REQUIREMENT_TEMPLATE_COVER_LETTER_FILENAME        | The filename of cover letter template including its extension.                                                       | surat-pengantar-template.docx                        |
| TEMP_DISMISSAL_DIR                                | Directory relative to TEMP_BUCKET without leading/trailing slash to store temporary req. files                       | dismissal                                            |
| DISMISSAL_BUCKET                                  | Bucket name to store dismissal support doc files                                                                     |                                                      |
| DISMISSAL_DIR                                     | Directory relative to DISMISSAL_BUCKET without leading/trailing slash to store dismissal files                       | dismissal                                            |
| DISMISSAL_TEMPLATE_DIR                            | Directory for storing template documents.                                                                            | dismissal-template                                   |
| DISMISSAL_TEMPLATE_ACCEPTANCE_LETTER_FILENAME     | The filename of acceptance letter template including its extension.                                                  | surat-pemberhentian-template.docx                    |
| PROMOTION_BUCKET                                  | Bucket name to store promotion related doc files                                                                     |                                                      |
| TEMP_PROMOTION_DIR                                | Directory relative to TEMP_BUCKET without leading/trailing slash to store temporary req. files                       | promotion                                            |
| PROMOTION_DIR                                     | Directory relative to PROMOTION_BUCKET without leading/trailing slash to store promotion related files               | promotion                                            |
| PROMOTION_TEMPLATE_DIR                            | Directory for storing template documents.                                                                            | promotion-template                                   |
| PROMOTION_TEMPLATE_PAK_LETTER_FILENAME            | The filename of PAK letter template including its extension.                                                         | surat-pak-template.docx                              |
| PROMOTION_TEMPLATE_RECOMMENDATION_LETTER_FILENAME | The filename of recommendation letter template including its extension.                                              | surat-rekomendasi-template.docx                      |
| PROMOTION_CPNS_BUCKET                             | Bucket name to store promotion for CPNS related doc files                                                            |                                                      |
| TEMP_PROMOTION_CPNS_DIR                           | Directory relative to TEMP_BUCKET without leading/trailing slash to store temporary req. files                       | promotion-cpns                                       |
| PROMOTION_CPNS_DIR                                | Directory relative to PROMOTION_CPNS_BUCKET without leading/trailing slash to store promotion for CPNS related files | promotion-cpns                                       |
| SIASN_DOCX_CMD                                    | siasn-docx command name                                                                                              | siasn-docx                                           |
| SOFFICE_CMD                                       | soffice command name                                                                                                 | soffice                                              |
| LOGGING_TO_STD                                    | Whether to log to stdout or not                                                                                      | 1                                                    |
| LOGGING_STD_COLOR                                 | Whether to log to stdout with color codes                                                                            | 1                                                    |
| LOGGING_TO_FILE                                   | Whether to log to file                                                                                               | 1                                                    |
| LOGGING_FILE_PATH                                 | File path to output the logs into                                                                                    | logs/siasn-jf-backend.log                            |
| ASSESSMENT_TEAM_BUCKET                            | Bucket name to store assessment team related doc files                                                               |                                                      |
| TEMP_ASSESSMENT_TEAM_DIR                          | Directory relative to TEMP_BUCKET without leading/trailing slash to store temporary assessment team files            | assessment-team                                      |
| ASSESSMENT_TEAM_DIR                               | Directory relative to ASSESSMENT_TEAM_BUCKET without leading/trailing slash to store assessment team related files   | assessment-team                                      |

## CORS Default Settings

By default, all methods are allowed for CORS_ALLOWED_METHODS.

These are the headers allowed by CORS_ALLOWED_HEADERS (apart from standard headers that are always allowed):

```text
Accept
Accept-Language
Content-Type
Content-Language
Content-Disposition
Origin
X-Requested-With
X-Forwarded-For
```

CORS_ALLOWED_ORIGINS has emtpy default values.

Credentials are allowed to be sent.

## About `GET` and `DELETE` Queries

It is mandatory that all GET and DELETE queries do *not* have any request body content. This follows the fact that HTTP
standard does not define what happens when a client includes request body in GET/DELETE request and send it to some HTTP
services. Because of that, people do not include request body in GET/DELETE to avoid instability and inconsistency as
different services from different developers might understand it differently. Because of that, we choose to forbid
request body to be sent with GET request and request body will not be read for such request.

## About `PUT` Requests

Endpoints that need PUT method usually replace the data that the endpoint handles with newer data that you supplied.
This is usually done to update some data. Therefore, to make use of those endpoints correctly, make sure to supply all
data fields including those that are unchanged as the request.

Because PUT endpoints replace old data with new data, if empty entry is supplied in the request, old entry pointed by
the ID supplied in the request will be replaced with empty entry instead. This happens usually if you leave out some of
the fields in the data model with empty strings.

## Building this Service

Building this service follows the standard backend service build steps for this project.

### Getting Dependencies

Because this repository uses vendoring, all dependencies should be available under the `vendor` directory and
`go.mod` and `go.sum` files are all available and synchronized with dependencies. Because some dependencies might only
be available in private repositories, you supposedly need `GOPRIVATE` environment and authentication to pull
dependencies from those private repositories. But because these dependencies have been included in the `vendor`
directory, you don't need to do anything. `go mod` commands do require `GOPRIVATE` environment and credentials to be set
though, so avoid running those command before building.

In other words, you supposedly don't need to do anything to fetch the dependencies required for this service.

### Building

To build this repository, run:

```bash
make build
```

If `make` is not available, you can run `go build` manually. Just run

```bash
go build
```

`make build` or `go build` will create an executable file called `siasn-jf-backend`. You can then run the service as
usual like so: `./siasn-jf-backend`.

### Running the Service

You can run the service directly by calling the service name in the terminal but that likely won't work. That's because
this service requires some configurations which do not have default values to begin with. Some configurations are rather
confidential, such as the database connection string which could contain specific URL and credentials.

Because this service is a 12-factor app, all configurations can be supplied through the use of environment variables.
See above table to see which configurations are needed.

You can run this service by supplying environment variables inline like so:

```bash
SIASN_JF_POSTGRES_URL=postgres://username:password@hostname:port/dbname SIASN_JF_LISTEN_URL=0.0.0.0:8080 ./siasn-jf-backend
```

Or you can `export` the environment variables first:

```bash
export SIASN_JF_POSTGRES_URL=postgres://username:password@hostname:port/dbname
./siasn-jf-backend
```

Exported variables will persist only for the terminal session where it is run until the terminal session is
closed/terminated.

### `siasn-docx` and `soffice` Binary

The `siasn-docx` binary is used to render docx template. It is custom made for the SIASN project, and so cannot be
installed through package managers. It can be downloaded through the siasn-docx repository release page, but to reduce
deployment overhead, it is better to just include the script in this repository instead. It can be moved into the Docker
container during build.

`soffice` can be installed with package managers (included with libreoffice), just make sure it is installed during
Docker build, or installed in your local system if you intend to run the service locally.

**However**, if you are deploying this service with Docker, you can use the `siasn-runner` base image instead, which
already contains these up-to-date binaries.

## Legal and Acknowledgements

This repository was built by:

* Sergio Ryan \[[sergioryan@potatobeans.id](mailto:sergioryan@potatobeans.id)]
* Eka Novendra \[[novendraw@potatobeans.id](mailto:novendraw@potatobeans.id)]
* Stefanus Ardi Mulia \[[stefanusardi@potatobeans.id](mailto:stefanusardi@potatobeans.id)]
* Audie Masola Putra \[[audiemasola@potatobeans.id](mailto:audiemasola@potatobeans.id)]

Copyright &copy; 2021 Indonesian National Civil Service Agency (Badan Kepegawaian Negara, BKN).  
All rights reserved.#   l i b s - b a c k e n d - v 2  
 