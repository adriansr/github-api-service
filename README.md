# github-api-service
A service to query contributors info from GitHub API


Rate-limits:
    GitHub imposes a rate limit (check).
    a. Take into account
    b. Authenticate to increase limit

User-Agent:
    Header required in requests.
    Include username / repo path so I can be contacted

Missing:
    Cache to avoid frequent queries to GitHub for the same city

Response:
    Add metadata like `count`?

Ref:
https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1


go test ./...