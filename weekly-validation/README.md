# Weekly Validation Module

This module provides endpoints to support automatic weekly validation of all verified accounts in the key/value store.

## Overview

The weekly validation system ensures that verified accounts remain valid over time by:

1. **Weekly GitHub Workflow**: Runs every Sunday at 2:00 AM UTC
2. **Account Retrieval**: Gets all verified accounts from the `/admin/data` endpoint
3. **Re-validation**: Checks each account using their original verification method
4. **Failure Tracking**: Maintains a failure count for each account
5. **Automatic Cleanup**: Removes accounts after 4 consecutive validation failures

## Endpoints

### GET `/weekly-validation/{bskyHandle}/{password}`

Checks the validation status of a specific Bluesky handle. Requires authentication via password in URL path.

**Response:**
```json
{
  "bskyHandle": "example.bsky.social",
  "isValid": true,
  "failureCount": 0,
  "action": "none"
}
```

### POST `/weekly-validation/{password}`

Updates the failure count for a Bluesky handle. Requires authentication via password in URL path.

**Request:**
```json
{
  "bskyHandle": "example.bsky.social", 
  "failureCount": 1
}
```

**Response:**
```json
{
  "bskyHandle": "example.bsky.social",
  "failureCount": 1,
  "action": "none"
}
```

If `failureCount` reaches 4, the response will include `"action": "removed"` and the account will be:
- Removed from the key/value store
- Removed from all relevant Bluesky lists and starter packs  
- Have their verification label removed

## Authentication

Both endpoints require authentication using the same method as the admin endpoints:
- The Bluesky password must be provided as the last segment of the URL path
- This authenticates the request against the configured Bluesky account
- Unauthorized requests will receive a 401 status code

## Failure Count Logic

- **Valid account**: Failure count is reset to 0
- **Invalid account**: Failure count is incremented by 1
- **4 failures**: Account is completely removed from the system

## GitHub Workflow

The workflow (`weekly-validation.yml`) runs automatically every Sunday and:

1. Retrieves all accounts using the admin data endpoint
2. Validates each account by calling their specific validation endpoint
3. Updates failure counts based on validation results
4. Logs all actions for monitoring

## Manual Trigger

The workflow can also be triggered manually from the GitHub Actions interface for testing or emergency validation runs.

## Security

The workflow requires the following GitHub secrets to be configured:
- `BSKY_PASSWORD`: The Bluesky password for authenticating with admin endpoints

Note: The Bluesky handle is already configured as an environment variable in the Spin application configuration.

## Monitoring

All validation actions are logged in the GitHub workflow output, including:
- Accounts validated
- Failure count updates  
- Account removals
- Any errors during the process
