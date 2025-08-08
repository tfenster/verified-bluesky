#! /bin/bash

baseurl="http://localhost:3000"
# baseurl="https://verifiedbsky.net"

#accounts=$(curl -s "$baseurl/admin/data/${SPIN_VARIABLE_BSKY_PASSWORD}" | jq -r '.[].Value')
accounts=$(echo '[{"Key":"dynamicsminds-Tobias Fenster","Value":"tobiasfenster.io"},{"Key":"rd-2efc9bb2-6a8c-e711-811e-3863bb36edf8","Value":"tobiasfenster.io"},{"Key":"colorcloud-Tobias Fenster","Value":"tobiasfenster.io"},{"Key":"mvp-2efc9bb2-6a8c-e711-811e-3863bb36edf8","Value":"tobiasfenster.io"}]' | jq -r '.[].Value')
accounts=$(echo "$accounts" | tr ' ' '\n' | sort -u | tr '\n' ' ')
# echo $accounts

for handle in $accounts; do
    echo "Validating account: $handle"
            
    # Get current validation status for all modules (with authentication)
    validation_response=$(curl -s "$baseurl/weekly-validation/$handle/${SPIN_VARIABLE_BSKY_PASSWORD}")
    echo "Validation response: $validation_response"
    
    # Parse the response and process each module
    echo "$validation_response" | jq -r '.moduleResults | to_entries[] | @base64' | while IFS= read -r module_data; do
        module_info=$(echo "$module_data" | base64 --decode)
        module_key=$(echo "$module_info" | jq -r '.key')
        is_valid=$(echo "$module_info" | jq -r '.value.isValid')
        current_failure_count=$(echo "$module_info" | jq -r '.value.failureCount')
        
        echo "Account $handle, Module $module_key - Valid: $is_valid, Current failures: $current_failure_count"
        
        if [ "$is_valid" = "true" ]; then
        # Module validation is valid, reset failure count to 0
        echo "Account $handle module $module_key is valid, resetting failure count to 0"
        curl -s -X POST "$baseurl/weekly-validation/${SPIN_VARIABLE_BSKY_PASSWORD}" \
            -H "Content-Type: application/json" \
            -d "{\"bskyHandle\": \"$handle\", \"moduleKey\": \"$module_key\", \"failureCount\": 0}"
        else
        # Module validation is invalid, increment failure count
        new_failure_count=$((current_failure_count + 1))
        echo "Account $handle module $module_key is invalid, incrementing failure count to $new_failure_count"

        response=$(curl -s -X POST "$baseurl/weekly-validation/${SPIN_VARIABLE_BSKY_PASSWORD}" \
            -H "Content-Type: application/json" \
            -d "{\"bskyHandle\": \"$handle\", \"moduleKey\": \"$module_key\", \"failureCount\": $new_failure_count}")
        
        echo "Update response: $response"
        
        # Check if user was removed from this module
        action=$(echo "$response" | jq -r '.action')
        if [ "$action" = "partial_removal" ]; then
            echo "⚠️ Account $handle was removed from module $module_key after the configured (MaxFailureCount) number of consecutive validation failures"
        else
            echo "Account $handle module $module_key now has $new_failure_count consecutive failures"
        fi
        fi
        
        # Add a small delay between module checks to avoid rate limiting
        sleep 1
    done
    
    # Add a delay between accounts to avoid rate limiting
    sleep 2
done